package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Node struct {
	metav1.ObjectMeta
	Status  corev1.NodeStatus
	Metrics metricsv1beta1.NodeMetrics
}

func (n *Node) ConditionTrue(conditionType corev1.NodeConditionType) bool {
	for _, condition := range n.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func (n *Node) ConditionTransition(conditionType corev1.NodeConditionType) time.Time {
	for _, condition := range n.Status.Conditions {
		if condition.Type == conditionType {
			return condition.LastTransitionTime.Time
		}
	}

	return time.UnixMilli(0)
}

func (n *Node) Roles() []string {
	const nodeRolePrefix = "node-role.kubernetes.io/"

	var roles []string
	for label, value := range n.Labels {
		if strings.HasPrefix(label, nodeRolePrefix) && value == "true" {
			roles = append(roles, label[len(nodeRolePrefix):])
		}
	}

	sort.Strings(roles)
	return roles
}

func (n *Node) CpuRatio() float64 {
	usage := n.Metrics.Usage.Cpu().AsApproximateFloat64()
	capacity := n.Status.Capacity.Cpu().AsApproximateFloat64()

	return usage / capacity
}

func (n *Node) MemRatio() float64 {
	usage := n.Metrics.Usage.Memory().AsApproximateFloat64()
	capacity := n.Status.Capacity.Memory().AsApproximateFloat64()

	return usage / capacity
}

type NodeSlice []Node

func (n NodeSlice) Len() int {
	return len(n)
}

func (n NodeSlice) Less(i, j int) bool {
	return n[i].Name < n[j].Name
}

func (n NodeSlice) Swap(i int, j int) {
	n[i], n[j] = n[j], n[i]
}

func (c *Client) ListNodes(ctx context.Context) (NodeSlice, error) {
	clusterNodes, err := fetchContinue(ctx, c.listNodes)
	if err != nil {
		return nil, fmt.Errorf("could not fetch nodes: %w", err)
	}

	clusterNodeMetrics, err := fetchContinue(ctx, c.listNodeMetrics)
	if err != nil {
		return nil, fmt.Errorf("could not fetch node metrics: %w", err)
	}

	nodes := NodeSlice(lo.Map(clusterNodes, mapNode(clusterNodeMetrics)))
	sort.Stable(nodes)

	return nodes, nil
}

func (c *Client) listNodes(ctx context.Context, opts metav1.ListOptions) ([]corev1.Node, string, error) {
	nodesList, err := c.kube.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	return nodesList.Items, nodesList.Continue, nil
}

func (c *Client) listNodeMetrics(ctx context.Context, opts metav1.ListOptions) ([]metricsv1beta1.NodeMetrics, string, error) {
	metricsList, err := c.metrics.MetricsV1beta1().NodeMetricses().List(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	return metricsList.Items, metricsList.Continue, nil
}

func mapNode(metricsSlice []metricsv1beta1.NodeMetrics) func(corev1.Node, int) Node {
	metricsMap := lo.SliceToMap(
		metricsSlice,
		func(metrics metricsv1beta1.NodeMetrics) (string, metricsv1beta1.NodeMetrics) {
			return metrics.Name, metrics
		},
	)

	return func(node corev1.Node, _ int) Node {
		metrics := metricsMap[node.Name]

		return Node{
			ObjectMeta: node.ObjectMeta,
			Status:     node.Status,
			Metrics:    metrics,
		}
	}
}
