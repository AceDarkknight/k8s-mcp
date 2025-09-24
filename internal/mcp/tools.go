// Package mcp implements the MCP (Model Context Protocol) server for Kubernetes management.
// 包 mcp 实现了 Kubernetes 管理的 MCP (Model Context Protocol) 服务器。
package mcp

import (
	"context"
	"fmt"
	"strings"

	"k8s-mcp/internal/k8s"
)

// HandleListTools handles tools/list requests
// HandleListTools 处理工具列表请求
func (s *Server) HandleListTools() (*ListToolsResult, error) {
	tools := []Tool{
		{
			Name:        "list_clusters",
			Title:       "List Clusters",
			Description: "List all available Kubernetes clusters",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "switch_cluster",
			Title:       "Switch Cluster",
			Description: "Switch to a different Kubernetes cluster",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cluster_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the cluster to switch to",
					},
				},
				"required": []string{"cluster_name"},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "get_current_cluster",
			Title:       "Get Current Cluster",
			Description: "Get the name of the currently active cluster",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "list_namespaces",
			Title:       "List Namespaces",
			Description: "List all namespaces in the specified or current cluster",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cluster_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the cluster (optional, uses current cluster if not specified)",
					},
				},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "list_resources",
			Title:       "List Resources",
			Description: "List Kubernetes resources of a specific type",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of resource to list (pods, services, deployments, configmaps, secrets, namespaces, nodes, events)",
						"enum":        []string{"pods", "services", "deployments", "configmaps", "secrets", "namespaces", "nodes", "events"},
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace to list resources from (optional for cluster-scoped resources)",
					},
					"cluster_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the cluster (optional, uses current cluster if not specified)",
					},
				},
				"required": []string{"resource_type"},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "get_resource",
			Title:       "Get Resource Details",
			Description: "Get detailed information about a specific Kubernetes resource",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of resource",
						"enum":        []string{"pods", "services", "deployments", "configmaps", "secrets", "namespaces", "nodes"},
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the resource",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the resource (optional for cluster-scoped resources)",
					},
					"cluster_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the cluster (optional, uses current cluster if not specified)",
					},
				},
				"required": []string{"resource_type", "name"},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
		{
			Name:        "describe_resource",
			Title:       "Describe Resource",
			Description: "Get a detailed description of a Kubernetes resource in JSON format",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of resource",
						"enum":        []string{"pods", "services", "deployments", "configmaps", "secrets", "namespaces", "nodes"},
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the resource",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace of the resource (optional for cluster-scoped resources)",
					},
					"cluster_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the cluster (optional, uses current cluster if not specified)",
					},
				},
				"required": []string{"resource_type", "name"},
			},
			Annotations: &ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  false,
			},
		},
	}

	return &ListToolsResult{
		Tools: tools,
	}, nil
}

// HandleCallTool handles tools/call requests
// HandleCallTool 处理工具调用请求
func (s *Server) HandleCallTool(req *CallToolRequest) (*CallToolResult, error) {
	ctx := context.Background()

	switch req.Name {
	case "list_clusters":
		return s.handleListClusters(ctx)
	case "switch_cluster":
		return s.handleSwitchCluster(ctx, req.Arguments)
	case "get_current_cluster":
		return s.handleGetCurrentCluster(ctx)
	case "list_namespaces":
		return s.handleListNamespaces(ctx, req.Arguments)
	case "list_resources":
		return s.handleListResources(ctx, req.Arguments)
	case "get_resource":
		return s.handleGetResource(ctx, req.Arguments)
	case "describe_resource":
		return s.handleDescribeResource(ctx, req.Arguments)
	default:
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", req.Name),
				},
			},
			IsError: true,
		}, nil
	}
}

// Tool handlers
// 工具处理函数

// handleListClusters lists available clusters
// handleListClusters 列出可用集群
func (s *Server) handleListClusters(ctx context.Context) (*CallToolResult, error) {
	clusters := s.clusterManager.GetClusters()
	current := s.clusterManager.GetCurrentCluster()

	var clusterList []string
	for _, cluster := range clusters {
		if cluster == current {
			clusterList = append(clusterList, fmt.Sprintf("%s (current)", cluster))
		} else {
			clusterList = append(clusterList, cluster)
		}
	}

	text := fmt.Sprintf("Available clusters:\n%s", strings.Join(clusterList, "\n"))

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

// handleSwitchCluster switches to a different cluster
// handleSwitchCluster 切换到不同的集群
func (s *Server) handleSwitchCluster(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	clusterName, ok := args["cluster_name"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "cluster_name parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	err := s.clusterManager.SwitchCluster(clusterName)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to switch to cluster %s: %v", clusterName, err),
				},
			},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully switched to cluster: %s", clusterName),
			},
		},
	}, nil
}

