# 使用 Cobra 框架重构 k8s-mcp 计划

## 1. 背景
目前 `k8s-mcp` 的 Server 和 Client 均使用标准库 `flag` 进行命令行参数解析。为了提高命令行工具的易用性和可扩展性，并支持从环境变量读取配置，计划使用 `github.com/spf13/cobra` 和 `github.com/spf13/viper` 进行重构。

## 2. 目标
- 引入 `github.com/spf13/cobra` 和 `github.com/spf13/viper` 依赖。
- 重构 Server 端 (`cmd/server`)，使用 Cobra 构建命令结构。
- 重构 Client 端 (`cmd/client`)，使用 Cobra 构建命令结构。
- **功能保持不变**: 现有的所有参数和功能必须完全保留。
- **环境变量支持**: 如果命令行参数未指定，自动从环境变量中读取（例如 `--port` 对应 `MCP_PORT`）。

## 3. 架构设计

### Server (`cmd/server`)
- **Root Command**: `k8s-mcp-server`
- **Flags**:
    - `--port` / `MCP_PORT`
    - `--cert` / `MCP_CERT`
    - `--key` / `MCP_KEY`
    - `--insecure` / `MCP_INSECURE`
    - `--token` / `MCP_TOKEN`
    - `--kubeconfig` / `MCP_KUBECONFIG`

### Client (`cmd/client`)
- **Root Command**: `k8s-mcp-client`
- **Flags**:
    - `--server` / `MCP_CLIENT_SERVER`
    - `--token` / `MCP_CLIENT_TOKEN`
    - `--insecure-skip-verify` / `MCP_CLIENT_INSECURE_SKIP_VERIFY`
- **Interactive Mode**: 保持现有的 REPL 交互模式作为默认行为。

## 4. 实施步骤

### 第一步：依赖管理
- 添加 `github.com/spf13/cobra` 和 `github.com/spf13/viper`。

### 第二步：Server 端重构
1.  创建 `cmd/server/cmd/root.go`，定义 Root Command。
2.  使用 `viper` 绑定 Flags 和环境变量。
3.  将原 `main` 函数逻辑迁移到 Root Command 的 `Run` 函数中。
4.  在 `cmd/server/main.go` 中调用 `cmd.Execute()`。

### 第三步：Client 端重构
1.  创建 `cmd/client/cmd/root.go`，定义 Root Command。
2.  使用 `viper` 绑定 Flags 和环境变量。
3.  将原 `main` 函数逻辑迁移到 Root Command 的 `Run` 函数中。
4.  在 `cmd/client/main.go` 中调用 `cmd.Execute()`。

## 5. 预期效果
- 命令行参数解析更加健壮。
- 支持 `k8s-mcp-server --help` 查看详细帮助。
- 支持通过环境变量配置，便于容器化部署。

## 6. 验证计划
1.  构建 Server 和 Client。
2.  使用 `--help` 检查参数说明。
3.  测试命令行参数启动。
4.  测试环境变量启动（不带命令行参数）。