package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceType represents supported k8s resource types
type ResourceType string

const (
	ResourceTypePod        ResourceType = "pods"
	ResourceTypeService    ResourceType = "services"
	ResourceTypeDeployment ResourceType = "deployments"
	ResourceTypeConfigMap  ResourceType = "configmaps"
	ResourceTypeSecret     ResourceType = "secrets"
	ResourceTypeNamespace  ResourceType = "namespaces"
	ResourceTypeNode       ResourceType = "nodes"
	ResourceTypeEvent      ResourceType = "events"
)

// ResourceInfo holds basic information about a k8s resource
type ResourceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Kind      string            `json:"kind"`
	Status    string            `json:"status,omitempty"`
	Age       string            `json:"age,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ResourceOperations provides k8s resource operations
type ResourceOperations struct {
	clusterManager *ClusterManager
}

// NewResourceOperations creates a new resource operations instance
func NewResourceOperations(cm *ClusterManager) *ResourceOperations {
	return &ResourceOperations{
		clusterManager: cm,
	}
}

// ListNamespaces lists all namespaces in the current cluster
func (ro *ResourceOperations) ListNamespaces(ctx context.Context, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var resources []ResourceInfo
	for _, ns := range namespaces.Items {
		resources = append(resources, ResourceInfo{
			Name:   ns.Name,
			Kind:   "Namespace",
			Status: string(ns.Status.Phase),
			Age:    ns.CreationTimestamp.String(),
			Labels: ns.Labels,
		})
	}

	return resources, nil
}

// ListPods lists pods in a namespace
func (ro *ResourceOperations) ListPods(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var resources []ResourceInfo
	for _, pod := range pods.Items {
		resources = append(resources, ResourceInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Kind:      "Pod",
			Status:    string(pod.Status.Phase),
			Age:       pod.CreationTimestamp.String(),
			Labels:    pod.Labels,
		})
	}

	return resources, nil
}

// ListServices lists services in a namespace
func (ro *ResourceOperations) ListServices(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	services, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var resources []ResourceInfo
	for _, svc := range services.Items {
		resources = append(resources, ResourceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Kind:      "Service",
			Status:    fmt.Sprintf("Type: %s", svc.Spec.Type),
			Age:       svc.CreationTimestamp.String(),
			Labels:    svc.Labels,
		})
	}

	return resources, nil
}

// ListDeployments lists deployments in a namespace
func (ro *ResourceOperations) ListDeployments(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var resources []ResourceInfo
	for _, dep := range deployments.Items {
		status := fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas)
		resources = append(resources, ResourceInfo{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Kind:      "Deployment",
			Status:    status,
			Age:       dep.CreationTimestamp.String(),
			Labels:    dep.Labels,
		})
	}

	return resources, nil
}

// GetResourceDetails gets detailed information about a specific resource
func (ro *ResourceOperations) GetResourceDetails(ctx context.Context, resourceType ResourceType, namespace, name, clusterName string) (interface{}, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case ResourceTypePod:
		return client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeService:
		return client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeDeployment:
		return client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeConfigMap:
		return client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeSecret:
		return client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeNamespace:
		return client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeNode:
		return client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ListResourcesByType lists resources of a specific type
func (ro *ResourceOperations) ListResourcesByType(ctx context.Context, resourceType ResourceType, namespace, clusterName string) ([]ResourceInfo, error) {
	switch resourceType {
	case ResourceTypePod:
		return ro.ListPods(ctx, namespace, clusterName)
	case ResourceTypeService:
		return ro.ListServices(ctx, namespace, clusterName)
	case ResourceTypeDeployment:
		return ro.ListDeployments(ctx, namespace, clusterName)
	case ResourceTypeNamespace:
		return ro.ListNamespaces(ctx, clusterName)
	case ResourceTypeConfigMap:
		return ro.listConfigMaps(ctx, namespace, clusterName)
	case ResourceTypeSecret:
		return ro.listSecrets(ctx, namespace, clusterName)
	case ResourceTypeNode:
		return ro.listNodes(ctx, clusterName)
	case ResourceTypeEvent:
		return ro.listEvents(ctx, namespace, clusterName)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// listConfigMaps lists configmaps in a namespace
func (ro *ResourceOperations) listConfigMaps(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	configMaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	var resources []ResourceInfo
	for _, cm := range configMaps.Items {
		resources = append(resources, ResourceInfo{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			Kind:      "ConfigMap",
			Status:    fmt.Sprintf("%d keys", len(cm.Data)),
			Age:       cm.CreationTimestamp.String(),
			Labels:    cm.Labels,
		})
	}

	return resources, nil
}

// listSecrets lists secrets in a namespace
func (ro *ResourceOperations) listSecrets(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	secrets, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var resources []ResourceInfo
	for _, secret := range secrets.Items {
		resources = append(resources, ResourceInfo{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Kind:      "Secret",
			Status:    fmt.Sprintf("Type: %s", secret.Type),
			Age:       secret.CreationTimestamp.String(),
			Labels:    secret.Labels,
		})
	}

	return resources, nil
}

// listNodes lists nodes in the cluster
func (ro *ResourceOperations) listNodes(ctx context.Context, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var resources []ResourceInfo
	for _, node := range nodes.Items {
		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		resources = append(resources, ResourceInfo{
			Name:   node.Name,
			Kind:   "Node",
			Status: status,
			Age:    node.CreationTimestamp.String(),
			Labels: node.Labels,
		})
	}

	return resources, nil
}

// listEvents lists events in a namespace
func (ro *ResourceOperations) listEvents(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var resources []ResourceInfo
	for _, event := range events.Items {
		resources = append(resources, ResourceInfo{
			Name:      event.Name,
			Namespace: event.Namespace,
			Kind:      "Event",
			Status:    fmt.Sprintf("%s: %s", event.Type, event.Reason),
			Age:       event.CreationTimestamp.String(),
		})
	}

	return resources, nil
}

// GetSupportedResourceTypes returns all supported resource types
func (ro *ResourceOperations) GetSupportedResourceTypes() []ResourceType {
	return []ResourceType{
		ResourceTypePod,
		ResourceTypeService,
		ResourceTypeDeployment,
		ResourceTypeConfigMap,
		ResourceTypeSecret,
		ResourceTypeNamespace,
		ResourceTypeNode,
		ResourceTypeEvent,
	}
}

// SerializeResource converts a k8s resource to JSON string
func (ro *ResourceOperations) SerializeResource(resource interface{}) (string, error) {
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize resource: %w", err)
	}
	return string(data), nil
}

// DescribeResource provides detailed description of a resource
func (ro *ResourceOperations) DescribeResource(ctx context.Context, resourceType ResourceType, namespace, name, clusterName string) (string, error) {
	resource, err := ro.GetResourceDetails(ctx, resourceType, namespace, name, clusterName)
	if err != nil {
		return "", err
	}

	// Convert to JSON for detailed description
	jsonStr, err := ro.SerializeResource(resource)
	if err != nil {
		return "", err
	}

	return jsonStr, nil
}

// GetClusterInfo gets basic cluster information
func (ro *ResourceOperations) GetClusterInfo(ctx context.Context, clusterName string) (map[string]interface{}, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetClient()
	}
	if err != nil {
		return nil, err
	}

	// Get server version
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get nodes for basic cluster info
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get namespaces count
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	info := map[string]interface{}{
		"version":        version.GitVersion,
		"platform":       version.Platform,
		"nodeCount":      len(nodes.Items),
		"namespaceCount": len(namespaces.Items),
		"buildDate":      version.BuildDate,
	}

	return info, nil
}
