// Package mcp implements the MCP (Model Context Protocol) server for Kubernetes management.
// 包 mcp 实现了 Kubernetes 管理的 MCP (Model Context Protocol) 服务器。
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// HandleListResources handles resources/list requests
// HandleListResources 处理资源列表请求
func (s *Server) HandleListResources() (*ListResourcesResult, error) {
	clusters := s.clusterManager.GetClusters()
	current := s.clusterManager.GetCurrentCluster()

	var resources []Resource

	// Add cluster list resource
	// 添加集群列表资源
	resources = append(resources, Resource{
		URI:         "k8s://clusters",
		Name:        "clusters",
		Title:       "Kubernetes Clusters",                   // Kubernetes 集群
		Description: "List of available Kubernetes clusters", // 可用 Kubernetes 集群列表
		MimeType:    "application/json",
	})

	// Add cluster-specific resources
	// 添加集群特定资源
	for _, cluster := range clusters {
		// Cluster info resource
		// 集群信息资源
		resources = append(resources, Resource{
			URI:         fmt.Sprintf("k8s://cluster/%s/info", cluster),
			Name:        fmt.Sprintf("cluster-%s-info", cluster),
			Title:       fmt.Sprintf("Cluster %s Information", cluster),             // 集群 %s 信息
			Description: fmt.Sprintf("Basic information about cluster %s", cluster), // 关于集群 %s 的基本信息
			MimeType:    "application/json",
		})

		// Namespaces resource
		// 命名空间资源
		resources = append(resources, Resource{
			URI:         fmt.Sprintf("k8s://cluster/%s/namespaces", cluster),
			Name:        fmt.Sprintf("cluster-%s-namespaces", cluster),
			Title:       fmt.Sprintf("Cluster %s Namespaces", cluster),            // 集群 %s 命名空间
			Description: fmt.Sprintf("List of namespaces in cluster %s", cluster), // 集群 %s 中的命名空间列表
			MimeType:    "application/json",
		})

		// Mark current cluster
		// 标记当前集群
		if cluster == current {
			resources[len(resources)-2].Title += " (Current)" // (当前)
			resources[len(resources)-1].Title += " (Current)" // (当前)
		}
	}

	return &ListResourcesResult{
		Resources: resources,
	}, nil
}

// HandleReadResource handles resources/read requests
// HandleReadResource 处理资源读取请求
func (s *Server) HandleReadResource(req *ReadResourceRequest) (*ReadResourceResult, error) {
	ctx := context.Background()

	// Parse URI to determine what to return
	if req.URI == "k8s://clusters" {
		return s.readClustersResource(ctx)
	}

	// Parse cluster-specific URIs
	if strings.HasPrefix(req.URI, "k8s://cluster/") {
		return s.readClusterResource(ctx, req.URI)
	}

	return nil, fmt.Errorf("unsupported resource URI: %s", req.URI)
}

// readClustersResource reads clusters resource
// readClustersResource 读取集群资源
func (s *Server) readClustersResource(ctx context.Context) (*ReadResourceResult, error) {
	clusters := s.clusterManager.GetClusters()
	current := s.clusterManager.GetCurrentCluster()

	clusterData := map[string]interface{}{
		"clusters": clusters,
		"current":  current,
		"count":    len(clusters),
	}

	jsonStr, err := json.MarshalIndent(clusterData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize clusters data: %w", err)
	}

	return &ReadResourceResult{
		Contents: []ResourceContents{
			{
				URI:      "k8s://clusters",
				Name:     "clusters",
				MimeType: "application/json",
				Text:     string(jsonStr),
			},
		},
	}, nil
}

// readClusterResource reads cluster-specific resource
// readClusterResource 读取集群特定资源
func (s *Server) readClusterResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	// Parse URI: k8s://cluster/{cluster-name}/{resource-type}
	parts := strings.Split(strings.TrimPrefix(uri, "k8s://cluster/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid cluster resource URI format")
	}

	clusterName := parts[0]
	resourceType := parts[1]

	switch resourceType {
	case "info":
		return s.readClusterInfo(ctx, clusterName, uri)
	case "namespaces":
		return s.readClusterNamespaces(ctx, clusterName, uri)
	default:
		return nil, fmt.Errorf("unsupported cluster resource type: %s", resourceType)
	}
}

// readClusterInfo reads cluster info
// readClusterInfo 读取集群信息
func (s *Server) readClusterInfo(ctx context.Context, clusterName, uri string) (*ReadResourceResult, error) {
	info, err := s.resourceOps.GetClusterInfo(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	jsonStr, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize cluster info: %w", err)
	}

	return &ReadResourceResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				Name:     fmt.Sprintf("cluster-%s-info", clusterName),
				MimeType: "application/json",
				Text:     string(jsonStr),
			},
		},
	}, nil
}

// readClusterNamespaces reads cluster namespaces
// readClusterNamespaces 读取集群命名空间
func (s *Server) readClusterNamespaces(ctx context.Context, clusterName, uri string) (*ReadResourceResult, error) {
	namespaces, err := s.resourceOps.ListNamespaces(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	jsonStr, err := json.MarshalIndent(namespaces, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize namespaces: %w", err)
	}

	return &ReadResourceResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				Name:     fmt.Sprintf("cluster-%s-namespaces", clusterName),
				MimeType: "application/json",
				Text:     string(jsonStr),
			},
		},
	}, nil
}
