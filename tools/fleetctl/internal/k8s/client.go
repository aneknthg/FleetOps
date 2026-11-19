package k8s

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func LoadKubeconfig(path string) (*clientcmdapi.Config, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	if path != "" {
		loader.ExplicitPath = path
	}
	config, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return config, nil
}

func NewClient(path, contextName string) (*kubernetes.Clientset, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	if path != "" {
		loader.ExplicitPath = path
	}
	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build Kubernetes client for context %q: %w", contextName, err)
	}
	config.Timeout = 10 * time.Second
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}
	return client, nil
}

func DefaultKubeconfigPath() string {
	if path := os.Getenv("KUBECONFIG"); path != "" {
		return path
	}
	return ""
}


