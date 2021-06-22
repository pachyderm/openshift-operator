package generators

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func kubeConfig() (*rest.Config, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		// if not running in container get kube config from env or ~/.kube/config
		kubeConfig, ok := os.LookupEnv("KUBECONFIG")
		if !ok {
			kubeConfig = filepath.Join("~/", ".kube", "config")
		}
		conf, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func kubeClientset() (*kubernetes.Clientset, error) {
	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}
