package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceType represents supported k8s resource types
type ResourceType string

const (
	ResourceTypePods        ResourceType = "pods"
	ResourceTypePod         ResourceType = "pod"
	ResourceTypeServices    ResourceType = "services"
	ResourceTypeService     ResourceType = "service"
	ResourceTypeDeployments ResourceType = "deployments"
	ResourceTypeDeployment  ResourceType = "deployment"
	ResourceTypeConfigMaps  ResourceType = "configmaps"
	ResourceTypeConfigMap   ResourceType = "configmap"
	ResourceTypeSecrets     ResourceType = "secrets"
	ResourceTypeSecret      ResourceType = "secret"
	ResourceTypeNamespaces  ResourceType = "namespaces"
	ResourceTypeNamespace   ResourceType = "namespace"
	ResourceTypeNodes       ResourceType = "nodes"
	ResourceTypeNode        ResourceType = "node"
	ResourceTypeEvents      ResourceType = "events"
	ResourceTypeEvent       ResourceType = "event"
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

// ListNamespaces lists all namespaces in current cluster
func (ro *ResourceOperations) ListNamespaces(ctx context.Context, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetCurrentClient()
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
		client, err = ro.clusterManager.GetCurrentClient()
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
			Status:    getPodStatus(&pod),
			Age:       pod.CreationTimestamp.String(),
			Labels:    pod.Labels,
		})
	}

	return resources, nil
}

// getPodStatus calculates a high-level status for a pod, similar to kubectl
// getPodStatus 计算Pod的高级状态，类似于kubectl
func getPodStatus(pod *corev1.Pod) string {
	// 1. Check if pod is being deleted
	// 1. 检查 Pod 是否正在删除
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	// 2. Check failed reason
	// 2. 检查失败原因
	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}

	// 3. Check Init Containers
	// 3. 检查 Init 容器
	// Init containers run before main containers and are used for setup tasks
	// Init 容器在主容器之前运行，用于设置任务
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		// Check if init container terminated with error
		// 检查 init 容器是否以错误状态终止
		if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
			if containerStatus.State.Terminated.Reason != "" {
				return "Init:" + containerStatus.State.Terminated.Reason
			}
			return "Init:Error"
		}
		// Check if init container is waiting with a reason
		// 检查 init 容器是否因某个原因在等待
		if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason != "" {
			// PodInitializing is a normal transient state, don't report it
			// PodInitializing 是一个正常的瞬态，不报告它
			if containerStatus.State.Waiting.Reason != "PodInitializing" {
				return "Init:" + containerStatus.State.Waiting.Reason
			}
		}
	}

	// 4. Check Containers
	// 4. 检查主容器
	// Priority: Waiting (CrashLoopBackOff etc.) > Terminated (Error) > Running
	// 优先级：等待（如 CrashLoopBackOff）> 终止（错误）> 运行
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Check if container is waiting with a reason (highest priority)
		// 检查容器是否因某个原因在等待（最高优先级）
		if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason != "" {
			return containerStatus.State.Waiting.Reason
		}
		// Check if container terminated with error (second priority)
		// 检查容器是否以错误状态终止（第二优先级）
		if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
			if containerStatus.State.Terminated.Reason != "" {
				return containerStatus.State.Terminated.Reason
			}
			return "Error"
		}
	}

	// 5. If everything looks fine, return the Phase (Running, Pending, Succeeded)
	// 5. 如果一切看起来正常，返回 Phase（Running, Pending, Succeeded）
	return string(pod.Status.Phase)
}

// ListServices lists services in a namespace
func (ro *ResourceOperations) ListServices(ctx context.Context, namespace, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetCurrentClient()
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
		client, err = ro.clusterManager.GetCurrentClient()
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
		client, err = ro.clusterManager.GetCurrentClient()
	}
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case ResourceTypePods, ResourceTypePod:
		return client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeServices, ResourceTypeService:
		return client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeDeployments, ResourceTypeDeployment:
		return client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeConfigMaps, ResourceTypeConfigMap:
		return client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeSecrets, ResourceTypeSecret:
		return client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeNamespaces, ResourceTypeNamespace:
		return client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	case ResourceTypeNodes, ResourceTypeNode:
		return client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ListResourcesByType lists resources of a specific type
