package logger

import (
	"github.com/spf13/pflag"
)

// BindFlags 将日志配置绑定到 pflag.FlagSet
// 这样可以通过命令行参数配置日志
func BindFlags(fs *pflag.FlagSet, cfg *Config) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	// 日志级别
	fs.StringVar(&cfg.Level, "log-level", cfg.Level,
		"日志级别 (debug, info, warn, error)")

	// 日志格式
	fs.StringVar(&cfg.Format, "log-format", cfg.Format,
		"日志格式 (json, text)")

	// 是否启用文件输出
	logToFile := false
	fs.BoolVar(&logToFile, "log-to-file", false,
		"是否启用日志文件输出")

	// 日志文件路径
	fs.StringVar(&cfg.RotationConfig.Filename, "log-file", cfg.RotationConfig.Filename,
		"日志文件路径")

	// 单个日志文件最大大小 (MB)
	fs.IntVar(&cfg.RotationConfig.MaxSize, "log-max-size", cfg.RotationConfig.MaxSize,
		"单个日志文件最大大小 (MB)")

	// 保留的旧日志文件最大数量
	fs.IntVar(&cfg.RotationConfig.MaxBackups, "log-max-backups", cfg.RotationConfig.MaxBackups,
		"保留的旧日志文件最大数量")

	// 保留旧日志文件的最大天数
	fs.IntVar(&cfg.RotationConfig.MaxAge, "log-max-age", cfg.RotationConfig.MaxAge,
		"保留旧日志文件的最大天数")

	// 是否压缩旧日志文件
	fs.BoolVar(&cfg.RotationConfig.Compress, "log-compress", cfg.RotationConfig.Compress,
		"是否压缩旧日志文件")

	// 是否记录调用者信息
	fs.BoolVar(&cfg.EnableCaller, "log-caller", cfg.EnableCaller,
		"是否记录调用者信息（文件名和行号）")

	// 是否在错误级别记录堆栈信息
	fs.BoolVar(&cfg.EnableStacktrace, "log-stacktrace", cfg.EnableStacktrace,
		"是否在错误级别记录堆栈信息")

	// 注册一个 flag 解析后的回调，处理 log-to-file 逻辑
	if fs != nil {
		// 注意：这里需要在 flag 解析后手动调用 AdjustOutputPaths
		// 因为 pflag 没有直接的 post-parse 钩子
	}
}

// AdjustOutputPaths 根据 log-to-file 标志调整输出路径
// 应该在 flag 解析后调用
func AdjustOutputPaths(cfg *Config, logToFile bool) {
	if logToFile {
		// 如果启用了文件输出，同时输出到控制台和文件
		cfg.OutputPaths = []string{"stdout", cfg.RotationConfig.Filename}
		cfg.ErrorOutputPaths = []string{"stderr", cfg.RotationConfig.Filename}
	} else {
		// 仅输出到控制台
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	}
}
