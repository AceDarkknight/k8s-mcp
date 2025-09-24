// Package mcp implements the MCP (Model Context Protocol) server for Kubernetes management.
// 包 mcp 实现了 Kubernetes 管理的 MCP (Model Context Protocol) 服务器。
package mcp

import (
	"fmt"
	"io"
	"log"

	"k8s-mcp/internal/k8s"
)

// Server implements the MCP server
// Server 实现了 MCP 服务器
type Server struct {
	clusterManager *k8s.ClusterManager     // Kubernetes 集群管理器
	resourceOps    *k8s.ResourceOperations // 资源操作处理器
	transport      Transport               // 传输层
	dispatcher     *MessageDispatcher      // 消息分发器
}

// NewServer creates a new MCP server
// NewServer 创建一个新的 MCP 服务器
func NewServer() *Server {
	cm := k8s.NewClusterManager()
	resourceOps := k8s.NewResourceOperations(cm)

	server := &Server{
		clusterManager: cm,
		resourceOps:    resourceOps,
	}

	server.dispatcher = NewMessageDispatcher(server)
	return server
}

// SetTransport sets the transport for the server
// SetTransport 为服务器设置传输层
func (s *Server) SetTransport(transport Transport) {
	s.transport = transport
}

// LoadKubeConfig loads kubeconfig
// LoadKubeConfig 加载 kubeconfig 配置
func (s *Server) LoadKubeConfig(configPath string) error {
	return s.clusterManager.LoadKubeConfig(configPath)
}

// Run starts the MCP server
// Run 启动 MCP 服务器
func (s *Server) Run() error {
	if s.transport == nil {
		return fmt.Errorf("transport not set")
	}

	log.Println("Starting k8s MCP server...")

	for {
		request, err := s.transport.Receive()
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
				break
			}
			log.Printf("Error receiving message: %v", err)
			continue
		}

		response := s.dispatcher.Dispatch(request)
		if response != nil {
			if err := s.transport.Send(response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
		}
	}

	return nil
}

// Close closes the server
// Close 关闭服务器
func (s *Server) Close() error {
	if s.transport != nil {
		return s.transport.Close()
	}
	return nil
}

// MessageHandler implementation

// HandleInitialize handles initialization requests
// HandleInitialize 处理初始化请求
func (s *Server) HandleInitialize(req *InitializeRequest, id interface{}) (*InitializeResult, error) {
	log.Printf("Initialize request: protocol=%s, client=%s", req.ProtocolVersion, req.ClientInfo.Name)

	// Check protocol version compatibility
	if req.ProtocolVersion != ProtocolVersion {
		log.Printf("Warning: Protocol version mismatch. Client: %s, Server: %s", req.ProtocolVersion, ProtocolVersion)
	}

	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Resources: &ResourcesCapability{
				Subscribe:   false, // Not implementing subscriptions for now
				ListChanged: false,
			},
			Logging: &LoggingCapability{},
		},
		ServerInfo: Implementation{
			Name:    "k8s-mcp-server",
			Title:   "Kubernetes MCP Server",
			Version: "1.0.0",
		},
		Instructions: "Kubernetes MCP Server provides read-only access to Kubernetes cluster resources. Use tools to list clusters, switch between them, and view resources. Use resources to access cluster information and resource details.",
	}, nil
}
