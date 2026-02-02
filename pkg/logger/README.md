# logger 包

## 概述

`logger` 包提供了一个高性能、结构化的日志系统，基于 Uber Zap 实现。该包设计简洁，支持多种输出格式、日志轮转、命令行参数配置等功能。

## 主要特性

- **高性能**: 基于 Uber Zap，确保高并发下的日志记录效率
- **日志轮转**: 支持按大小和日期轮转日志文件，避免磁盘溢出
- **多格式支持**: 支持 JSON 和 Text 两种输出格式
- **灵活配置**: 支持通过命令行参数 (Cobra flags) 配置日志级别、文件路径等
- **接口解耦**: 提供简洁的 `Logger` 接口，便于其他日志库实现
- **零配置友好**: Client 端默认使用控制台输出，无需手动初始化

## 文件说明

- `config.go`: 定义日志配置结构体 `Config` 和 `RotationConfig`
- `logger.go`: 实现 `Logger` 接口和核心初始化逻辑
- `flags.go`: 提供与 Cobra 命令行框架的集成

## 快速开始

### 1. 使用默认控制台 Logger（推荐用于 Client）

```go
import "github.com/your-org/k8s-mcp/pkg/logger"

// 创建默认控制台 logger，无需初始化
log := logger.NewDefaultConsoleLogger()

log.Info("Client started", "version", "1.0.0")
log.Debug("Processing request", "id", 123)
```

### 2. 全局初始化（推荐用于 Server）

```go
import (
    "github.com/spf13/cobra"
    "github.com/your-org/k8s-mcp/pkg/logger"
)

var logConfig = logger.NewDefaultConfig()

func init() {
    // 绑定命令行参数
    logger.BindFlags(rootCmd.Flags(), logConfig)
}

func main() {
    // 解析命令行参数
    rootCmd.Execute()

    // 根据参数调整输出路径
    logToFile, _ := rootCmd.Flags().GetBool("log-to-file")
    logger.AdjustOutputPaths(logConfig, logToFile)

    // 初始化全局 logger
    if err := logger.Init(logConfig); err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 使用全局 logger
    log := logger.Get()
    log.Info("Server started", "port", 8080)
}
```

### 3. 自定义配置

```go
import "github.com/your-org/k8s-mcp/pkg/logger"

// 创建自定义配置
cfg := &logger.Config{
    Level:            "debug",
    Format:           "json",
    OutputPaths:      []string{"stdout", "logs/app.log"},
    ErrorOutputPaths: []string{"stderr", "logs/app.log"},
    EnableCaller:     true,
    EnableStacktrace: true,
    RotationConfig: &logger.RotationConfig{
        Filename:   "logs/app.log",
        MaxSize:    100,
        MaxBackups: 5,
        MaxAge:     30,
        Compress:   true,
    },
}

// 初始化 logger
if err := logger.Init(cfg); err != nil {
    panic(err)
}
defer logger.Sync()

log := logger.Get()
log.Info("Application started")
```

## Logger 接口

```go
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    With(keysAndValues ...interface{}) Logger
}
```

### 使用示例

```go
// 基本日志
log.Info("User logged in", "username", "alice")

// 使用 With 创建带字段的子 logger
userLogger := log.With("user_id", 123, "username", "alice")
userLogger.Info("Processing request")

// 不同级别
log.Debug("Debug information", "detail", "value")
log.Info("Normal operation")
log.Warn("Warning message", "reason", "rate limit")
log.Error("Error occurred", "error", err)
```

## 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--log-level` | string | info | 日志级别 (debug, info, warn, error) |
| `--log-format` | string | text | 日志格式 (json, text) |
| `--log-to-file` | bool | false | 是否启用日志文件输出 |
| `--log-file` | string | logs/app.log | 日志文件路径 |
| `--log-max-size` | int | 100 | 单个日志文件最大大小 (MB) |
| `--log-max-backups` | int | 3 | 保留的旧日志文件最大数量 |
| `--log-max-age` | int | 30 | 保留旧日志文件的最大天数 |
| `--log-compress` | bool | true | 是否压缩旧日志文件 |
| `--log-caller` | bool | true | 是否记录调用者信息 |
| `--log-stacktrace` | bool | false | 是否在错误级别记录堆栈信息 |

## 配置说明

### Config 结构体

```go
type Config struct {
    Level            string              // 日志级别
    Format           string              // 日志格式
    OutputPaths      []string            // 输出路径
    ErrorOutputPaths []string            // 错误输出路径
    EnableCaller     bool                // 是否记录调用者
    EnableStacktrace bool                // 是否记录堆栈
    InitialFields    map[string]interface{} // 初始字段
    EncoderConfig    *zapcore.EncoderConfig // 编码器配置
    RotationConfig   *RotationConfig     // 轮转配置
}
```

### RotationConfig 结构体

```go
type RotationConfig struct {
    Filename   string // 日志文件路径
    MaxSize    int    // 单个日志文件最大大小 (MB)
    MaxBackups int    // 保留的旧日志文件最大数量
    MaxAge     int    // 保留旧日志文件的最大天数
    Compress   bool   // 是否压缩旧日志文件
}
```

## 最佳实践

1. **Client 端**: 使用 `NewDefaultConsoleLogger()`，无需全局初始化
2. **Server 端**: 使用 `Init()` 全局初始化，支持文件输出和轮转
3. **结构化日志**: 使用键值对而非字符串拼接
4. **日志级别**: 
   - Debug: 调试信息
   - Info: 正常操作信息
   - Warn: 警告信息
   - Error: 错误信息
5. **程序退出**: 调用 `logger.Sync()` 确保所有日志写入磁盘

## 依赖

- `go.uber.org/zap`: 高性能日志库
- `gopkg.in/natefinch/lumberjack.v2`: 日志轮转库
- `github.com/spf13/pflag`: 命令行参数解析
