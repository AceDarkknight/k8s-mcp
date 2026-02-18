# API 接口文档

本文档详细介绍了 k8s-mcp 服务器提供的所有工具接口。每个工具都遵循 MCP (Model Context Protocol) 规范。

## 目录

- [数据结构](#数据结构)
    - [Pod](#pod)
    - [Service](#service)
    - [Deployment](#deployment)
    - [Node](#node)
    - [Namespace](#namespace)
    - [ConfigMap](#configmap)
    - [StatefulSet](#statefulset)
    - [Event](#event)
- [集群管理](#集群管理)
    - [get_cluster_status](#get_cluster_status)
    - [list_nodes](#list_nodes)
    - [list_namespaces](#list_namespaces)
- [资源管理](#资源管理)
    - [list_pods](#list_pods)
    - [list_services](#list_services)
    - [list_deployments](#list_deployments)
    - [list_configmaps](#list_configmaps)
    - [list_statefulsets](#list_statefulsets)
    - [get_resource](#get_resource)
    - [get_resource_yaml](#get_resource_yaml)
- [可观测性与调试](#可观测性与调试)
    - [get_events](#get_events)
    - [get_pod_logs](#get_pod_logs)
- [安全](#安全)
    - [check_rbac_permission](#check_rbac_permission)

---

## 数据结构

### Pod

`Pod` 包含 Kubernetes Pod 的详细信息。

```go
type Pod struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Ready     string            `json:"ready"`
	Restarts  int               `json:"restarts"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

### Service

`Service` 包含 Kubernetes Service 的详细信息。

```go
type Service struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     string            `json:"ports"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

### Deployment

`Deployment` 包含 Kubernetes Deployment 的详细信息。

```go
type Deployment struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	UpToDate  string            `json:"up_to_date"`
	Available string            `json:"available"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

### Node

`Node` 包含 Kubernetes 节点的详细信息。

```go
type Node struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Roles   string            `json:"roles"`
	Version string            `json:"version"`
	Age     string            `json:"age"`
	Labels  map[string]string `json:"labels,omitempty"`
}
```

### Namespace

`Namespace` 包含 Kubernetes 命名空间的详细信息。

```go
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}
```

### ConfigMap

`ConfigMap` 包含 Kubernetes ConfigMap 的基本信息。

```go
type ConfigMap struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	DataCount int               `json:"data_count"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

### StatefulSet

`StatefulSet` 包含 Kubernetes StatefulSet 的详细信息。

```go
type StatefulSet struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

### Event

`Event` 包含 Kubernetes 事件的信息。

```go
type Event struct {
	Type      string            `json:"type"`
	Reason    string            `json:"reason"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Count     int               `json:"count"`
	FirstSeen string            `json:"first_seen"`
	LastSeen  string            `json:"last_seen"`
	Labels    map[string]string `json:"labels,omitempty"`
}
```

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

### list_nodes

列出集群中的所有节点及其状态。

- **函数签名**: `handleListNodes`
- **描述**: List all nodes in the cluster

#### 参数

无

#### 返回值

返回 `NodesResult` 对象，包含 `Node` 对象的 JSON 数组字符串。

```json
{
  "nodes": "[{\"name\":\"node-1\",\"status\":\"Ready\",\"roles\":\"control-plane\",\"version\":\"v1.28.0\",\"age\":\"10d\",\"labels\":{\"kubernetes.io/hostname\":\"node-1\"}}]"
}
```

### list_namespaces

列出集群中的所有命名空间。

- **函数签名**: `handleListNamespaces`
- **描述**: List all namespaces in the cluster

#### 参数

无

#### 返回值

返回 `NamespacesResult` 对象，包含 `Namespace` 对象的 JSON 数组字符串。

```json
{
  "namespaces": "[{\"name\":\"default\",\"status\":\"Active\",\"age\":\"10d\"},{\"name\":\"kube-system\",\"status\":\"Active\",\"age\":\"10d\"}]"
}
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

返回 `PodsResult` 对象，包含 `Pod` 对象的 JSON 数组字符串。

```json
{
  "pods": "[{\"name\":\"nginx-pod\",\"namespace\":\"default\",\"status\":\"Running\",\"ready\":\"1/1\",\"restarts\":0,\"age\":\"10d\",\"labels\":{\"app\":\"nginx\"}}]"
}
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

返回 `ServicesResult` 对象，包含 `Service` 对象的 JSON 数组字符串。

```json
{
  "services": "[{\"name\":\"nginx-svc\",\"namespace\":\"default\",\"type\":\"ClusterIP\",\"cluster_ip\":\"10.96.0.10\",\"ports\":\"80/TCP\",\"age\":\"10d\",\"labels\":{\"app\":\"nginx\"}}]"
}
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

返回 `DeploymentsResult` 对象，包含 `Deployment` 对象的 JSON 数组字符串。

```json
{
  "deployments": "[{\"name\":\"nginx-deploy\",\"namespace\":\"default\",\"ready\":\"3/3\",\"up_to_date\":\"3\",\"available\":\"3\",\"age\":\"10d\",\"labels\":{\"app\":\"nginx\"}}]"
}
```

### list_configmaps

列出指定命名空间中的 ConfigMap。

- **函数签名**: `handleListConfigMaps`
- **描述**: List configmaps in a namespace

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `ConfigMapsResult` 对象，包含 `ConfigMap` 对象的 JSON 数组字符串。

```json
{
  "configmaps": "[{\"name\":\"kube-root-ca.crt\",\"namespace\":\"default\",\"data_count\":1,\"age\":\"10d\"}]"
}
```

### list_statefulsets

列出指定命名空间中的 StatefulSet。

- **函数签名**: `handleListStatefulSets`
- **描述**: List statefulsets in a namespace

#### 参数

| 参数名 | 类型 | 必填 | 描述 |
|:---|:---|:---|:---|
| `namespace` | string | 是 | 命名空间名称 |

#### 返回值

返回 `StatefulSetsResult` 对象，包含 `StatefulSet` 对象的 JSON 数组字符串。

```json
{
  "statefulsets": "[{\"name\":\"web\",\"namespace\":\"default\",\"ready\":\"3/3\",\"age\":\"10d\",\"labels\":{\"app\":\"nginx\"}}]"
}
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

返回 `EventsResult` 对象，包含 `Event` 对象的 JSON 数组字符串。

```json
{
  "events": "[{\"type\":\"Normal\",\"reason\":\"Scheduled\",\"message\":\"Successfully assigned default/nginx-pod to node-1\",\"source\":\"default-scheduler\",\"count\":1,\"first_seen\":\"2024-01-01T00:00:00Z\",\"last_seen\":\"2024-01-01T00:00:00Z\"}]"
}
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
