package k8s

import (
	"context"

	"github.com/lukasdietrich/glance-k8s/internal/k8s/api"
)

// apiClient is the subset of api.Client used by Cluster, defined on the
// consumer side so it can be wrapped with a caching decorator.
type apiClient interface {
	Deployments(ctx context.Context) ([]api.Deployment, error)
	StatefulSets(ctx context.Context) ([]api.StatefulSet, error)
	DaemonSets(ctx context.Context) ([]api.DaemonSet, error)
	Services(ctx context.Context) ([]api.Service, error)
	Ingresses(ctx context.Context) ([]api.Ingress, error)
	HTTPRoutes(ctx context.Context) ([]api.HTTPRoute, error)
	Nodes(ctx context.Context) ([]api.Node, error)
	NodeMetrics(ctx context.Context) ([]api.NodeMetrics, error)
}

// cachedClient wraps an apiClient with one read-through cache per
// resource type. See cache.go for TTL and concurrency semantics.
type cachedClient struct {
	inner apiClient

	deployments  cache[api.Deployment]
	statefulSets cache[api.StatefulSet]
	daemonSets   cache[api.DaemonSet]
	services     cache[api.Service]
	ingresses    cache[api.Ingress]
	httpRoutes   cache[api.HTTPRoute]
	nodes        cache[api.Node]
	nodeMetrics  cache[api.NodeMetrics]
}

func newCachedClient(inner apiClient) *cachedClient {
	return &cachedClient{inner: inner}
}

func (c *cachedClient) Deployments(ctx context.Context) ([]api.Deployment, error) {
	return c.deployments.get(ctx, c.inner.Deployments)
}

func (c *cachedClient) StatefulSets(ctx context.Context) ([]api.StatefulSet, error) {
	return c.statefulSets.get(ctx, c.inner.StatefulSets)
}

func (c *cachedClient) DaemonSets(ctx context.Context) ([]api.DaemonSet, error) {
	return c.daemonSets.get(ctx, c.inner.DaemonSets)
}

func (c *cachedClient) Services(ctx context.Context) ([]api.Service, error) {
	return c.services.get(ctx, c.inner.Services)
}

func (c *cachedClient) Ingresses(ctx context.Context) ([]api.Ingress, error) {
	return c.ingresses.get(ctx, c.inner.Ingresses)
}

func (c *cachedClient) HTTPRoutes(ctx context.Context) ([]api.HTTPRoute, error) {
	return c.httpRoutes.get(ctx, c.inner.HTTPRoutes)
}

func (c *cachedClient) Nodes(ctx context.Context) ([]api.Node, error) {
	return c.nodes.get(ctx, c.inner.Nodes)
}

func (c *cachedClient) NodeMetrics(ctx context.Context) ([]api.NodeMetrics, error) {
	return c.nodeMetrics.get(ctx, c.inner.NodeMetrics)
}