func (ro *ResourceOperations) ListResourcesByType(ctx context.Context, resourceType ResourceType, namespace, clusterName string) ([]ResourceInfo, error) {
	switch resourceType {
	case ResourceTypePods, ResourceTypePod:
		return ro.ListPods(ctx, namespace, clusterName)
	case ResourceTypeServices, ResourceTypeService:
		return ro.ListServices(ctx, namespace, clusterName)
	case ResourceTypeDeployments, ResourceTypeDeployment:
		return ro.ListDeployments(ctx, namespace, clusterName)
	case ResourceTypeNamespaces, ResourceTypeNamespace:
		return ro.ListNamespaces(ctx, clusterName)
	case ResourceTypeConfigMaps, ResourceTypeConfigMap:
		return ro.listConfigMaps(ctx, namespace, clusterName)
	case ResourceTypeSecrets, ResourceTypeSecret:
		return ro.listSecrets(ctx, namespace, clusterName)
	case ResourceTypeNodes, ResourceTypeNode:
		return ro.listNodes(ctx, clusterName)
	case ResourceTypeEvents, ResourceTypeEvent:
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
		client, err = ro.clusterManager.GetCurrentClient()
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
		client, err = ro.clusterManager.GetCurrentClient()
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

// listNodes lists nodes in cluster
func (ro *ResourceOperations) listNodes(ctx context.Context, clusterName string) ([]ResourceInfo, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetCurrentClient()
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
		client, err = ro.clusterManager.GetCurrentClient()
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
		ResourceTypePods,
		ResourceTypePod,
		ResourceTypeServices,
		ResourceTypeService,
		ResourceTypeDeployments,
		ResourceTypeDeployment,
		ResourceTypeConfigMaps,
		ResourceTypeConfigMap,
		ResourceTypeSecrets,
		ResourceTypeSecret,
		ResourceTypeNamespaces,
		ResourceTypeNamespace,
		ResourceTypeNodes,
		ResourceTypeNode,
		ResourceTypeEvents,
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
		client, err = ro.clusterManager.GetCurrentClient()
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

// GetPodLogs retrieves logs from a pod
// GetPodLogs 从 Pod 获取日志
func (ro *ResourceOperations) GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines *int64, previous bool, clusterName string) (string, error) {
	var client *kubernetes.Clientset
	var err error

	if clusterName != "" {
		client, err = ro.clusterManager.GetClientForCluster(clusterName)
	} else {
		client, err = ro.clusterManager.GetCurrentClient()
	}
	if err != nil {
		return "", err
	}

	// Default tail lines to 100 if not specified
	// 如果未指定，默认 tail lines 为 100
	if tailLines == nil {
		defaultLines := int64(100)
		tailLines = &defaultLines
	}

	// Get pod to determine container name if not specified
	// 如果未指定容器名称，获取 Pod 以确定容器名称
	if containerName == "" {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get pod: %w", err)
		}
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			return "", fmt.Errorf("no containers found in pod %s", podName)
		}
	}

	// Create log request options
	// 创建日志请求选项
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: tailLines,
		Previous:  previous,
	}

	// Get logs as a stream
	// 获取日志流
	req := client.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logStream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get log stream: %w", err)
	}
	defer logStream.Close()

	// Read logs with a limit to prevent memory issues
	// 读取日志并限制大小以防止内存问题
	const maxBytes = 1 * 1024 * 1024 // 1MB
	limitedReader := io.LimitReader(logStream, maxBytes)
	logBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	logs := string(logBytes)

	// Check if logs were truncated
	// 检查日志是否被截断
	if int64(len(logBytes)) >= maxBytes {
		logs += "\n\n[Logs truncated: exceeded 1MB limit]"
	}

	return logs, nil
}

// CheckRBACPermission checks if the current user has permission to perform an action
// CheckRBACPermission 检查当前用户是否有权限执行某个操作
func (ro *ResourceOperations) CheckRBACPermission(ctx context.Context, verb, resource, namespace string) (bool, error) {
	var client *kubernetes.Clientset
	var err error

	client, err = ro.clusterManager.GetCurrentClient()
	if err != nil {
		return false, err
	}

	// Create SelfSubjectAccessReview to check permissions
	// 创建 SelfSubjectAccessReview 来检查权限
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Resource:  resource,
			},
		},
	}

	// Create the review
	// 创建审查
	response, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to check RBAC permission: %w", err)
	}

	return response.Status.Allowed, nil
}
