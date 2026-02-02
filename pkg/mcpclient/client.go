package mcpclient

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client MCP 客户端封装
// Client wraps the MCP client
type Client struct {
	config        Config
	customHeaders map[string]string
	mcpClient     *mcp.Client
	session       *mcp.ClientSession
}

// NewClient 创建客户端实例，支持通过 Option 自定义配置
// NewClient creates a client instance with optional customization via Option
func NewClient(config Config, opts ...Option) (*Client, error) {
	// 验证必需参数
	// Validate required parameters
	if config.AuthToken == "" {
		return nil, fmt.Errorf("AuthToken is required")
	}

	client := &Client{
		config:        config,
		customHeaders: make(map[string]string),
	}

	// 应用可选配置
	// Apply optional configurations
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Connect 建立连接
// Connect establishes a connection to the MCP server
func (c *Client) Connect(ctx context.Context) error {
	// 创建 HTTP 客户端和传输层
	// Create HTTP client and transport
	httpClient := createHTTPClient(c.config, c.customHeaders)

	// 创建 MCP 客户端
	// Create MCP client
	c.mcpClient = mcp.NewClient(&mcp.Implementation{
		Name:    c.config.UserAgent,
		Version: "1.0.0",
	}, nil)

	// 创建可流式传输
	// Create streamable transport
	transport := &mcp.StreamableClientTransport{
		Endpoint:   c.config.ServerURL,
		HTTPClient: httpClient,
	}

	// 连接到服务器
	// Connect to server
	session, err := c.mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.session = session
	return nil
}

// Close 关闭连接
// Close closes the connection to the MCP server
func (c *Client) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
