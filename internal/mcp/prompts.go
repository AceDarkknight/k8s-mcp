// Package mcp implements the MCP (Model Context Protocol) server for Kubernetes management.
// 包 mcp 实现了 Kubernetes 管理的 MCP (Model Context Protocol) 服务器。
package mcp

import "fmt"

// HandleListPrompts handles prompts/list requests
// HandleListPrompts 处理提示列表请求
func (s *Server) HandleListPrompts() (*ListPromptsResult, error) {
	prompts := []Prompt{
		{
			Name:        "analyze_cluster_health",
			Title:       "Analyze Cluster Health",
			Description: "Analyze the health status of a Kubernetes cluster",
			Arguments: []PromptArgument{
				{
					Name:        "cluster_name",
					Title:       "Cluster Name",
					Description: "Name of the cluster to analyze (optional, uses current cluster if not specified)",
					Required:    false,
				},
			},
		},
		{
			Name:        "troubleshoot_pods",
			Title:       "Troubleshoot Pods",
			Description: "Help troubleshoot pod issues in a specific namespace",
			Arguments: []PromptArgument{
				{
					Name:        "namespace",
					Title:       "Namespace",
					Description: "Namespace to analyze pods in",
					Required:    true,
				},
				{
					Name:        "cluster_name",
					Title:       "Cluster Name",
					Description: "Name of the cluster (optional, uses current cluster if not specified)",
					Required:    false,
				},
			},
		},
		{
			Name:        "resource_summary",
			Title:       "Resource Summary",
			Description: "Generate a summary of resources in a cluster or namespace",
			Arguments: []PromptArgument{
				{
					Name:        "namespace",
					Title:       "Namespace",
					Description: "Namespace to summarize (optional, summarizes entire cluster if not specified)",
					Required:    false,
				},
				{
					Name:        "cluster_name",
					Title:       "Cluster Name",
					Description: "Name of the cluster (optional, uses current cluster if not specified)",
					Required:    false,
				},
			},
		},
	}

	return &ListPromptsResult{
		Prompts: prompts,
	}, nil
}

// HandleGetPrompt handles prompts/get requests
// HandleGetPrompt 处理获取提示请求
func (s *Server) HandleGetPrompt(req *GetPromptRequest) (*GetPromptResult, error) {
	switch req.Name {
	case "analyze_cluster_health":
		return s.getAnalyzeClusterHealthPrompt(req.Arguments)
	case "troubleshoot_pods":
		return s.getTroubleshootPodsPrompt(req.Arguments)
	case "resource_summary":
		return s.getResourceSummaryPrompt(req.Arguments)
	default:
		return nil, fmt.Errorf("unknown prompt: %s", req.Name)
	}
}

// getAnalyzeClusterHealthPrompt generates cluster health prompt
// getAnalyzeClusterHealthPrompt 生成集群健康提示
func (s *Server) getAnalyzeClusterHealthPrompt(args map[string]string) (*GetPromptResult, error) {
	clusterName := args["cluster_name"]
	if clusterName == "" {
		clusterName = s.clusterManager.GetCurrentCluster()
	}

	prompt := fmt.Sprintf(`Analyze the health of Kubernetes cluster "%s". Please:

1. Check the overall cluster status and version
2. Review node health and readiness
3. Examine critical system pods and their status
4. Look for any error events or warnings
5. Assess resource utilization if possible
6. Provide recommendations for any issues found

Focus on identifying potential problems and suggesting solutions.

请用中文提供你的回答。`, clusterName)

	return &GetPromptResult{
		Description: "Cluster health analysis prompt",
		Messages: []PromptMessage{
			{
				Role: "user",
				Content: TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

// getTroubleshootPodsPrompt generates pod troubleshooting prompt
// getTroubleshootPodsPrompt 生成 Pod 排查提示
func (s *Server) getTroubleshootPodsPrompt(args map[string]string) (*GetPromptResult, error) {
	namespace := args["namespace"]
	clusterName := args["cluster_name"]
	if clusterName == "" {
		clusterName = s.clusterManager.GetCurrentCluster()
	}

	prompt := fmt.Sprintf(`Help troubleshoot pod issues in namespace "%s" of cluster "%s". Please:

1. List all pods in the namespace and their current status
2. Identify any pods that are not in Running state
3. Check for any error events related to the problematic pods
4. Review resource requests and limits
5. Look for patterns in failing pods
6. Suggest specific troubleshooting steps for each issue found

Provide actionable recommendations to resolve any pod-related problems.

请用中文提供你的回答。`, namespace, clusterName)

	return &GetPromptResult{
		Description: "Pod troubleshooting prompt",
		Messages: []PromptMessage{
			{
				Role: "user",
				Content: TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

// getResourceSummaryPrompt generates resource summary prompt
// getResourceSummaryPrompt 生成资源摘要提示
func (s *Server) getResourceSummaryPrompt(args map[string]string) (*GetPromptResult, error) {
	namespace := args["namespace"]
	clusterName := args["cluster_name"]
	if clusterName == "" {
		clusterName = s.clusterManager.GetCurrentCluster()
	}

	var scope string
	if namespace != "" {
		scope = fmt.Sprintf(`namespace "%s" in cluster "%s"`, namespace, clusterName)
	} else {
		scope = fmt.Sprintf(`cluster "%s"`, clusterName)
	}

	prompt := fmt.Sprintf(`Generate a comprehensive summary of Kubernetes resources in %s. Please:

1. Provide an overview of resource counts by type (pods, services, deployments, etc.)
2. Highlight any resources with concerning status
3. Summarize resource utilization patterns
4. Identify any configuration inconsistencies
5. Note any security-related observations
6. Suggest optimizations or improvements

Create a well-organized summary that gives insight into the current state and health of the resources.

请用中文提供你的回答。`, scope)

	return &GetPromptResult{
		Description: "Resource summary analysis prompt",
		Messages: []PromptMessage{
			{
				Role: "user",
				Content: TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}
