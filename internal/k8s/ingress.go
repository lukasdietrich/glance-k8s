package k8s

import (
	"context"
	"fmt"
	"net/url"

	"github.com/samber/lo"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scheme string

type Ingress struct {
	metav1.ObjectMeta
	URL url.URL
}

func (c *Client) ListIngress(ctx context.Context) ([]Ingress, error) {
	opts := metav1.ListOptions{}

	ingressList, err := c.kube.NetworkingV1().Ingresses("").List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ingress list: %w", err)
	}

	return lo.Map(ingressList.Items, mapIngress), nil
}

func mapIngress(ingress networkingv1.Ingress, _ int) Ingress {
	return Ingress{
		ObjectMeta: ingress.ObjectMeta,
		URL: url.URL{
			Scheme: mapIngressScheme(ingress),
			Host:   mapIngressHost(ingress),
		},
	}
}

func mapIngressScheme(ingress networkingv1.Ingress) string {
	if len(ingress.Spec.TLS) > 0 {
		return "https"
	}

	return "http"
}

func mapIngressHost(ingress networkingv1.Ingress) string {
	if rules := ingress.Spec.Rules; len(rules) > 0 {
		return rules[0].Host
	}

	return ""
}
