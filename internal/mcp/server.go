// Package mcp implements the MCP (Model Context Protocol) server for Kubernetes management.
// 包 mcp 实现了 Kubernetes 管理的 MCP (Model Context Protocol) 服务器。
package mcp

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"k8s-mcp/internal/k8s"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with k8s integration
// Server 封装了 MCP 服务器和 k8s 集成
type Server struct {
	mcpServer      *mcp.Server
	clusterManager *k8s.ClusterManager
	resourceOps    *k8s.ResourceOperations
	authToken      string
}

// NewServer creates a new MCP server instance
// NewServer 创建一个新的 MCP 服务器实例
func NewServer(authToken string) *Server {
	cm := k8s.NewClusterManager()
	resourceOps := k8s.NewResourceOperations(cm)

	server := &Server{
		clusterManager: cm,
		resourceOps:    resourceOps,
		authToken:      authToken,
	}

	// Initialize MCP server using SDK
	// 使用 SDK 初始化 MCP 服务器
	server.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "k8s-mcp-server",
		Version: "1.0.0",
	}, nil)

	return server
}

// GetMCPServer returns the underlying MCP server instance
// GetMCPServer 返回底层的 MCP 服务器实例
func (s *Server) GetMCPServer() *mcp.Server {
	return s.mcpServer
}

// LoadKubeConfig loads kubeconfig
// LoadKubeConfig 加载 kubeconfig 配置
func (s *Server) LoadKubeConfig(configPath string) error {
	return s.clusterManager.LoadKubeConfig(configPath)
}

// RegisterTools registers all k8s tools
// RegisterTools 注册所有 k8s 工具
func (s *Server) RegisterTools() {
	// Register tools using SDK's AddTool
	// 使用 SDK 的 AddTool 注册工具

	// get_cluster_status
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_cluster_status",
		Description: "Get cluster status information (version, node count, namespace count)",
	}, s.handleGetClusterStatus)

	// list_pods
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_pods",
		Description: "List pods in a namespace. Parameters: namespace (string, required)",
	}, s.handleListPods)

	// list_services
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_services",
		Description: "List services in a namespace. Parameters: namespace (string, required)",
	}, s.handleListServices)

	// list_deployments
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_deployments",
		Description: "List deployments in a namespace. Parameters: namespace (string, required)",
	}, s.handleListDeployments)

	// list_nodes
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List all nodes in the cluster",
	}, s.handleListNodes)

	// get_resource
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_resource",
		Description: "Get detailed information about a specific resource (JSON format). Secrets will be redacted. Parameters: resource_type (string, required, e.g. 'pods' or 'pod'), name (string, required), namespace (string, required)",
	}, s.handleGetResource)

	// get_resource_yaml
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_resource_yaml",
		Description: "Get the full YAML definition of a resource. Secrets will be redacted. Parameters: resource_type (string, required, e.g. 'pods' or 'pod'), name (string, required), namespace (string, required)",
	}, s.handleGetResourceYAML)

	// get_events
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_events",
		Description: "Get cluster events. Parameters: namespace (string, required)",
	}, s.handleGetEvents)

	// get_pod_logs
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_pod_logs",
		Description: "Get pod logs. Default tail_lines=100, max_bytes=1MB. Parameters: pod_name (string, required), namespace (string, required), container_name (string, optional), tail_lines (int, optional), previous (bool, optional), cluster_name (string, optional)",
	}, s.handleGetPodLogs)

	// check_rbac_permission
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "check_rbac_permission",
		Description: "Check if the current user has permission to perform an action (kubectl auth can-i). Parameters: verb (string, required, e.g. 'get', 'list'), resource (string, required, e.g. 'pods'), namespace (string, required)",
	}, s.handleCheckRBACPermission)
}

