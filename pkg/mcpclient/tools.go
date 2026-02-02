package mcpclient

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListTools 获取工具列表
// ListTools retrieves the list of available tools
func (c *Client) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	if c.session == nil {
		return nil, fmt.Errorf("client not connected")
	}

	result, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return result.Tools, nil
}

// CallTool 调用工具
// CallTool calls a specific tool with arguments
func (c *Client) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	if c.session == nil {
		return nil, fmt.Errorf("client not connected")
	}

	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	return result, nil
}
