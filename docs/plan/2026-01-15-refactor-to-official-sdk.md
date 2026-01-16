# 使用官方 Go SDK 重构 k8s-mcp 计划

## 1. 背景
目前的 `k8s-mcp` 项目使用了自定义的 MCP 协议实现。为了跟进官方标准，提高维护性和扩展性，计划使用官方提供的 SDK (`github.com/modelcontextprotocol/go-sdk`) 进行重构，并将通信方式从 Stdio 调整为 HTTPS/HTTP (SSE)。

## 2. 目标
- 引入 `github.com/modelcontextprotocol/go-sdk` 依赖。
- 重构 Server 端，使用 SDK 实现 MCP Server。
- 重构 Client 端，使用 SDK 实现 MCP Client。
- **架构调整**:
    - 一个 k8s 集群部署一个 Server 实例。
    - Client 与 Server 一对一连接。
- **通信安全**:
    - **默认 HTTPS**: Server 默认使用 HTTPS 通信。需提供证书和密钥。
    - **HTTP 降级**: 提供启动参数（如 `--insecure`）允许降级为 HTTP。
    - **认证**: Client 和 Server 通信需要认证（Token 机制）。Token 在 Server 启动时配置。
- **功能增强 (完整性与安全性)**:
    - 增加日志查看、权限检查、YAML 获取等实用接口。
    - **敏感数据脱敏**: 针对 Secret 等资源进行自动脱敏。
    - **大日志处理**: 针对日志获取增加行数和字节限制。

## 3. 架构设计
遵循 [MCP Architecture](https://modelcontextprotocol.io/docs/learn/architecture) 标准。

- **Server**:
    - 运行模式：HTTPS (默认) 或 HTTP (可选)。
    - 认证方式：Token Auth (Middleware)。
    - 内部维护 k8s client 连接。
- **Client**:
    - 通过 HTTPS/HTTP 协议连接 Server。
    - 请求头携带认证 Token。
- **Transport**: SSE (Server-Sent Events) for server-to-client, POST for client-to-server.

## 4. 实施步骤

### 第一步：依赖管理
- 清理旧的模块依赖。
- 添加 `github.com/modelcontextprotocol/go-sdk` 依赖。

### 第二步：Server 端重构 (`cmd/server` & `internal/mcp`)
1.  **配置与启动参数**:
    - 添加 flags: `--port`, `--cert`, `--key`, `--insecure`, `--token`, `--kubeconfig`。
2.  **认证中间件**:
    - 实现 HTTP Middleware，验证请求头 `Authorization: Bearer <token>`。
3.  **MCP Server 初始化**:
    - 使用 `mcp.NewServer` + `mcp.StreamableHTTPHandler`。
4.  **HTTPS/HTTP 启动**:
    - 支持 TLS 和非 TLS 启动模式。
5.  **工具注册 (`internal/mcp`)**:
    使用 SDK `mcp.AddTool` 注册以下工具：
    
    **基础资源管理**:
    - `get_cluster_status`: 获取集群版本、节点数等概览。
    - `list_pods`: 列出 Pod (支持 namespace 过滤)。
    - `list_services`: 列出 Service。
    - `list_deployments`: 列出 Deployment。
    - `list_nodes`: 列出 Node 及其状态。
    - `get_resource`: 获取资源详情 (JSON)。**安全增强**: 若资源类型为 `Secret`，默认将 `data` 和 `stringData` 字段内容替换为 `***REDACTED***`。
    - `get_resource_yaml`: 获取资源的完整 YAML 定义。**安全增强**: 同上，针对 Secret 进行脱敏处理。
    
    **可观测性与调试**:
    - `get_events`: 获取集群事件 (排查问题首选)。
    - `get_pod_logs`: 获取 Pod 日志。**性能增强**: 
        - 强制默认 `tail_lines=100`。
        - 增加 `max_bytes` 限制 (例如 1MB)，防止日志过大导致传输超时或内存溢出。
        - 参数支持: `container_name`, `tail_lines`, `previous` (查看崩溃前的日志)。

    **安全与权限**:
    - `check_rbac_permission`: 检查当前 Token 对应的 K8s 用户是否有权限执行某操作 (模拟 `kubectl auth can-i`)。

### 第三步：Client 端重构 (`cmd/client`)
1.  **配置参数**:
    - `--server`, `--token`, `--insecure-skip-verify`。
2.  **Transport 初始化**:
    - 配置 `mcp.StreamableClientTransport`。
    - 注入 Token 到请求头。
3.  **Client 初始化**:
    - 使用 `mcp.NewClient` 连接。

## 5. 预期效果
- Server 默认监听 HTTPS 端口，需 Token 访问。
- Client 可以携带 Token 连接 Server 并执行指令。
- 敏感数据默认不泄露。
- 大日志获取可控。

## 6. 验证计划
1.  启动 Server (HTTPS + Token)。
2.  验证 `get_resource type=secret` 返回的数据已被脱敏。
3.  验证 `get_pod_logs` 默认只返回最近 100 行。
4.  验证 `check_rbac_permission` 功能。