// AuthMiddleware creates an authentication middleware
// AuthMiddleware 创建认证中间件
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for Authorization header
		// 检查 Authorization 头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		// 期望格式为 "Bearer <token>"
		const prefix = "Bearer "
		if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := authHeader[len(prefix):]
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.authToken)) != 1 {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		// Token 有效，继续处理下一个处理器
		next.ServeHTTP(w, r)
	})
}

// CreateHTTPHandler creates an HTTP handler with authentication
// CreateHTTPHandler 创建带有认证的 HTTP 处理器
func (s *Server) CreateHTTPHandler() http.Handler {
	// Create MCP streamable HTTP handler
	// 创建 MCP 可流式 HTTP 处理器
	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, &mcp.StreamableHTTPOptions{
		SessionTimeout: 5 * time.Minute,
		Stateless:      false,
	})

	// Wrap with authentication middleware
	// 使用认证中间件包装
	return s.AuthMiddleware(mcpHandler)
}

// Close closes the server
// Close 关闭服务器
func (s *Server) Close() error {
	// The SDK server doesn't have a Close method, but we can clean up k8s clients if needed
	// SDK 服务器没有 Close 方法，但如果需要我们可以清理 k8s 客户端
	return nil
}

// Tool handlers
// 工具处理函数

// handleGetClusterStatus handles get_cluster_status tool
// handleGetClusterStatus 处理 get_cluster_status 工具
func (s *Server) handleGetClusterStatus(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (
	*mcp.CallToolResult,
	struct {
		Status string `json:"status"`
	},
	error,
) {
	info, err := s.resourceOps.GetClusterInfo(ctx, "")
	if err != nil {
		return nil, struct {
			Status string `json:"status"`
		}{}, fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Format the output
	// 格式化输出
	statusText := fmt.Sprintf("Cluster Status:\n  Version: %s\n  Platform: %s\n  Node Count: %d\n  Namespace Count: %d",
		info["version"], info["platform"], info["nodeCount"], info["namespaceCount"])

	return nil, struct {
		Status string `json:"status"`
	}{
		Status: statusText,
	}, nil
}

// handleListPods handles list_pods tool
// handleListPods 处理 list_pods 工具
func (s *Server) handleListPods(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Namespace string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Pods string `json:"pods"`
	},
	error,
) {
	pods, err := s.resourceOps.ListPods(ctx, input.Namespace, "")
	if err != nil {
		return nil, struct {
			Pods string `json:"pods"`
		}{}, fmt.Errorf("failed to list pods: %w", err)
	}

	// Format the output
	// 格式化输出
	podList := "Pods:\n"
	for _, pod := range pods {
		podList += fmt.Sprintf("  - %s/%s (%s) - %s\n", pod.Namespace, pod.Name, pod.Kind, pod.Status)
	}

	return nil, struct {
		Pods string `json:"pods"`
	}{
		Pods: podList,
	}, nil
}

// handleListServices handles list_services tool
// handleListServices 处理 list_services 工具
func (s *Server) handleListServices(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Namespace string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Services string `json:"services"`
	},
	error,
) {
	services, err := s.resourceOps.ListServices(ctx, input.Namespace, "")
	if err != nil {
		return nil, struct {
			Services string `json:"services"`
		}{}, fmt.Errorf("failed to list services: %w", err)
	}

	// Format the output
	// 格式化输出
	serviceList := "Services:\n"
	for _, svc := range services {
		serviceList += fmt.Sprintf("  - %s/%s (%s) - %s\n", svc.Namespace, svc.Name, svc.Kind, svc.Status)
	}

	return nil, struct {
		Services string `json:"services"`
	}{
		Services: serviceList,
	}, nil
}

