# 2026-01-26 客户端重构计划

## 1. 目标

将 `cmd/client` 中的 MCP 客户端初始化和交互逻辑提取到独立的包 `pkg/mcpclient` 中，以便其他 Go 程序可以复用该客户端逻辑连接到 MCP 服务器。

## 2. 分析

当前 `cmd/client` 的实现紧耦合在 `cobra` 的 `Run` 函数中 (`cmd/client/cmd/root.go`)。
主要逻辑包括：
1.  **配置读取**：从 Viper/Flags 读取 Server URL, Token, InsecureSkipVerify。
2.  **HTTP 客户端创建**：配置 TLS 和自定义的 `tokenAuthTransport` 用于注入 Bearer Token。
3.  **MCP 连接**：使用 `mcp.NewClient` 和 `mcp.StreamableClientTransport` 建立连接。
4.  **交互逻辑**：简单的 REPL (Read-Eval-Print Loop) 处理 `help`, `tools`, `call` 命令。

重构的目标是将 1, 2, 3 封装到 `pkg/mcpclient`，而 `cmd/client` 仅保留命令行参数解析和 REPL 交互逻辑。

## 3. 设计方案

### 3.1 目录结构变更

```text
k8s-mcp/
├── cmd/
│   └── client/
│       └── cmd/
│           └── root.go      # 更新：使用 pkg/mcpclient
├── pkg/
│   └── mcpclient/           # 新增：独立客户端包
│       ├── client.go        # 核心客户端逻辑 (Connect, Close)
│       ├── config.go        # 配置定义
│       ├── tools.go         # 工具调用封装 (ListTools, CallTool)
│       └── transport.go     # HTTP 传输层封装 (Auth)
```

### 3.2 接口设计 (pkg/mcpclient)

**Config 结构体**:
```go
type Config struct {
    ServerURL          string
    AuthToken          string
    InsecureSkipVerify bool
    UserAgent          string // 可选：标识客户端身份
}
```

**Client 结构体**:
```go
type Client struct {
    config  Config
    session *mcp.ClientSession
    mcpClient *mcp.Client
}
```

**导出方法**:
```go
// LoadConfig 从环境变量或配置文件加载配置
func LoadConfig() (Config, error)

// Option 定义配置选项函数
type Option func(*Client)

// WithHeader 添加自定义 HTTP 头
func WithHeader(key, value string) Option

// NewClient 创建客户端实例，支持通过 Option 自定义配置
func NewClient(config Config, opts ...Option) (*Client, error)

// Connect 建立连接
func (c *Client) Connect(ctx context.Context) error

// Close 关闭连接
func (c *Client) Close() error

// ListTools 获取工具列表
func (c *Client) ListTools(ctx context.Context) ([]mcp.Tool, error)

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error)
```

## 4. 实施步骤

1.  **创建包目录**: 创建 `pkg/mcpclient` 目录。
2.  **实现配置与传输层**:
    *   在 `pkg/mcpclient/config.go` 中定义 `Config` 结构体和 `LoadConfig` 函数。
    *   在 `pkg/mcpclient/options.go` (新建) 中定义 `Option` 类型及 `WithHeader` 等选项函数。
    *   将 `cmd/client/cmd/root.go` 中的 `tokenAuthTransport` 移动到 `pkg/mcpclient/transport.go`。
3.  **实现核心客户端**:
    *   在 `pkg/mcpclient/client.go` 中实现 `NewClient` (支持 Option 可变参数), `Connect`, `Close`。
    *   封装 `http.Client` 创建和 `mcp` SDK 的初始化逻辑。
4.  **实现功能方法**:
    *   在 `pkg/mcpclient/tools.go` 中实现 `ListTools` 和 `CallTool`。
5.  **重构命令行入口**:
    *   修改 `cmd/client/cmd/root.go`。
    *   移除原本直接的 HTTP/MCP 初始化代码。
    *   使用 `mcpclient.Config` 接收 Viper 配置。
    *   实例化 `mcpclient.Client` 并调用其方法执行操作。
6.  **更新文档**:
    *   更新 `README.md` 和 `README_zh.md`，反映项目结构的变更（如新增的 `pkg/mcpclient`）。
    *   添加关于如何作为库使用 `pkg/mcpclient` 的简要说明或示例。
7.  **验证**:
    *   编写单元测试 (如果可行)。
    *   手动运行客户端连接本地 Server 进行验证。

## 5. 验证步骤

1.  **编译**: 运行 `go build ./cmd/client` 确保编译通过。
2.  **启动服务端**: 启动 `k8s-mcp` 服务端。
3.  **运行客户端**:
    ```bash
    ./client.exe --server https://localhost:8443 --token <your-token> --insecure-skip-verify
    ```
4.  **测试交互**:
    *   输入 `tools`：应列出可用工具。
    *   输入 `call <tool> ...`：应成功执行工具。

## 6. 注意事项

*   保持 `cmd/client` 的用户体验不变（命令行参数、交互命令）。
*   确保 `pkg/mcpclient` 对 `cobra` 或 `viper` 无依赖，保持纯净的 Go 库风格。
