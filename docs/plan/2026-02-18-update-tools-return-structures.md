# 更新 MCP Tools 返回结构计划

## 1. 目标

更新 MCP Tools 的返回值结构，从通用的 `ResourceInfo` 转换为针对不同 Kubernetes 资源类型的具体结构体。此举旨在提供更详细、更规范的资源信息，增强客户端对集群状态的感知能力。

## 2. 实现步骤

### 2.1 创建类型定义文件
新建 [`internal/k8s/types.go`](internal/k8s/types.go) 文件，用于存放所有 Kubernetes 资源的自定义结构体定义。这有助于保持 [`internal/k8s/resources.go`](internal/k8s/resources.go) 的逻辑清晰。

### 2.2 定义资源结构体
在 [`internal/k8s/types.go`](internal/k8s/types.go) 中定义以下 11 个结构体，用于精确描述 Kubernetes 资源：
- `Namespace`: 命名空间详细信息。
- `NamespacesResult`: 命名空间列表结果。
- `Pod`: Pod 的详细状态，包括 Ready 状态、重启次数、IP 等。
- `Service`: 服务信息，包括 ClusterIP、端口映射等。
- `Deployment`: 部署状态，包括副本数、更新策略等。
- `StatefulSet`: 有状态副本集信息。
- `ConfigMap`: 配置映射信息。
- `Node`: 节点信息，包括状态、版本、资源容量等。
- `Event`: 集群事件。
- `RBACPermission`: 角色/权限信息。
- `PodLogOptions`: Pod 日志查询配置。

### 2.3 修改及新增资源处理函数
调整 [`internal/k8s/resources.go`](internal/k8s/resources.go)：
- 引用 [`internal/k8s/types.go`](internal/k8s/types.go) 中定义的结构体。
- 将现有函数（List/Get）的返回签名修改为具体的结构体，并丰富字段提取逻辑。
- **新增** `ListConfigMaps`, `GetConfigMap`, `ListStatefulSets`, `GetStatefulSet` 等函数，用于支持新资源的查询。

### 2.4 更新 MCP 服务端逻辑
修改 [`internal/mcp/server.go`](internal/mcp/server.go)：
- 更新 `Result` 结构体，使其能够承载新的资源结构。
- **新增** MCP Tool 定义，注册 `list_configmaps`, `get_configmap`, `list_statefulsets`, `get_statefulset` 工具。
- 调整/新增 `handle...` 系列函数（包括现有的和新加的），确保它们正确调用 `internal/k8s` 逻辑并返回符合新结构的 JSON 字符串。

### 2.5 验证与测试
- 确保所有更改不破坏现有的 MCP 协议交互。
- 验证生成的 JSON 数据中包含预期的新字段（如 Pod 的 Ready 状态）。

## 3. 预期效果

- 工具返回的 JSON 数据将更加结构化和详细。
- 客户端（如 AI 助手）可以根据更丰富的字段（如 Service 的 ClusterIP、Pod 的状态详情）做出更精准的判断。
- 提高代码的可维护性和类型安全性。