// handleGetCurrentCluster gets the current cluster
// handleGetCurrentCluster 获取当前集群
func (s *Server) handleGetCurrentCluster(ctx context.Context) (*CallToolResult, error) {
	current := s.clusterManager.GetCurrentCluster()
	if current == "" {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "No current cluster set",
				},
			},
		}, nil
	}

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: fmt.Sprintf("Current cluster: %s", current),
			},
		},
	}, nil
}

// handleListNamespaces lists namespaces
// handleListNamespaces 列出命名空间
func (s *Server) handleListNamespaces(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	clusterName, _ := args["cluster_name"].(string)

	namespaces, err := s.resourceOps.ListNamespaces(ctx, clusterName)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to list namespaces: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	var nameList []string
	for _, ns := range namespaces {
		nameList = append(nameList, ns.Name)
	}

	text := fmt.Sprintf("Namespaces in cluster %s:\n%s", clusterName, strings.Join(nameList, "\n")) // 集群 %s 中的命名空间：\n%s

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

// handleListResources lists resources by type
// handleListResources 按类型列出资源
func (s *Server) handleListResources(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	resourceType, ok := args["resource_type"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "resource_type parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	namespace, _ := args["namespace"].(string)
	clusterName, _ := args["cluster_name"].(string)

	resources, err := s.resourceOps.ListResourcesByType(ctx, k8s.ResourceType(resourceType), namespace, clusterName)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to list %s: %v", resourceType, err),
				},
			},
			IsError: true,
		}, nil
	}

	if len(resources) == 0 {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("No %s found", resourceType), // 未找到 %s
				},
			},
		}, nil
	}

	var resourceList []string
	for _, resource := range resources {
		if resource.Namespace != "" {
			resourceList = append(resourceList, fmt.Sprintf("- %s/%s (%s) - %s", resource.Namespace, resource.Name, resource.Kind, resource.Status)) // - %s/%s (%s) - %s
		} else {
			resourceList = append(resourceList, fmt.Sprintf("- %s (%s) - %s", resource.Name, resource.Kind, resource.Status)) // - %s (%s) - %s
		}
	}

	text := fmt.Sprintf("%s:\n%s", resourceType, strings.Join(resourceList, "\n")) // %s：\n%s

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

// handleGetResource gets resource details
// handleGetResource 获取资源详情
func (s *Server) handleGetResource(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	resourceType, ok := args["resource_type"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "resource_type parameter is required", // resource_type 参数是必需的
				},
			},
			IsError: true,
		}, nil
	}

	name, ok := args["name"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "name parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	namespace, _ := args["namespace"].(string)
	clusterName, _ := args["cluster_name"].(string)

	resource, err := s.resourceOps.GetResourceDetails(ctx, k8s.ResourceType(resourceType), namespace, name, clusterName)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get %s/%s: %v", resourceType, name, err), // 获取 %s/%s 失败：%v
				},
			},
			IsError: true,
		}, nil
	}

	// Convert resource to JSON string
	jsonStr, err := s.resourceOps.SerializeResource(resource)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to serialize resource: %v", err), // 序列化资源失败：%v
				},
			},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: jsonStr,
			},
		},
	}, nil
}

// handleDescribeResource describes a resource
// handleDescribeResource 描述资源
func (s *Server) handleDescribeResource(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	resourceType, ok := args["resource_type"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "resource_type parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	name, ok := args["name"].(string)
	if !ok {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: "name parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	namespace, _ := args["namespace"].(string)
	clusterName, _ := args["cluster_name"].(string)

	description, err := s.resourceOps.DescribeResource(ctx, k8s.ResourceType(resourceType), namespace, name, clusterName)
	if err != nil {
		return &CallToolResult{
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to describe %s/%s: %v", resourceType, name, err), // 描述 %s/%s 失败：%v
				},
			},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: description,
			},
		},
	}, nil
}
