package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// Config 定义日志配置结构体
type Config struct {
	// Level 日志级别 (debug, info, warn, error)
	Level string

	// Format 日志格式 (json, text)
	Format string

	// OutputPaths 输出路径列表，支持控制台和文件
	OutputPaths []string

	// ErrorOutputPaths 错误输出路径列表
	ErrorOutputPaths []string

	// EnableCaller 是否记录调用者信息（文件名和行号）
	EnableCaller bool

	// EnableStacktrace 是否在错误级别记录堆栈信息
	EnableStacktrace bool

	// InitialFields 初始化字段，将出现在所有日志中
	InitialFields map[string]interface{}

	// EncoderConfig 自定义编码器配置
	EncoderConfig *zapcore.EncoderConfig

	// RotationConfig 日志轮转配置
	RotationConfig *RotationConfig
}

// RotationConfig 定义日志轮转配置
type RotationConfig struct {
	// Filename 日志文件路径
	Filename string

	// MaxSize 单个日志文件的最大大小（MB）
	MaxSize int

	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int

	// MaxAge 保留旧日志文件的最大天数
	MaxAge int

	// Compress 是否压缩旧日志文件
	Compress bool
}

// NewDefaultConfig 返回默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Level:            "info",
		Format:           "text",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EnableCaller:     true,
		EnableStacktrace: false,
		InitialFields:    make(map[string]interface{}),
		RotationConfig: &RotationConfig{
			Filename:   "logs/app.log",
			MaxSize:    100, // 100 MB
			MaxBackups: 3,
			MaxAge:     30, // 30 天
			Compress:   true,
		},
	}
}

// NewProductionConfig 返回生产环境配置（JSON格式，输出到文件）
func NewProductionConfig() *Config {
	cfg := NewDefaultConfig()
	cfg.Format = "json"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg
}

// NewDevelopmentConfig 返回开发环境配置（Text格式，输出到控制台）
func NewDevelopmentConfig() *Config {
	cfg := NewDefaultConfig()
	cfg.Format = "text"
	cfg.Level = "debug"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg
}

// toZapLevel 将字符串日志级别转换为 zapcore.Level
func (c *Config) toZapLevel() zapcore.Level {
	switch c.Level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEncoderConfig 获取编码器配置
func (c *Config) getEncoderConfig() zapcore.EncoderConfig {
	if c.EncoderConfig != nil {
		return *c.EncoderConfig
	}

	// 默认编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Text 格式下使用更友好的编码器
	if c.Format == "text" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		}
	}

	return encoderConfig
}
