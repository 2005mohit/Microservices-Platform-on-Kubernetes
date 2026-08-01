package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EnsurePullSecret creates (or refreshes) the GHCR dockerconfigjson secret so
// Kaniko can push images and tenant deployments can pull them. The credential
// is derived from GITHUB_TOKEN (a PAT with write:packages + read:packages).
func EnsurePullSecret(ns string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN not set")
	}
	registry := os.Getenv("REGISTRY")
	if registry == "" {
		registry = "ghcr.io/2005mohit"
	}
	parts := strings.Split(registry, "/")
	username := "oauth2"
	if len(parts) > 1 && parts[1] != "" {
		username = parts[1]
	}
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + token))
	cfg := fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":"%s"}}}`, auth)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "ghcr-push", Namespace: ns},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       map[string][]byte{corev1.DockerConfigJsonKey: []byte(cfg)},
	}
	_, err := Client.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil && !isAlreadyExists(err) {
		return err
	}
	return nil
}

// DeployUserApp deploys a user-built image into an isolated tenant namespace
// with a Service and a TLS Ingress (cert-manager).
func DeployUserApp(name, hostname, image string, port int32) error {
	ns := "tenant-" + name
	ctx := context.Background()

	if err := EnsureNamespace(ns); err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("namespace: %w", err)
	}
	if err := EnsurePullSecret(ns); err != nil {
		return fmt.Errorf("pull secret: %w", err)
	}

	labels := map[string]string{"app": name, "managed-by": "devplatform"}
	replicas := int32(1)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: ns, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "ghcr-push"}},
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: image,
							Ports: []corev1.ContainerPort{{ContainerPort: port}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	cur, err := Client.AppsV1().Deployments(ns).Get(ctx, "app", metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		if _, err := Client.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("deployment: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("deployment get: %w", err)
	} else {
		cur.Spec.Template.Spec.Containers[0].Image = image
		if _, err := Client.AppsV1().Deployments(ns).Update(ctx, cur, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("deployment update: %w", err)
		}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: ns, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports: []corev1.ServicePort{
				{Port: 80, TargetPort: intstr.FromInt32(port), Protocol: corev1.ProtocolTCP},
			},
		},
	}
	if _, err := Client.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{}); err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("service: %w", err)
	}

	pathType := networkingv1.PathTypePrefix
	ingressClass := "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Labels:      labels,
			Annotations: map[string]string{"cert-manager.io/cluster-issuer": "letsencrypt-prod"},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClass,
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "app",
											Port: networkingv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{{Hosts: []string{hostname}, SecretName: name + "-tls"}},
		},
	}
	if _, err := Client.NetworkingV1().Ingresses(ns).Create(ctx, ingress, metav1.CreateOptions{}); err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("ingress: %w", err)
	}
	return nil
}

func DeployApplication(name, domain string) error {
	ns := "tenant-" + name

	if err := EnsureNamespace(ns); err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("namespace: %w", err)
	}

	if err := createDeployment(ns, name); err != nil {
		return fmt.Errorf("deployment: %w", err)
	}

	if err := createService(ns, name); err != nil {
		return fmt.Errorf("service: %w", err)
	}

	if err := createIngress(ns, name, domain); err != nil {
		return fmt.Errorf("ingress: %w", err)
	}

	return nil
}

func createDeployment(ns, name string) error {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{"app": name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := Client.AppsV1().Deployments(ns).Create(context.Background(), deployment, metav1.CreateOptions{})
	return err
}

func createService(ns, name string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{"app": name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := Client.CoreV1().Services(ns).Create(context.Background(), svc, metav1.CreateOptions{})
	return err
}

func createIngress(ns, name, domain string) error {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: domain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: name,
											Port: networkingv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := Client.NetworkingV1().Ingresses(ns).Create(context.Background(), ingress, metav1.CreateOptions{})
	return err
}

func isAlreadyExists(err error) bool {
	return err != nil && contains(err.Error(), "already exists")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
