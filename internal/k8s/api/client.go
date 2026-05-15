package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	gatewayv "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

type Client struct {
	kube    *kubernetes.Clientset
	metrics *metricsv.Clientset
	gateway *gatewayv.Clientset
}

func Connect() (*Client, error) {
	config, err := readKubernetesConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read kubernetes config: %w", err)
	}

	applyRateLimitOverrides(config)

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes client: %w", err)
	}

	metrics, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create metrics client: %w", err)
	}

	gateway, err := gatewayv.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create gatewayClientset client: %w", err)
	}

	return &Client{
		kube:    kube,
		metrics: metrics,
		gateway: gateway,
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

// applyRateLimitOverrides lets operators raise client-go's default
// rate limit (5 QPS / 10 burst) when a dashboard fans out enough
// parallel widget calls to exhaust it. Leaves the defaults in place
// when the env vars are unset or malformed.
func applyRateLimitOverrides(config *rest.Config) {
	if v := os.Getenv("GLANCE_KUBE_API_QPS"); v != "" {
		if qps, err := strconv.ParseFloat(v, 32); err == nil && qps > 0 {
			config.QPS = float32(qps)
			slog.Debug("overriding kube api qps", slog.Float64("qps", qps))
		} else {
			slog.Warn("ignoring invalid GLANCE_KUBE_API_QPS", slog.String("value", v))
		}
	}

	if v := os.Getenv("GLANCE_KUBE_API_BURST"); v != "" {
		if burst, err := strconv.Atoi(v); err == nil && burst > 0 {
			config.Burst = burst
			slog.Debug("overriding kube api burst", slog.Int("burst", burst))
		} else {
			slog.Warn("ignoring invalid GLANCE_KUBE_API_BURST", slog.String("value", v))
		}
	}
}

type fetchFunc[Item any] func(context.Context, listOptions) ([]Item, string, error)

func fetchContinue[Item any](ctx context.Context, fetch fetchFunc[Item]) ([]Item, error) {
	var opts listOptions
	var items []Item

	for {
		itemsChunk, continueToken, err := fetch(ctx, opts)
		if err != nil {
			return nil, err
		}

		items = append(items, itemsChunk...)

		if continueToken == "" {
			return items, nil
		}

		slog.Debug("continue api call", slog.String("continue", continueToken))
		opts.Continue = continueToken
	}
}
