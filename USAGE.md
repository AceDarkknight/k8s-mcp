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

## 使用方式

### 方式 1：作为 MCP 服务器使用

k8s-mcp-server 主要设计为 MCP 服务器，可以被其他 MCP 客户端（如 Claude Desktop、VS Code 等）使用。

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
      "args": ["-kubeconfig", "C:\\Users\\your-username\\.kube\\config"]
    }
  }
}
```

3. 重启 Claude Desktop

#### 配置其他 MCP 客户端

对于支持 MCP 的其他客户端，按照相似的方式配置：
- **command**: k8s-mcp-server.exe 的完整路径
- **args**: 可选的命令行参数（如 kubeconfig 路径）

### 方式 2：使用测试客户端

我们提供了一个简单的测试客户端来验证功能：

```bash
# 启动测试客户端
./bin/k8s-mcp-client.exe ./bin/k8s-mcp-server.exe

# 或使用测试脚本
./scripts/test.bat
```

## 功能详解

### 1. Tools（工具）

k8s-mcp 提供以下工具，AI 可以自动调用：

#### `list_clusters`
列出所有可用的 Kubernetes 集群
```
参数：无
示例：call list_clusters
```

#### `switch_cluster`
切换到指定的集群
```
参数：
- cluster_name (string, 必需): 要切换到的集群名称

示例：call switch_cluster cluster_name=my-cluster
```

#### `get_current_cluster`
获取当前活动集群的名称
```
参数：无
示例：call get_current_cluster
```

#### `list_namespaces`
列出指定集群中的所有命名空间
```
参数：
- cluster_name (string, 可选): 集群名称，不指定则使用当前集群

示例：call list_namespaces
示例：call list_namespaces cluster_name=my-cluster
```

#### `list_resources`
列出指定类型的 Kubernetes 资源
```
参数：
- resource_type (string, 必需): 资源类型
  支持：pods, services, deployments, configmaps, secrets, namespaces, nodes, events
- namespace (string, 可选): 命名空间（对于集群级资源可省略）
- cluster_name (string, 可选): 集群名称

示例：call list_resources resource_type=pods
示例：call list_resources resource_type=pods namespace=default
示例：call list_resources resource_type=pods namespace=kube-system cluster_name=my-cluster
```

#### `get_resource`
获取特定资源的详细信息
```
参数：
- resource_type (string, 必需): 资源类型
- name (string, 必需): 资源名称
- namespace (string, 可选): 命名空间
- cluster_name (string, 可选): 集群名称

示例：call get_resource resource_type=pods name=my-pod namespace=default
```

#### `describe_resource`
获取资源的详细 JSON 描述
```
参数：同 get_resource

示例：call describe_resource resource_type=pods name=my-pod namespace=default
```

### 2. Resources（资源）

k8s-mcp 提供以下资源，应用程序可以读取作为上下文：

#### `k8s://clusters`
包含所有可用集群的列表信息

#### `k8s://cluster/{cluster-name}/info`
指定集群的基本信息（版本、节点数量等）

#### `k8s://cluster/{cluster-name}/namespaces`
指定集群中的命名空间列表

### 3. Prompts（提示模板）

k8s-mcp 提供以下提示模板：

#### `analyze_cluster_health`
分析集群健康状况的提示模板
```
参数：
- cluster_name (string, 可选): 要分析的集群名称

使用：prompt analyze_cluster_health cluster_name=my-cluster
```

#### `troubleshoot_pods`
Pod 故障排查提示模板
```
参数：
- namespace (string, 必需): 要分析的命名空间
- cluster_name (string, 可选): 集群名称

使用：prompt troubleshoot_pods namespace=default
```

#### `resource_summary`
资源摘要分析提示模板
```
参数：
- namespace (string, 可选): 命名空间，不指定则分析整个集群
- cluster_name (string, 可选): 集群名称

使用：prompt resource_summary
使用：prompt resource_summary namespace=kube-system
```

## 测试客户端命令

在测试客户端中，您可以使用以下命令：

### 基本命令
- `help` - 显示帮助信息
- `tools` - 列出所有可用工具
- `resources` - 列出所有可用资源
- `prompts` - 列出所有可用提示模板
- `quit` / `exit` - 退出客户端

### 调用工具
```bash
call <tool_name> [key=value ...]

# 示例
call list_clusters
call switch_cluster cluster_name=my-cluster
call list_resources resource_type=pods namespace=default
call get_resource resource_type=pods name=my-pod namespace=default
```

### 读取资源
```bash
read <resource_uri>

# 示例
read k8s://clusters
read k8s://cluster/my-cluster/info
read k8s://cluster/my-cluster/namespaces
```

### 获取提示
```bash
prompt <prompt_name> [key=value ...]

# 示例
prompt analyze_cluster_health cluster_name=my-cluster
prompt troubleshoot_pods namespace=default
prompt resource_summary namespace=kube-system
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

### 调试模式

服务器会将调试信息输出到 stderr，在测试客户端中可以看到这些日志。

### 命令行参数

服务器支持以下命令行参数：
- `-kubeconfig`: 指定 kubeconfig 文件路径

```bash
./bin/k8s-mcp-server.exe -kubeconfig /path/to/kubeconfig
```

## 安全注意事项

- 所有操作都是只读的，不会修改集群资源
- 服务器仅提供查看权限，符合 MCP 安全最佳实践
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
2. **新资源**: 在资源处理函数中添加新的 URI 模式
3. **新提示**: 在提示处理函数中添加新的模板

项目采用模块化设计，便于扩展和维护。
