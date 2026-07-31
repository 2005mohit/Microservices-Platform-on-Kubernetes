package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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
