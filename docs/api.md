# API 接口文档

本文档详细介绍了 k8s-mcp 服务器提供的所有工具接口。每个工具都遵循 MCP (Model Context Protocol) 规范。

## 目录

- [集群管理](#集群管理)
    - [get_cluster_status](#get_cluster_status)
    - [list_nodes](#list_nodes)
    - [list_namespaces](#list_namespaces)
- [资源管理](#资源管理)
    - [list_pods](#list_pods)
    - [list_services](#list_services)
    - [list_deployments](#list_deployments)
    - [get_resource](#get_resource)
    - [get_resource_yaml](#get_resource_yaml)
- [可观测性与调试](#可观测性与调试)
    - [get_events](#get_events)
    - [get_pod_logs](#get_pod_logs)
- [安全](#安全)
    - [check_rbac_permission](#check_rbac_permission)

---

## 集群管理

### get_cluster_status

获取集群状态信息，包括 Kubernetes 版本、节点数量和命名空间数量。

- **函数签名**: `handleGetClusterStatus`
- **描述**: Get cluster status information (version, node count, namespace count)

#### 参数

无

#### 返回值

返回 `ClusterStatusResult` 对象，包含格式化的状态文本。

```json
{
  "status": "Cluster Status:\n  Version: v1.28.0\n  Platform: linux/amd64\n  Node Count: 3\n  Namespace Count: 10"
}
```

#### 示例代码 (MCP Client)

```go
result, err := client.CallTool(ctx, "get_cluster_status", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Content[0].Text)
```

### list_nodes

列出集群中的所有节点及其状态。

- **函数签名**: `handleListNodes`
- **描述**: List all nodes in the cluster

#### 参数

无

#### 返回值

返回 `NodesResult` 对象，包含格式化的节点列表。

```json
{
  "nodes": "Nodes:\n  - node-1 (Node) - Ready\n  - node-2 (Node) - Ready"
}
```

#### 示例代码 (MCP Client)

```go
result, err := client.CallTool(ctx, "list_nodes", nil)
```

### list_namespaces

列出集群中的所有命名空间。

- **函数签名**: `handleListNamespaces`
- **描述**: List all namespaces in the cluster

#### 参数

无

#### 返回值

返回 `NamespacesResult` 对象，包含格式化的命名空间列表。

```json
{
  "namespaces": "Namespaces:\n  - default (Namespace) - Active\n  - kube-system (Namespace) - Active"
}
```

#### 示例代码 (MCP Client)

```go
result, err := client.CallTool(ctx, "list_namespaces", nil)
```

---

## 资源管理

### list_pods

列出指定命名空间中的 Pod。

- **函数签名**: `handleListPods`
- **描述**: List pods in a namespace

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `PodsResult` 对象。

```json
{
  "pods": "Pods:\n  - default/nginx-pod (Pod) - Running"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "namespace": "default",
}
result, err := client.CallTool(ctx, "list_pods", args)
```

### list_services

列出指定命名空间中的 Service。

- **函数签名**: `handleListServices`
- **描述**: List services in a namespace

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `ServicesResult` 对象。

```json
{
  "services": "Services:\n  - default/kubernetes (Service) - Type: ClusterIP"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "namespace": "default",
}
result, err := client.CallTool(ctx, "list_services", args)
```

### list_deployments

列出指定命名空间中的 Deployment。

- **函数签名**: `handleListDeployments`
- **描述**: List deployments in a namespace

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `DeploymentsResult` 对象。

```json
{
  "deployments": "Deployments:\n  - default/nginx-deployment (Deployment) - 3/3"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "namespace": "default",
}
result, err := client.CallTool(ctx, "list_deployments", args)
```

### get_resource

获取特定资源的详细信息（JSON 格式）。如果是 Secret 资源，敏感数据会被脱敏。

- **函数签名**: `handleGetResource`
- **描述**: Get detailed information about a specific resource (JSON format)

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `resource_type` | string | 是 | 资源类型 (例如: 'pods', 'services', 'deployments') |
| `name` | string | 是 | 资源名称 |
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `ResourceResult` 对象，包含资源的 JSON 字符串。

```json
{
  "resource": "{\n  \"kind\": \"Pod\",\n  \"apiVersion\": \"v1\",\n  ...\n}"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "resource_type": "pods",
    "name": "nginx-pod",
    "namespace": "default",
}
result, err := client.CallTool(ctx, "get_resource", args)
```

### get_resource_yaml

获取资源的完整 YAML 定义。Secret 数据会被脱敏。

- **函数签名**: `handleGetResourceYAML`
- **描述**: Get the full YAML definition of a resource

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `resource_type` | string | 是 | 资源类型 |
| `name` | string | 是 | 资源名称 |
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `YAMLResult` 对象。注意：当前实现返回的是 JSON 格式的序列化字符串，客户端可以根据需要处理。

```json
{
  "yaml": "{\n  \"kind\": \"Pod\",\n  ...\n}"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "resource_type": "deployments",
    "name": "nginx-deployment",
    "namespace": "default",
}
result, err := client.CallTool(ctx, "get_resource_yaml", args)
```

---

## 可观测性与调试

### get_events

获取指定命名空间的集群事件。

- **函数签名**: `handleGetEvents`
- **描述**: Get cluster events

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `EventsResult` 对象。

```json
{
  "events": "Events:\n  - default/nginx-pod.1234 (Event) - Normal: Scheduled"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "namespace": "default",
}
result, err := client.CallTool(ctx, "get_events", args)
```

### get_pod_logs

获取 Pod 日志。

- **函数签名**: `handleGetPodLogs`
- **描述**: Get pod logs

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `pod_name` | string | 是 | Pod 名称 |
| `namespace` | string | 是 | 命名空间名称 |
| `container_name` | string | 否 | 容器名称（如果是多容器 Pod 则需要指定） |
| `tail_lines` | int | 否 | 返回日志的尾部行数 (默认 100) |
| `previous` | bool | 否 | 是否获取前一个实例的日志 (默认为 false) |
| `cluster_name` | string | 否 | 集群名称 (可选) |

#### 返回值

返回 `LogsResult` 对象。

```json
{
  "logs": "2023-10-01T12:00:00Z INFO Starting application..."
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "pod_name": "nginx-pod",
    "namespace": "default",
    "tail_lines": 50,
}
result, err := client.CallTool(ctx, "get_pod_logs", args)
```

---

## 安全

### check_rbac_permission

检查当前用户是否有权限执行特定操作 (类似于 `kubectl auth can-i`)。

- **函数签名**: `handleCheckRBACPermission`
- **描述**: Check if the current user has permission to perform an action

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `verb` | string | 是 | 操作动词 (例如: 'get', 'list', 'create', 'delete') |
| `resource` | string | 是 | 资源类型 (例如: 'pods', 'deployments') |
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `RBACPermissionResult` 对象。

```json
{
  "allowed": true,
  "reason": "Permission granted"
}
```

#### 示例代码 (MCP Client)

```go
args := map[string]interface{}{
    "verb": "delete",
    "resource": "pods",
    "namespace": "default",
}
result, err := client.CallTool(ctx, "check_rbac_permission", args)
```
