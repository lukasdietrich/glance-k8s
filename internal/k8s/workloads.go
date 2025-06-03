package k8s

import (
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"

	"github.com/lukasdietrich/glance-k8s/internal/k8s/api"
)

var (
	_ Workload       = &deployment{}
	_ Workload       = &statefulSet{}
	_ Workload       = &daemonSet{}
	_ sort.Interface = WorkloadSlice{}
)

type WorkloadSlice []Workload

func (w WorkloadSlice) Len() int {
	return len(w)
}

func (w WorkloadSlice) Less(i, j int) bool {
	return w[i].GetName() < w[j].GetName()
}

func (w WorkloadSlice) Swap(i int, j int) {
	w[i], w[j] = w[j], w[i]
}

type Workload interface {
	GetAnnotations() map[string]string
	GetName() string
	GetNamespace() string
	GetSpec() WorkloadSpec
	GetStatus() WorkloadStatus
}

type WorkloadSpec struct {
	Selector *api.LabelSelector
	Template api.PodTemplateSpec
}

type WorkloadStatus struct {
	Replicas      int32
	ReadyReplicas int32
}

func (s WorkloadStatus) Ready() bool {
	return s.ReadyReplicas == s.Replicas
}

type deployment struct {
	api.Deployment
}

func (d deployment) GetAnnotations() map[string]string {
	return lo.Assign(d.Spec.Template.GetAnnotations(), d.Deployment.GetAnnotations())
}

func (d deployment) GetSpec() WorkloadSpec {
	return WorkloadSpec{
		Selector: d.Spec.Selector,
		Template: d.Spec.Template,
	}
}

func (d deployment) GetStatus() WorkloadStatus {
	return WorkloadStatus{
		Replicas:      d.Status.Replicas,
		ReadyReplicas: d.Status.ReadyReplicas,
	}
}

type statefulSet struct {
	api.StatefulSet
}

func (s statefulSet) GetAnnotations() map[string]string {
	return lo.Assign(s.Spec.Template.GetAnnotations(), s.StatefulSet.GetAnnotations())
}

func (s statefulSet) GetSpec() WorkloadSpec {
	return WorkloadSpec{
		Selector: s.Spec.Selector,
		Template: s.Spec.Template,
	}
}

func (s statefulSet) GetStatus() WorkloadStatus {
	return WorkloadStatus{
		Replicas:      s.Status.Replicas,
		ReadyReplicas: s.Status.ReadyReplicas,
	}
}

type daemonSet struct {
	api.DaemonSet
}

func (d daemonSet) GetAnnotations() map[string]string {
	return lo.Assign(d.Spec.Template.GetAnnotations(), d.DaemonSet.GetAnnotations())
}

func (d daemonSet) GetSpec() WorkloadSpec {
	return WorkloadSpec{
		Selector: d.Spec.Selector,
		Template: d.Spec.Template,
	}
}

func (d daemonSet) GetStatus() WorkloadStatus {
	return WorkloadStatus{
		Replicas:      d.Status.DesiredNumberScheduled,
		ReadyReplicas: d.Status.NumberReady,
	}
}

func (c *Cluster) workloads(ctx context.Context) ([]Workload, error) {
	deployments, err := c.client.Deployments(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch deployments: %w", err)
	}

	statefulSets, err := c.client.StatefulSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch statefulsets: %w", err)
	}

	daemonSets, err := c.client.DaemonSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch daemonsets: %w", err)
	}

	workloads := make([]Workload, 0, len(deployments)+len(statefulSets)+len(daemonSets))

	workloads = append(workloads, lo.Map(deployments, wrapDeployment)...)
	workloads = append(workloads, lo.Map(statefulSets, wrapStatefulSet)...)
	workloads = append(workloads, lo.Map(daemonSets, wrapDaemonSet)...)

	return workloads, nil
}

func wrapDeployment(d api.Deployment, _ int) Workload {
	return &deployment{d}
}

func wrapStatefulSet(s api.StatefulSet, _ int) Workload {
	return &statefulSet{s}
}

func wrapDaemonSet(d api.DaemonSet, _ int) Workload {
	return &daemonSet{d}
}
