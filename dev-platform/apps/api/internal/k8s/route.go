package k8s

import (
 "context"
 "fmt"
 "regexp"
 "strings"

 corev1 "k8s.io/api/core/v1"
 networkingv1 "k8s.io/api/networking/v1"
 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 "k8s.io/apimachinery/pkg/util/intstr"
)

func sanitize(name string) string {
 reg := regexp.MustCompile("[^a-zA-Z0-9-]")
 s := reg.ReplaceAllString(name, "-")
 if len(s) > 40 {
  s = s[:40]
 }
 return strings.Trim(s, "-")
}

func EnsureExternalRoute(name, hostname, targetIP string, targetPort int32) error {
 ns := "devplatform"
 svcName := "app-" + sanitize(name)
 ctx := context.Background()

 _ = Client.NetworkingV1().Ingresses(ns).Delete(ctx, svcName, metav1.DeleteOptions{})
 _ = Client.CoreV1().Endpoints(ns).Delete(ctx, svcName, metav1.DeleteOptions{})
 _ = Client.CoreV1().Services(ns).Delete(ctx, svcName, metav1.DeleteOptions{})

 svc := &corev1.Service{
 ObjectMeta: metav1.ObjectMeta{
    Name:   svcName,
    Labels: map[string]string{"app": "devplatform", "managed-by": "devplatform"},
 },
 Spec: corev1.ServiceSpec{
    Ports: []corev1.ServicePort{
        {
            Port:       80,
            TargetPort: intstr.FromInt32(targetPort),
            Protocol:   corev1.ProtocolTCP,
        },
    },
 },
 }
 if _, err := Client.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
    return fmt.Errorf("service: %w", err)
 }

 ep := &corev1.Endpoints{
 ObjectMeta: metav1.ObjectMeta{
    Name:   svcName,
    Labels: map[string]string{"app": "devplatform", "managed-by": "devplatform"},
 },
 Subsets: []corev1.EndpointSubset{
    {
        Addresses: []corev1.EndpointAddress{{IP: targetIP}},
        Ports:     []corev1.EndpointPort{{Port: targetPort, Protocol: corev1.ProtocolTCP}},
    },
 },
 }
 if _, err := Client.CoreV1().Endpoints(ns).Create(ctx, ep, metav1.CreateOptions{}); err != nil {
    return fmt.Errorf("endpoints: %w", err)
 }

 pathType := networkingv1.PathTypePrefix
 ingressClass := "nginx"
 ing := &networkingv1.Ingress{
 ObjectMeta: metav1.ObjectMeta{
    Name:   svcName,
    Labels: map[string]string{"app": "devplatform", "managed-by": "devplatform"},
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
                            Path:    "/",
                            PathType: &pathType,
                            Backend: networkingv1.IngressBackend{
                                Service: &networkingv1.IngressServiceBackend{
                                    Name: svcName,
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
 if _, err := Client.NetworkingV1().Ingresses(ns).Create(ctx, ing, metav1.CreateOptions{}); err != nil {
    return fmt.Errorf("ingress: %w", err)
 }

 return nil
}
