package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var Client *kubernetes.Clientset

func Init() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return err
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	}

	Client, err = kubernetes.NewForConfig(config)
	return err
}
