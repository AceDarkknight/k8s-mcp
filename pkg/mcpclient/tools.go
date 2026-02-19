package mcpclient

import (
	"context"
	"encoding/json"
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

// DecodeResult 将 MCP 工具调用结果解码为指定的结构体
// DecodeResult decodes the MCP tool call result into the specified struct type
func DecodeResult[T any](result *mcp.CallToolResult) (*T, error) {
	// 检查 result 是否为 nil
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	// 检查是否有错误
	if result.IsError {
		return nil, fmt.Errorf("tool call returned error")
	}

	// 遍历 Content，寻找 TextContent 并解码
	for _, content := range result.Content {
		// 使用类型断言判断是否为 TextContent
		if textContent, ok := content.(*mcp.TextContent); ok {
			if textContent.Text == "" {
				continue
			}
			var target T
			if err := json.Unmarshal([]byte(textContent.Text), &target); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return &target, nil
		}
	}

	return nil, fmt.Errorf("no TextContent found in result")
}
