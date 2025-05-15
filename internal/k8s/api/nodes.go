package api

import (
	"context"
)

func (c *Client) Nodes(ctx context.Context) ([]Node, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]Node, string, error) {
			nodesList, err := c.kube.CoreV1().Nodes().List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return nodesList.Items, nodesList.Continue, nil
		})
}

func (c *Client) NodeMetrics(ctx context.Context) ([]NodeMetrics, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]NodeMetrics, string, error) {
			metricsList, err := c.metrics.MetricsV1beta1().NodeMetricses().List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return metricsList.Items, metricsList.Continue, nil
		})
}
