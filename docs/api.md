# API 接口文档

本文档详细介绍了 k8s-mcp 服务器提供的所有工具接口。每个工具都遵循 MCP (Model Context Protocol) 规范。

## 目录

- [数据结构](#数据结构)
    - [ResourceInfo](#resourceinfo)
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

## 数据结构

### ResourceInfo

`ResourceInfo` 是所有列表工具返回的核心数据结构，包含 Kubernetes 资源的基本信息。

```go
type ResourceInfo struct {
    Name      string            `json:"name"`
    Namespace string            `json:"namespace,omitempty"`
    Kind      string            `json:"kind"`
    Status    string            `json:"status,omitempty"`
    Age       string            `json:"age,omitempty"`
    Labels    map[string]string `json:"labels,omitempty"`
}
```

| 字段 | 类型 | 描述 |
|:---|:---|:---|
| `name` | string | 资源名称 |
| `namespace` | string | 命名空间（集群级别资源此项为空） |
| `kind` | string | 资源类型（如 Pod, Service, Deployment 等） |
| `status` | string | 资源状态（不同资源类型状态格式不同） |
| `age` | string | 资源创建时间 |
| `labels` | map | 资源标签 |

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

返回 `NodesResult` 对象，包含 `ResourceInfo` JSON 数组字符串。

```json
{
  "nodes": [
    {
      "name": "node-1",
      "kind": "Node",
      "status": "Ready",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "kubernetes.io/hostname": "node-1",
        "node-role.kubernetes.io/master": ""
      }
    },
    {
      "name": "node-2",
      "kind": "Node",
      "status": "Ready",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "kubernetes.io/hostname": "node-2"
      }
    }
  ]
}
```

#### 示例代码 (MCP Client)

```go
result, err := client.CallTool(ctx, "list_nodes", nil)
if err != nil {
    log.Fatal(err)
}
// 解析 JSON
var nodes []k8s.ResourceInfo
json.Unmarshal([]byte(result.Content[0].Text), &nodes)
```

### list_namespaces

列出集群中的所有命名空间。

- **函数签名**: `handleListNamespaces`
- **描述**: List all namespaces in the cluster

#### 参数

无

#### 返回值

返回 `NamespacesResult` 对象，包含 `ResourceInfo` JSON 数组字符串。

```json
{
  "namespaces": [
    {
      "name": "default",
      "kind": "Namespace",
      "status": "Active",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "kubernetes.io/metadata.name": "default"
      }
    },
    {
      "name": "kube-system",
      "kind": "Namespace",
      "status": "Active",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "kubernetes.io/metadata.name": "kube-system"
      }
    }
  ]
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

返回 `PodsResult` 对象，包含 `ResourceInfo` JSON 数组字符串。

```json
{
  "pods": [
    {
      "name": "nginx-pod",
      "namespace": "default",
      "kind": "Pod",
      "status": "Running",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "app": "nginx"
      }
    },
    {
      "name": "redis-pod",
      "namespace": "default",
      "kind": "Pod",
      "status": "Running",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "app": "redis"
      }
    }
  ]
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

返回 `ServicesResult` 对象，包含 `ResourceInfo` JSON 数组字符串。

```json
{
  "services": [
    {
      "name": "kubernetes",
      "namespace": "default",
      "kind": "Service",
      "status": "Type: ClusterIP",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "component": "apiserver",
        "provider": "kubernetes"
      }
    },
    {
      "name": "nginx-service",
      "namespace": "default",
      "kind": "Service",
      "status": "Type: LoadBalancer",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "app": "nginx"
      }
    }
  ]
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

返回 `DeploymentsResult` 对象，包含 `ResourceInfo` JSON 数组字符串。状态格式为 `就绪副本数/总副本数`。

```json
{
  "deployments": [
    {
      "name": "nginx-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "status": "3/3",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "app": "nginx"
      }
    },
    {
      "name": "redis-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "status": "1/1",
      "age": "2024-01-01 00:00:00 +0000 UTC",
      "labels": {
        "app": "redis"
      }
    }
  ]
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

返回 `ResourceResult` 对象，包含资源的完整 JSON 字符串。

```json
{
  "resource": "{\n  \"kind\": \"Pod\",\n  \"apiVersion\": \"v1\",\n  \"metadata\": {\n    \"name\": \"nginx-pod\",\n    \"namespace\": \"default\",\n    ...\n  },\n  \"spec\": {\n    ...\n  },\n  \"status\": {\n    ...\n  }\n}"
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

返回 `EventsResult` 对象，包含 `ResourceInfo` JSON 数组字符串。状态格式为 `事件类型: 原因`。

```json
{
  "events": [
    {
      "name": "nginx-pod.12345678",
      "namespace": "default",
      "kind": "Event",
      "status": "Normal: Scheduled",
      "age": "2024-01-01 00:00:00 +0000 UTC"
    },
    {
      "name": "nginx-pod.87654321",
      "namespace": "default",
      "kind": "Event",
      "status": "Normal: Pulled",
      "age": "2024-01-01 00:00:01 +0000 UTC"
    }
  ]
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
  "logs": "2023-10-01T12:00:00Z INFO Starting application...\n2023-10-01T12:00:01Z INFO Server listening on port 8080"
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
