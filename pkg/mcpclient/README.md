# mcpclient

MCP 客户端封装包，提供简化的 API 用于连接和与 MCP 服务器交互。

## 功能

- 支持通过配置文件或参数初始化客户端
- 支持 Token 认证
- 支持 TLS 证书验证配置
- 支持自定义 HTTP 头
- 封装了 MCP 基础方法（ListTools, CallTool）

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AceDarkknight/k8s-mcp/pkg/mcpclient"
)

func main() {
    // 创建配置
    config := mcpclient.Config{
        ServerURL:          "https://localhost:8443",
        AuthToken:          "your-token",
        InsecureSkipVerify: true,
    }

    // 创建客户端
    client, err := mcpclient.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 连接服务器
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // 列出工具
    tools, err := client.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, tool := range tools {
        fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
    }
}
```

### 使用环境变量

```go
config, err := mcpclient.LoadConfig()
if err != nil {
    log.Fatal(err)
}

client, err := mcpclient.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

### 自定义配置

```go
client, err := mcpclient.NewClient(config,
    mcpclient.WithHeader("X-Custom-Header", "value"),
    mcpclient.WithUserAgent("my-app/1.0.0"),
)
```

## API 参考

### Config

配置结构体，包含以下字段：

- `ServerURL` (string): MCP 服务器地址
- `AuthToken` (string): 认证 Token（必需）
- `InsecureSkipVerify` (bool): 是否跳过 TLS 证书验证
- `UserAgent` (string): 客户端标识

### Client

客户端结构体，提供以下方法：

- `NewClient(config Config, opts ...Option) (*Client, error)`: 创建客户端实例
- `Connect(ctx context.Context) error`: 建立连接
- `Close() error`: 关闭连接
- `ListTools(ctx context.Context) ([]*mcp.Tool, error)`: 获取工具列表
- `CallTool(ctx context.Context, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error)`: 调用工具

### Options

可选配置函数：

- `WithHeader(key, value string) Option`: 添加自定义 HTTP 头
- `WithUserAgent(userAgent string) Option`: 设置自定义 User-Agent

## 环境变量

支持以下环境变量：

- `MCP_CLIENT_SERVER`: MCP 服务器地址（默认: https://localhost:8443）
- `MCP_CLIENT_TOKEN`: 认证 Token（必需）
- `MCP_CLIENT_INSECURE_SKIP_VERIFY`: 是否跳过 TLS 证书验证（默认: false）
- `MCP_CLIENT_USER_AGENT`: 客户端标识（默认: k8s-mcp-client/1.0.0）
