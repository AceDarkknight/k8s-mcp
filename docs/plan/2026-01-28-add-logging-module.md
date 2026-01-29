# 2026-01-28 日志模块设计与实现计划 (已更新)

## 1. 需求分析
为了提高 Kubernetes MCP 项目的可观测性和调试能力，需要引入高性能、结构化的日志系统。基于用户反馈，需求更新如下：
- **性能**：基于 Uber Zap，确保高并发下的日志记录效率。
- **轮转**：支持日志文件按大小和日期轮转，避免磁盘溢出。
- **格式**：支持 JSON 格式，方便日志采集系统处理。
- **配置与参数**：支持通过命令行参数 (Cobra flags) 配置日志级别、文件路径、是否启用文件输出等。
- **Server 端集成**：Server 端在启动时应默认初始化并启用日志系统。
- **Client 端集成**：
    - **零配置友好**: Client 在完全不配置 logger 的情况下，能自动退回到一个健康的默认状态（输出到控制台）。
    - **接口解耦**: 采用简洁的接口定义，不强制绑定到特定的日志库实现（如 Zap），方便用户注入自定义实现。
    - **默认行为**: 默认使用 Text 格式（非 JSON）输出到标准输出，以便于控制台调试，且无需用户手动调用全局初始化函数。

## 2. 技术选型
- **日志库**: `go.uber.org/zap`
- **轮转库**: `gopkg.in/natefinch/lumberjack.v2`
- **命令行框架**: `github.com/spf13/cobra` (现有项目已使用)

## 3. 实现步骤

### 第一阶段：模块结构与接口设计
1. **创建新包**: `pkg/logger`
    - `logger.go`: 定义 `Logger` 接口及核心初始化逻辑。
    - `config.go`: 定义日志配置结构体。

2. **接口设计 (解耦关键)**:
    - 定义通用接口：
      ```go
      type Logger interface {
          Debug(msg string, keysAndValues ...interface{})
          Info(msg string, keysAndValues ...interface{})
          Warn(msg string, keysAndValues ...interface{})
          Error(msg string, keysAndValues ...interface{})
          With(keysAndValues ...interface{}) Logger
      }
      ```
    - 提供默认实现获取方法：`NewDefaultConsoleLogger() Logger`。

### 第二阶段：命令行参数支持 (Cobra Flags)
1. **全局 Flag 定义**:
    - `--log-level`: 日志级别 (debug, info, warn, error)，默认 `info`。
    - `--log-file`: 日志输出路径，默认 `logs/app.log`。
    - `--log-to-file`: 是否启用文件输出，默认 `false`。

2. **绑定逻辑**:
    - 在 `pkg/logger` 中提供 `BindFlags(cmd *cobra.Command)` 方法。

### 第三阶段：初始化逻辑实现
1. **Zap Core 配置**:
    - 支持 `zapcore.NewTee`：根据配置决定是否输出到控制台和文件。

2. **Client 默认逻辑**:
    - **自动初始化**: 在 `NewClusterManager` 中，如果 `Options.Logger` 为 `nil`，代码将逻辑自动赋值为 `logger.NewDefaultConsoleLogger()`。
    - **零依赖**: `NewDefaultConsoleLogger` 内部使用预设的 `zap.DevelopmentConfig` 风格构建，不依赖全局 `Init` 函数，确保“开箱即用”。

### 第四阶段：现有代码修改计划
1. **Server 端集成 (`cmd/server`)**:
    - 在 `PersistentPreRunE` 中完成全局初始化。

2. **Client 端集成 (`internal/k8s`)**:
    - 修改 `internal/k8s/client.go`:
        - 引入 `Options` 结构体，包含 `Logger` 字段。
        - 确保 `ClusterManager` 的所有方法使用内部持有的 `logger` 记录关键行为。

## 4. Client 端默认日志行为细化 (特别说明)

### 4.1 零配置初始化流程
当用户以最简方式初始化 Client 时：
```go
// 用户代码
cm := k8s.NewClusterManager() // 内部自动处理日志
```
`NewClusterManager` 内部逻辑：
1. 检查 `opts.Logger` 是否为 `nil`。
2. 若为 `nil`，调用 `pkg/logger.NewDefaultConsoleLogger()`。
3. `NewDefaultConsoleLogger` 会创建一个直接输出到 `os.Stdout` 的 `zap.SugaredLogger` 包装实例。

### 4.2 默认输出规格
- **格式**: **Text (Console)**。使用彩色编码（如果是在终端中），清晰展示时间戳、级别和消息。
- **级别**: `Info`。
- **内容**: 包含调用者信息 (caller) 和结构化键值对。
- **目的**: 确保在本地开发或容器日志查看时，开发者不需要通过 `jq` 等工具解析 JSON 即可直观看到 Client 的运行状态。

### 4.3 接口的简洁性
`pkg/logger.Logger` 接口采用 `Sugared` 风格（`keysAndValues ...interface{}`），这使得它非常容易被其他日志库（如标准库 `log` 的包装，或 `logrus`）适配。如果用户希望将 Client 日志集成到他们已有的系统中，只需实现该接口并注入即可。

## 5. 存量代码重构
在日志模块实现后，必须将项目中现有的所有日志输出（如 `fmt.Print*`, `log.Print*`, `println` 等）全部替换为使用新的 `pkg/logger` 模块。

### 5.1 重点检查和替换范围
- `cmd/server/...`
- `cmd/client/...`
- `internal/mcp/...`
- `internal/k8s/...`
- `pkg/mcpclient/...`

### 5.2 替换原则
- **错误信息**: 使用 `logger.Error` 或 `logger.Warn`。
- **启动与关键路径信息**: 使用 `logger.Info`。
- **调试信息**: 使用 `logger.Debug`。
- **统一格式**: 避免在日志消息中进行复杂的字符串拼接，优先使用结构化键值对（例如：`logger.Info("starting server", "port", 8080)`）。

## 6. 预期效果
- **Server**: 生产级日志，支持轮转和 JSON 格式。
- **Client**: 极简接入，默认提供高质量的控制台输出，同时具备高度的扩展性。

## 7. 模块依赖更新
- `go get go.uber.org/zap`
- `go get gopkg.in/natefinch/lumberjack.v2`
