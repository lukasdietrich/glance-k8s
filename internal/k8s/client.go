package k8s

import (
	"fmt"
	"log/slog"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	kube    *kubernetes.Clientset
	metrics *metricsv.Clientset
}

func Connect() (*Client, error) {
	config, err := readKubernetesConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read kubernetes config: %w", err)
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not connect to kubernetes: %w", err)
	}

	metrics, err := metricsv.NewForConfig(config)

	return &Client{
		kube:    kube,
		metrics: metrics,
	}, nil
}

func readKubernetesConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		slog.Debug("using in cluster config")
		return config, nil
	}

	filename := findKubernetesConfigFilename()
	config, err = clientcmd.BuildConfigFromFlags("", filename)
	if err == nil {
		slog.Debug("using out of cluster config at", slog.String("filename", filename))
		return config, nil
	}

	return nil, fmt.Errorf("no cluster config present")
}

func findKubernetesConfigFilename() string {
	if filename := os.Getenv("GLANCE_KUBE_CONFIG"); filename != "" {
		return filename
	}

	return os.ExpandEnv("${HOME}/.kube/config")
}