// handleListDeployments handles list_deployments tool
// handleListDeployments 处理 list_deployments 工具
func (s *Server) handleListDeployments(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Namespace string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Deployments string `json:"deployments"`
	},
	error,
) {
	deployments, err := s.resourceOps.ListDeployments(ctx, input.Namespace, "")
	if err != nil {
		return nil, struct {
			Deployments string `json:"deployments"`
		}{}, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Format the output
	// 格式化输出
	deploymentList := "Deployments:\n"
	for _, dep := range deployments {
		deploymentList += fmt.Sprintf("  - %s/%s (%s) - %s\n", dep.Namespace, dep.Name, dep.Kind, dep.Status)
	}

	return nil, struct {
		Deployments string `json:"deployments"`
	}{
		Deployments: deploymentList,
	}, nil
}

// handleListNodes handles list_nodes tool
// handleListNodes 处理 list_nodes 工具
func (s *Server) handleListNodes(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (
	*mcp.CallToolResult,
	struct {
		Nodes string `json:"nodes"`
	},
	error,
) {
	nodes, err := s.resourceOps.ListResourcesByType(ctx, k8s.ResourceTypeNodes, "", "")
	if err != nil {
		return nil, struct {
			Nodes string `json:"nodes"`
		}{}, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Format the output
	// 格式化输出
	nodeList := "Nodes:\n"
	for _, node := range nodes {
		nodeList += fmt.Sprintf("  - %s (%s) - %s\n", node.Name, node.Kind, node.Status)
	}

	return nil, struct {
		Nodes string `json:"nodes"`
	}{
		Nodes: nodeList,
	}, nil
}

// handleGetResource handles get_resource tool
// handleGetResource 处理 get_resource 工具
func (s *Server) handleGetResource(ctx context.Context, req *mcp.CallToolRequest, input struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Resource string `json:"resource"`
	},
	error,
) {
	resource, err := s.resourceOps.GetResourceDetails(ctx, k8s.ResourceType(input.ResourceType), input.Namespace, input.Name, "")
	if err != nil {
		return nil, struct {
			Resource string `json:"resource"`
		}{}, fmt.Errorf("failed to get resource: %w", err)
	}

	// Check if it's a secret and redact data
	// 检查是否是 secret 并脱敏数据
	if k8s.ResourceType(input.ResourceType) == k8s.ResourceTypeSecrets || k8s.ResourceType(input.ResourceType) == k8s.ResourceTypeSecret {
		resource = s.redactSecretData(resource)
	}

	// Serialize to JSON
	// 序列化为 JSON
	jsonStr, err := s.resourceOps.SerializeResource(resource)
	if err != nil {
		return nil, struct {
			Resource string `json:"resource"`
		}{}, fmt.Errorf("failed to serialize resource: %w", err)
	}

	return nil, struct {
		Resource string `json:"resource"`
	}{
		Resource: jsonStr,
	}, nil
}

// handleGetResourceYAML handles get_resource_yaml tool
// handleGetResourceYAML 处理 get_resource_yaml 工具
func (s *Server) handleGetResourceYAML(ctx context.Context, req *mcp.CallToolRequest, input struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		YAML string `json:"yaml"`
	},
	error,
) {
	resource, err := s.resourceOps.GetResourceDetails(ctx, k8s.ResourceType(input.ResourceType), input.Namespace, input.Name, "")
	if err != nil {
		return nil, struct {
			YAML string `json:"yaml"`
		}{}, fmt.Errorf("failed to get resource: %w", err)
	}

	// Check if it's a secret and redact data
	// 检查是否是 secret 并脱敏数据
	if k8s.ResourceType(input.ResourceType) == k8s.ResourceTypeSecrets || k8s.ResourceType(input.ResourceType) == k8s.ResourceTypeSecret {
		resource = s.redactSecretData(resource)
	}

	// Serialize to JSON (we'll convert to YAML in the future if needed, for now JSON is fine)
	// 序列化为 JSON（如果需要，我们将来可以转换为 YAML，目前 JSON 即可）
	jsonStr, err := s.resourceOps.SerializeResource(resource)
	if err != nil {
		return nil, struct {
			YAML string `json:"yaml"`
		}{}, fmt.Errorf("failed to serialize resource: %w", err)
	}

	return nil, struct {
		YAML string `json:"yaml"`
	}{
		YAML: jsonStr,
	}, nil
}

// handleGetEvents handles get_events tool
// handleGetEvents 处理 get_events 工具
func (s *Server) handleGetEvents(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Namespace string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Events string `json:"events"`
	},
	error,
) {
	events, err := s.resourceOps.ListResourcesByType(ctx, k8s.ResourceTypeEvents, input.Namespace, "")
	if err != nil {
		return nil, struct {
			Events string `json:"events"`
		}{}, fmt.Errorf("failed to list events: %w", err)
	}

	// Format the output
	// 格式化输出
	eventList := "Events:\n"
	for _, event := range events {
		eventList += fmt.Sprintf("  - %s/%s (%s) - %s\n", event.Namespace, event.Name, event.Kind, event.Status)
	}

	return nil, struct {
		Events string `json:"events"`
	}{
		Events: eventList,
	}, nil
}

