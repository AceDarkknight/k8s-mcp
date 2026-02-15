# k8s-mcp

一个用于 Kubernetes 集群管理和资源查看的模型上下文协议 (MCP) 服务器。

## 特性

- 🔗 通过 HTTP/SSE 连接到 Kubernetes 集群
- 👀 查看 Kubernetes 资源（只读）
- 🛡️ 使用 Token 认证的安全访问
- 📊 全面的集群管理工具集
- 🔒 Secret 数据脱敏以增强安全性

## 架构

- **MCP 服务器** (Golang): 通过 HTTP/SSE 提供 k8s 集群连接和资源查看功能
- **MCP 客户端** (Golang): 用于验证服务器功能的测试客户端
- **pkg/mcpclient** (Golang): 可复用的客户端库，用于在其他 Go 应用程序中集成 MCP 功能

## 快速开始

### 前置条件

- Go 1.23 或更高版本
- 已配置集群访问权限的 kubectl
- 有效的 kubeconfig 文件
- TLS 证书和密钥（用于 HTTPS 模式，默认）

### 构建

```bash
# 构建 MCP 服务器
go build -o bin/k8s-mcp-server ./cmd/server

# 构建测试客户端
go build -o bin/k8s-mcp-client ./cmd/client
```

### 运行

#### 1. 生成 TLS 证书（用于 HTTPS 模式）

```bash
# 生成用于测试的自签名证书
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout key.pem -out cert.pem -subj "/CN=localhost"
```

#### 2. 启动 MCP 服务器

```bash
# 以 HTTPS 模式启动（默认）
./bin/k8s-mcp-server --token my-secret-token --cert cert.pem --key key.pem

# 以 HTTP 模式启动（不安全）
./bin/k8s-mcp-server --token my-secret-token --insecure
```

#### 3. 使用客户端测试

```bash
# 连接到 HTTPS 服务器
./bin/k8s-mcp-client --server https://localhost:8443 --token my-secret-token

# 连接到 HTTP 服务器
./bin/k8s-mcp-client --server http://localhost:8443 --token my-secret-token --insecure-skip-verify
```

## 配置

服务器支持通过命令行标志进行配置：

### 服务器标志

- `--port`: 监听端口（默认：8443）
- `--cert`: TLS 证书文件路径（HTTPS 模式必需）
- `--key`: TLS 密钥文件路径（HTTPS 模式必需）
- `--insecure`: 以不安全的 HTTP 模式运行（默认为 HTTPS）
- `--token`: 认证 Token（必需）
- `--kubeconfig`: kubeconfig 文件路径（可选，未指定则使用默认值）

### 日志配置

服务器提供基于 Uber Zap 和 Lumberjack 的全面日志系统。

| 标志 | 环境变量 | 默认值 | 描述 |
|-------|---------------------|---------|-------------|
| `--log-level` | | info | 日志级别 (debug, info, warn, error) |
| `--log-format` | | text | 日志格式 (json, text) |
| `--log-to-file` | | false (Server: true) | 是否启用日志文件输出（Server 端默认为 true） |
| `--log-file` | | logs/app.log | 日志文件路径 |
| `--log-max-size` | | 100 | 单个日志文件最大大小 (MB) |
| `--log-max-backups`| | 3 | 保留的旧日志文件最大数量 |
| `--log-max-age` | | 30 | 保留旧日志文件的最大天数 |
| `--log-compress` | | true | 是否压缩旧日志文件 |
| `--log-caller` | | true | 是否记录调用者信息（文件名和行号） |
| `--log-stacktrace` | | false | 是否在错误级别记录堆栈信息 |

当启用 `--log-to-file` 时，日志将同时输出到控制台和指定的日志文件。日志系统会自动根据大小、日期和备份数量处理日志轮转。

### 客户端标志

- `--server`: MCP 服务器 URL（默认：https://localhost:8443）
- `--token`: 认证 Token（必需）
- `--insecure-skip-verify`: 跳过 TLS 证书验证（用于自签名证书）

## MCP 工具

有关每个工具的详细 API 文档，包括函数签名、参数说明和示例代码，请参阅 [API 文档](docs/api.md)。

服务器提供以下工具：

### 集群管理

- `get_cluster_status`: 获取集群状态信息（版本、节点数、命名空间数）
- `list_nodes`: 列出集群中的所有节点
- `list_namespaces`: 列出集群中的所有命名空间

### 资源管理

- `list_pods`: 列出命名空间中的 Pod
- `list_services`: 列出命名空间中的 Service
- `list_deployments`: 列出命名空间中的 Deployment

- `get_resource`: 获取特定资源的详细信息（JSON 格式）。Secret 将被脱敏。
- `get_resource_yaml`: 获取资源的完整 YAML 定义。Secret 将被脱敏。

### 可观测性和调试

- `get_events`: 获取集群事件
- `get_pod_logs`: 获取 Pod 日志。默认 tail_lines=100，最大 1MB

### 安全

- `check_rbac_permission`: 检查当前用户是否有权限执行某个操作（kubectl auth can-i）

## 安全性

- 默认情况下，所有操作都是只读的
- 所有连接都需要基于 Token 的认证
- 检索 Secret 数据时会自动脱敏
- 支持 RBAC 权限验证
- 安全的 kubeconfig 处理
- 连接超时和重试机制

## 集成

### 作为库使用

`pkg/mcpclient` 包可以导入到其他 Go 应用程序中使用：

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

    // 创建客户端，支持自定义头
    client, err := mcpclient.NewClient(config,
        mcpclient.WithHeader("X-Custom-Header", "value"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 连接到服务器
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
        fmt.Printf("工具: %s\n", tool.Name)
    }

    // 调用工具
    result, err := client.CallTool(ctx, "get_cluster_status", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("结果: %v\n", result)
}
```

更多详情请参阅 [`pkg/mcpclient/README.md`](pkg/mcpclient/README.md)。

### MCP 协议集成

k8s-mcp 遵循标准的 MCP 协议，可以集成到任何兼容 MCP 的应用程序中：

1. **Claude Desktop**: AI 助手可以查看和分析 Kubernetes 资源
2. **VS Code**: 通过 MCP 扩展获取 Kubernetes 上下文
3. **自定义应用程序**: 使用 MCP 客户端库进行集成

## 开发

要添加新功能：

1. **新工具**: 在 `internal/mcp/server.go` 中添加新的工具定义和处理函数
2. **新资源**: 在 `internal/k8s/resources.go` 中添加新的资源类型

项目采用模块化设计，便于扩展和维护。