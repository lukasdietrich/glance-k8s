package api

import (
	"context"
)

func (c *Client) Services(ctx context.Context) ([]Service, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]Service, string, error) {
			serviceList, err := c.kube.CoreV1().Services("").List(ctx, opts)
			if err != nil {
				return nil, "", err
			}

			return serviceList.Items, serviceList.Continue, nil
		})
}

func (c *Client) Ingresses(ctx context.Context) ([]Ingress, error) {
	return fetchContinue(ctx,
		func(ctx context.Context, opts listOptions) ([]Ingress, string, error) {
			ingressList, err := c.kube.NetworkingV1().Ingresses("").List(ctx, opts)
			if err != nil {
				return nil, "", err
			}
			return ingressList.Items, ingressList.Continue, nil
		})
}
