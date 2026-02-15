# k8s-mcp 使用指南

这个文档详细介绍如何使用 k8s-mcp (Kubernetes Model Context Protocol) 服务器。

## 安装和设置

### 1. 构建项目

```bash
# 构建服务器
go build -o bin/k8s-mcp-server ./cmd/server

# 构建测试客户端
go build -o bin/k8s-mcp-client ./cmd/client
```

### 2. 配置 Kubernetes 访问

确保您有有效的 kubeconfig 文件：

```bash
# 检查当前 kubeconfig
kubectl config view

# 或设置自定义 kubeconfig 路径
export KUBECONFIG=/path/to/your/kubeconfig
```

### 3. 生成 TLS 证书（用于 HTTPS 模式）

```bash
# 生成自签名证书用于测试
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout key.pem -out cert.pem -subj "/CN=localhost"
```

## 使用方式

### 方式 1：作为 MCP 服务器使用

k8s-mcp-server 设计为 MCP 服务器，可以被其他 MCP 客户端（如 Claude Desktop、VS Code 等）使用。

#### 配置 Claude Desktop

1. 打开 Claude Desktop 配置文件：
    - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
    - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

2. 添加 k8s-mcp 服务器配置：

```json
{
  "mcpServers": {
    "k8s-mcp": {
      "command": "E:\\code\\k8s-mcp\\bin\\k8s-mcp-server.exe",
      "args": []
    }
  }
}
```

3. 重启 Claude Desktop

#### 配置其他 MCP 客户端

对于支持 MCP 的其他客户端，按照相似的方式配置：
- **command**: k8s-mcp-server.exe 的完整路径
- **args**: 可选的命令行参数（如 kubeconfig 路径、token、证书路径）

### 方式 2：使用测试客户端

我们提供了一个简单的测试客户端来验证功能：

```bash
# 启动服务器（HTTPS 模式）
# 使用命令行参数
./bin/k8s-mcp-server --token my-secret-token --cert cert.pem --key key.pem

# 使用环境变量
export MCP_TOKEN=my-secret-token
export MCP_CERT=cert.pem
export MCP_KEY=key.pem
./bin/k8s-mcp-server
```

```bash
# 启动测试客户端
# 使用命令行参数
./bin/k8s-mcp-client --server https://localhost:8443 --token my-secret-token

# 使用环境变量
export MCP_CLIENT_SERVER=https://localhost:8443
export MCP_CLIENT_TOKEN=my-secret-token
./bin/k8s-mcp-client
```

## 功能详解

### 1. Tools（工具）

有关每个工具的详细 API 文档，请参阅 [API 文档](docs/api.md)。

k8s-mcp 提供以下工具，AI 可以自动调用：

#### `get_cluster_status`
获取集群状态信息（版本、节点数、命名空间数）。
```
参数：无
示例：call get_cluster_status
```

#### `list_pods`
列出指定命名空间中的 Pod。
```
参数：
- namespace (string, 必需): 命名空间
示例：call list_pods namespace=default
```

#### `list_services`
列出指定命名空间中的 Service。
```
参数：
- namespace (string, 必需): 命名空间
示例：call list_services namespace=default
```

#### `list_deployments`
列出指定命名空间中的 Deployment。
```
参数：
- namespace (string, 必需): 命名空间
示例：call list_deployments namespace=default
```

#### `list_nodes`
列出集群中的所有节点。
```
参数：无
示例：call list_nodes
```

#### `list_namespaces`
列出集群中的所有命名空间。
```
参数：无
示例：call list_namespaces
```

#### `get_resource`
获取特定资源的详细信息（JSON 格式）。Secret 数据将被脱敏。
```
参数：
- resource_type (string, 必需): 资源类型
- name (string, 必需): 资源名称
- namespace (string, 必需): 命名空间
示例：call get_resource resource_type=pods name=my-pod namespace=default
```

#### `get_resource_yaml`
获取资源的完整 YAML 定义。Secret 数据将被脱敏。
```
参数：
- resource_type (string, 必需): 资源类型
- name (string, 必需): 资源名称
- namespace (string, 必需): 命名空间
示例：call get_resource_yaml resource_type=pods name=my-pod namespace=default
```

#### `get_events`
获取集群事件。
```
参数：
- namespace (string, 必需): 命名空间
示例：call get_events namespace=default
```

#### `get_pod_logs`
获取 Pod 日志。默认 tail_lines=100，最大 1MB。
```
参数：
- pod_name (string, 必需): Pod 名称
- namespace (string, 必需): 命名空间
- container_name (string, 可选): 容器名称
- tail_lines (int64, 可选): 尾部行数
- previous (bool, 可选): 是否查看前一个容器的日志
示例：call get_pod_logs pod_name=my-pod namespace=default tail_lines=50
```

