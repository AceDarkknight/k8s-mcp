package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 定义日志接口，采用 SugaredLogger 风格
// 这个接口设计简洁，便于其他日志库实现
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	With(keysAndValues ...interface{}) Logger
}

// zapLoggerWrapper 是 Logger 接口的 zap 实现
type zapLoggerWrapper struct {
	sugar *zap.SugaredLogger
}

// Debug 记录调试级别日志
func (l *zapLoggerWrapper) Debug(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Info 记录信息级别日志
func (l *zapLoggerWrapper) Info(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warn 记录警告级别日志
func (l *zapLoggerWrapper) Warn(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error 记录错误级别日志
func (l *zapLoggerWrapper) Error(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// With 创建带有额外字段的子 logger
func (l *zapLoggerWrapper) With(keysAndValues ...interface{}) Logger {
	return &zapLoggerWrapper{sugar: l.sugar.With(keysAndValues...)}
}

// 全局 logger 实例
var globalLogger Logger

// Init 初始化全局 logger
// 使用提供的配置创建 logger 实例
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	// 构建 zap logger
	zapLogger, err := buildZapLogger(cfg)
	if err != nil {
		return err
	}

	// 包装为我们的 Logger 接口
	globalLogger = &zapLoggerWrapper{sugar: zapLogger.Sugar()}
	return nil
}

// Get 获取全局 logger 实例
// 如果未初始化，返回默认的 console logger
func Get() Logger {
	if globalLogger == nil {
		return NewDefaultConsoleLogger()
	}
	return globalLogger
}

// NewDefaultConsoleLogger 创建默认的控制台 logger
// 供 Client 默认使用，无需全局初始化
// 输出格式：Text（控制台友好），级别：Info
func NewDefaultConsoleLogger() Logger {
	// 使用开发配置，输出到控制台
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建控制台编码器
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建控制台输出
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// 创建 logger，启用调用者信息
	zapLogger := zap.New(consoleCore, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLoggerWrapper{sugar: zapLogger.Sugar()}
}

// buildZapLogger 根据配置构建 zap logger
func buildZapLogger(cfg *Config) (*zap.Logger, error) {
	// 获取日志级别
	level := cfg.toZapLevel()

	// 获取编码器配置
	encoderConfig := cfg.getEncoderConfig()

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 构建输出 cores
	var cores []zapcore.Core

	// 控制台输出
	for _, path := range cfg.OutputPaths {
		if path == "stdout" {
			cores = append(cores, zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stdout),
				level,
			))
		} else if path == "stderr" {
			cores = append(cores, zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stderr),
				level,
			))
		} else {
			// 文件输出，支持日志轮转
			writer := &lumberjack.Logger{
				Filename:   path,
				MaxSize:    cfg.RotationConfig.MaxSize,
				MaxBackups: cfg.RotationConfig.MaxBackups,
				MaxAge:     cfg.RotationConfig.MaxAge,
				Compress:   cfg.RotationConfig.Compress,
			}
			cores = append(cores, zapcore.NewCore(
				encoder,
				zapcore.AddSync(writer),
				level,
			))
		}
	}

	// 使用 Tee 组合多个 core
	core := zapcore.NewTee(cores...)

	// 构建 logger
	opts := []zap.Option{}
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	if len(cfg.InitialFields) > 0 {
		// 将 map 转换为 zap.Field 切片
		fields := make([]zap.Field, 0, len(cfg.InitialFields))
		for k, v := range cfg.InitialFields {
			fields = append(fields, zap.Any(k, v))
		}
		opts = append(opts, zap.Fields(fields...))
	}

	return zap.New(core, opts...), nil
}

// Sync 同步所有缓冲的日志条目
// 应该在程序退出前调用
func Sync() error {
	if globalLogger != nil {
		if zl, ok := globalLogger.(*zapLoggerWrapper); ok {
			return zl.sugar.Sync()
		}
	}
	return nil
}
