package api

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	ConditionTrue = corev1.ConditionTrue
)

type ObjectMeta = metav1.ObjectMeta
type listOptions = metav1.ListOptions

type Node = corev1.Node
type NodeCondition = corev1.NodeCondition
type NodeConditionType = corev1.NodeConditionType
type NodeStatus = corev1.NodeStatus
type NodeMetrics = metricsv1beta1.NodeMetrics

type Ingress = networkingv1.Ingress
type HTTPIngressPath = networkingv1.HTTPIngressPath
type Service = corev1.Service
type HTTPRoute = gatewayapiv1.HTTPRoute

type Deployment = appsv1.Deployment
type DeploymentSpec = appsv1.DeploymentSpec

type StatefulSet = appsv1.StatefulSet
type StatefulSetSpec = appsv1.StatefulSetSpec

type DaemonSet = appsv1.DaemonSet
type DaemonSetSpec = appsv1.DaemonSetSpec

type LabelSelector = metav1.LabelSelector
type PodTemplateSpec = corev1.PodTemplateSpec
