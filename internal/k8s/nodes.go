package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lukasdietrich/glance-k8s/internal/k8s/api"
	"github.com/samber/lo"
)

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

type Node struct {
	api.ObjectMeta
	Status  api.NodeStatus
	Metrics api.NodeMetrics
}

func (n *Node) ConditionTrue(conditionType api.NodeConditionType) bool {
	for _, condition := range n.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == api.ConditionTrue
		}
	}

	return false
}

func (n *Node) ConditionTransition(conditionType api.NodeConditionType) time.Time {
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

func (c *Cluster) Nodes(ctx context.Context) (NodeSlice, error) {
	nodeInfos, err := c.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch nodes: %w", err)
	}

	nodeMetrics, err := c.client.NodeMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch node metrics: %w", err)
	}

	nodes := NodeSlice(lo.Map(nodeInfos, wrapNodeWithMetrics(nodeMetrics)))
	sort.Stable(nodes)
	return nodes, nil
}

func wrapNodeWithMetrics(metricsSlice []api.NodeMetrics) func(api.Node, int) Node {
	metricsMap := lo.SliceToMap(
		metricsSlice,
		func(metrics api.NodeMetrics) (string, api.NodeMetrics) {
			return metrics.Name, metrics
		},
	)

	return func(node api.Node, _ int) Node {
		metrics := metricsMap[node.Name]

		return Node{
			ObjectMeta: node.ObjectMeta,
			Status:     node.Status,
			Metrics:    metrics,
		}
	}
}
