package api

import (
	"context"
)

func (c *Client) Deployments(ctx context.Context) ([]Deployment, error) {
	return cachedList(ctx, c.cache, "deployments", c.listDeployments)
}

func (c *Client) listDeployments(ctx context.Context) ([]Deployment, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]Deployment, string, error) {
			deploymentList, err := c.kube.AppsV1().Deployments("").List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return deploymentList.Items, deploymentList.Continue, err
		})
}

func (c *Client) StatefulSets(ctx context.Context) ([]StatefulSet, error) {
	return cachedList(ctx, c.cache, "statefulsets", c.listStatefulSets)
}

func (c *Client) listStatefulSets(ctx context.Context) ([]StatefulSet, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]StatefulSet, string, error) {
			statefulSetList, err := c.kube.AppsV1().StatefulSets("").List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return statefulSetList.Items, statefulSetList.Continue, err
		})
}

func (c *Client) DaemonSets(ctx context.Context) ([]DaemonSet, error) {
	return cachedList(ctx, c.cache, "daemonsets", c.listDaemonSets)
}

func (c *Client) listDaemonSets(ctx context.Context) ([]DaemonSet, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]DaemonSet, string, error) {
			daemonSetList, err := c.kube.AppsV1().DaemonSets("").List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return daemonSetList.Items, daemonSetList.Continue, err
		})
}
