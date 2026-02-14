# 计划：添加 list_namespaces 工具

## 1. 目标
在 MCP 服务器中添加一个新的工具 `list_namespaces`，用于列出 Kubernetes 集群中的所有命名空间。

## 2. 需求分析
- **工具名称**: `list_namespaces`
- **功能**: 列出集群中的所有命名空间。
- **参数**: 无需参数。
- **涉及文件**:
    - `internal/k8s/resources.go`: 确认并可能优化命名空间列表获取逻辑。
    - `internal/mcp/server.go`: 注册工具并实现处理函数。

## 3. 设计方案

### 3.1 Kubernetes 逻辑 (`internal/k8s/resources.go`)
虽然 `` 方法已经存在，但为了符合"更新 `internal/k8s/` 逻辑"的要求，我们将检查该方法是否需要优化或添加新的辅助方法。
- 目前 `ListNamespaces` 接受 `clusterName` 参数。
- 新工具不需要参数，我们将复用现有的 `ListNamespaces` 方法，传入空字符串作为 `clusterName` 以使用当前上下文。
- **计划变更**:
    - 检查 `ListNamespaces` 的实现，确保其返回的信息包含所需字段（Name, Status, Age, Labels）。目前看已包含。
    - 如果需要，添加注释或微调错误处理以匹配新工具的需求。

### 3.2 MCP 服务器集成 (`internal/mcp/server.go`)
1.  **定义结果结构体**:
    ```go
    type NamespacesResult struct {
        Namespaces string `json:"namespaces"`
    }
    ```
2.  **注册工具**:
    在 `RegisterTools` 方法中添加 `list_namespaces` 工具注册代码。
    ```go
    mcp.AddTool(s.mcpServer, &mcp.Tool{
        Name:        "list_namespaces",
        Description: "List all namespaces in the cluster",
    }, s.handleListNamespaces)
    ```
3.  **实现处理函数**:
    添加 `handleListNamespaces` 方法：
    - 调用 `s.resourceOps.ListNamespaces(ctx, "")`。
    - 格式化输出字符串（类似于 `kubectl get ns` 的输出格式）。
    - 返回 `NamespacesResult`。

## 4. 测试计划
1.  **单元测试**:
    - 如果存在 `internal/mcp/server_test.go`，添加针对 `handleListNamespaces` 的测试用例。
    - 模拟 `ListNamespaces` 返回数据，验证工具输出格式。
2.  **手动验证**:
    - 编译并运行服务器。
    - 使用 MCP 客户端调用 `list_namespaces`。
    - 验证输出是否列出了当前集群的所有命名空间。
    - 验证与 `kubectl get namespaces` 输出的一致性。

## 5. 文档更新
- 更新 `README.md` 和 `README_zh.md` 中的工具列表，添加 `list_namespaces` 说明。

## 6. 实施步骤
1.  修改 `internal/mcp/server.go` 添加 `NamespacesResult` 结构体。
2.  修改 `internal/mcp/server.go` 实现 `handleListNamespaces` 方法。
3.  修改 `internal/mcp/server.go` 在 `RegisterTools` 中注册工具。
4.  检查 `internal/k8s/resources.go` 确保 `ListNamespaces` 逻辑正确（无需重大修改）。
5.  执行测试。
6.  更新文档。
