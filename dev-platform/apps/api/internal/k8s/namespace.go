package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnsureNamespace(name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":   "devplatform",
				"owner": "devplatform",
			},
		},
	}
	_, err := Client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