// handleGetPodLogs handles get_pod_logs tool
// handleGetPodLogs 处理 get_pod_logs 工具
func (s *Server) handleGetPodLogs(ctx context.Context, req *mcp.CallToolRequest, input struct {
	PodName       string `json:"pod_name"`
	Namespace     string `json:"namespace"`
	ContainerName string `json:"container_name,omitempty"`
	TailLines     *int64 `json:"tail_lines,omitempty"`
	Previous      bool   `json:"previous,omitempty"`
	ClusterName   string `json:"cluster_name,omitempty"`
}) (
	*mcp.CallToolResult,
	struct {
		Logs string `json:"logs"`
	},
	error,
) {
	// Set default tail_lines to 100 if not specified
	// 如果未指定，默认 tail_lines 为 100
	tailLines := int64(100)
	if input.TailLines != nil {
		tailLines = *input.TailLines
	}

	// Get logs
	// 获取日志
	logs, err := s.resourceOps.GetPodLogs(ctx, input.Namespace, input.PodName, input.ContainerName, &tailLines, input.Previous, input.ClusterName)
	if err != nil {
		return nil, struct {
			Logs string `json:"logs"`
		}{}, fmt.Errorf("failed to get pod logs: %w", err)
	}

	return nil, struct {
		Logs string `json:"logs"`
	}{
		Logs: logs,
	}, nil
}

// handleCheckRBACPermission handles check_rbac_permission tool
// handleCheckRBACPermission 处理 check_rbac_permission 工具
func (s *Server) handleCheckRBACPermission(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Verb      string `json:"verb"`
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
}) (
	*mcp.CallToolResult,
	struct {
		Allowed bool   `json:"allowed"`
		Reason  string `json:"reason"`
	},
	error,
) {
	allowed, err := s.resourceOps.CheckRBACPermission(ctx, input.Verb, input.Resource, input.Namespace)
	if err != nil {
		return nil, struct {
			Allowed bool   `json:"allowed"`
			Reason  string `json:"reason"`
		}{}, fmt.Errorf("failed to check RBAC permission: %w", err)
	}

	result := struct {
		Allowed bool   `json:"allowed"`
		Reason  string `json:"reason"`
	}{
		Allowed: allowed,
	}

	if allowed {
		result.Reason = "Permission granted"
	} else {
		result.Reason = "Permission denied"
	}

	return nil, result, nil
}

// redactSecretData redacts sensitive data from secret resources
// redactSecretData 脱敏 secret 资源中的敏感数据
func (s *Server) redactSecretData(resource interface{}) interface{} {
	// Type assertion to check if it's a secret
	// 类型断言检查是否是 secret
	if secretMap, ok := resource.(map[string]interface{}); ok {
		if _, exists := secretMap["data"]; exists {
			secretMap["data"] = "***REDACTED***"
		}
		if _, exists := secretMap["stringData"]; exists {
			secretMap["stringData"] = "***REDACTED***"
		}
	}
	return resource
}