#### `check_rbac_permission`
检查当前用户是否有权限执行某个操作（kubectl auth can-i）。
```
参数：
- verb (string, 必需): 操作动词（如 get, list, create）
- resource (string, 必需): 资源类型（如 pods, services）
- namespace (string, 必需): 命名空间
示例：call check_rbac_permission verb=get resource=pods namespace=default
```

## 测试客户端命令

在测试客户端中，您可以使用以下命令：

### 基本命令
- `help` - 显示帮助信息
- `tools` - 列出所有可用工具
- `quit` / `exit` - 退出客户端

### 调用工具
```bash
call <tool_name> [key=value ...]

# 示例
call get_cluster_status
call list_pods namespace=default
call get_events namespace=default
call get_pod_logs pod_name=my-pod namespace=default
call check_rbac_permission verb=get resource=pods namespace=default
```

## 故障排除

### 常见问题

1. **连接集群失败**
   - 检查 kubeconfig 文件是否正确
   - 确认网络连接到 Kubernetes 集群
   - 验证认证信息是否有效

2. **权限错误**
   - 确认用户有足够的 RBAC 权限
   - 检查 ServiceAccount 权限设置

3. **服务器启动失败**
   - 检查 Go 版本是否兼容
   - 确认所有依赖已正确安装
   - 检查 TLS 证书和密钥路径是否正确

4. **客户端连接失败**
   - 检查 Server URL 是否正确
   - 检查 Token 是否正确
   - 如果使用自签名证书，确保客户端使用 `--insecure-skip-verify`

### 调试和日志

服务器提供了详细的日志输出，有助于排查连接和认证问题。

```bash
# 启动服务器并启用文件日志
./bin/k8s-mcp-server --token my-secret-token --log-to-file --log-level debug --log-file logs/server.log
```

### 命令行参数和环境变量

#### 服务器参数

| 参数 | 环境变量 | 默认值 | 说明 |
|-------|---------|---------|------|
| `--port` | `MCP_PORT` | 8443 | 监听端口 |
| `--cert` | `MCP_CERT` | | TLS 证书文件路径（HTTPS 模式必需） |
| `--key` | `MCP_KEY` | | TLS 密钥文件路径（HTTPS 模式必需） |
| `--insecure` | `MCP_INSECURE` | false | 使用不安全的 HTTP 模式（默认为 HTTPS） |
| `--token` | `MCP_TOKEN` | | 认证 Token（必需） |
| `--kubeconfig` | `MCP_KUBECONFIG` | | kubeconfig 文件路径（可选） |
| `--log-level` | | info | 日志级别 (debug, info, warn, error) |
| `--log-format` | | text | 日志格式 (json, text) |
| `--log-to-file` | | false | 是否启用日志文件输出 |
| `--log-file` | | logs/app.log | 日志文件路径 |
| `--log-max-size` | | 100 | 单个日志文件最大大小 (MB) |
| `--log-max-backups`| | 3 | 保留的旧日志文件最大数量 |
| `--log-max-age` | | 30 | 保留旧日志文件的最大天数 |
| `--log-compress` | | true | 是否压缩旧日志文件 |
| `--log-caller` | | true | 是否记录调用者信息 |
| `--log-stacktrace` | | false | 是否在错误级别记录堆栈信息 |

#### 客户端参数

| 参数 | 环境变量 | 默认值 | 说明 |
|-------|---------|---------|------|
| `--server` | `MCP_CLIENT_SERVER` | https://localhost:8443 | MCP 服务器 URL |
| `--token` | `MCP_CLIENT_TOKEN` | | 认证 Token（必需） |
| `--insecure-skip-verify` | `MCP_CLIENT_INSECURE_SKIP_VERIFY` | false | 跳过 TLS 证书验证（用于自签名证书） |

**注意**: 命令行参数的优先级高于环境变量。

## 安全注意事项

- 所有操作都是只读的，不会修改集群资源
- 服务器仅提供查看权限，符合 MCP 安全最佳实践
- Token 认证是所有连接所必需的
- Secret 数据在获取时自动脱敏
- 建议在生产环境中使用具有最小权限的 ServiceAccount
- kubeconfig 文件应妥善保管，包含敏感的认证信息

## 集成到其他应用

k8s-mcp 遵循标准的 MCP 协议，可以集成到任何支持 MCP 的应用中：

1. **Claude Desktop**: AI 助手可以查看和分析 Kubernetes 资源
2. **VS Code**: 通过 MCP 扩展获得 Kubernetes 上下文
3. **自定义应用**: 使用 MCP 客户端库集成

## 扩展和自定义

要添加新功能：

1. **新工具**: 在 `internal/mcp/server.go` 中添加新的工具定义和处理函数
2. **新资源**: 在 `internal/k8s/resources.go` 中添加新的资源类型

项目采用模块化设计，便于扩展和维护。
