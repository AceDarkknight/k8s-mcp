package logger

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

// TestNewDefaultConsoleLogger 测试默认控制台 logger
func TestNewDefaultConsoleLogger(t *testing.T) {
	log := NewDefaultConsoleLogger()

	log.Debug("This is a debug message", "key", "value")
	log.Info("This is an info message", "user", "alice")
	log.Warn("This is a warning message", "reason", "test")
	log.Error("This is an error message", "error", "test error")

	// 测试 With 方法
	subLogger := log.With("service", "test-service", "version", "1.0.0")
	subLogger.Info("Message from sub logger")
}

// TestInit 测试全局初始化
func TestInit(t *testing.T) {
	// 测试默认配置
	cfg := NewDefaultConfig()
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	log := Get()
	log.Info("Test global logger", "test", "init")

	// 测试开发配置
	devCfg := NewDevelopmentConfig()
	if err := Init(devCfg); err != nil {
		t.Fatalf("Init development config failed: %v", err)
	}
	log = Get()
	log.Debug("Development debug message")

	// 测试生产配置
	prodCfg := NewProductionConfig()
	if err := Init(prodCfg); err != nil {
		t.Fatalf("Init production config failed: %v", err)
	}
	log = Get()
	log.Info("Production info message")
}

// TestBindFlags 测试 flag 绑定
func TestBindFlags(t *testing.T) {
	cfg := NewDefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	BindFlags(fs, cfg)

	// 测试默认值
	if cfg.Level != "info" {
		t.Errorf("Expected default level 'info', got '%s'", cfg.Level)
	}
	if cfg.Format != "text" {
		t.Errorf("Expected default format 'text', got '%s'", cfg.Format)
	}

	// 测试设置值
	args := []string{
		"--log-level=debug",
		"--log-format=json",
		"--log-to-file=true",
		"--log-file=/tmp/test.log",
	}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse flags failed: %v", err)
	}

	if cfg.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", cfg.Format)
	}

	// 测试 AdjustOutputPaths
	AdjustOutputPaths(cfg, true)
	if len(cfg.OutputPaths) != 2 {
		t.Errorf("Expected 2 output paths, got %d", len(cfg.OutputPaths))
	}
}

// TestAdjustOutputPaths 测试输出路径调整
func TestAdjustOutputPaths(t *testing.T) {
	cfg := NewDefaultConfig()

	// 测试不启用文件输出
	AdjustOutputPaths(cfg, false)
	if len(cfg.OutputPaths) != 1 || cfg.OutputPaths[0] != "stdout" {
		t.Errorf("Expected only stdout, got %v", cfg.OutputPaths)
	}

	// 测试启用文件输出
	AdjustOutputPaths(cfg, true)
	if len(cfg.OutputPaths) != 2 {
		t.Errorf("Expected 2 output paths, got %d", len(cfg.OutputPaths))
	}
}

// TestConfigToZapLevel 测试日志级别转换
func TestConfigToZapLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"unknown", "info"}, // 默认为 info
	}

	for _, tt := range tests {
		cfg := &Config{Level: tt.level}
		zapLevel := cfg.toZapLevel()
		if zapLevel.String() != tt.expected {
			t.Errorf("Expected level '%s', got '%s'", tt.expected, zapLevel.String())
		}
	}
}

// TestLoggerWith 测试 With 方法
func TestLoggerWith(t *testing.T) {
	log := NewDefaultConsoleLogger()

	// 创建子 logger
	subLogger := log.With("key1", "value1", "key2", "value2")

	// 验证子 logger 是 Logger 接口的实现
	if _, ok := subLogger.(Logger); !ok {
		t.Error("subLogger should implement Logger interface")
	}

	// 测试嵌套 With
	nestedLogger := subLogger.With("key3", "value3")
	if _, ok := nestedLogger.(Logger); !ok {
		t.Error("nestedLogger should implement Logger interface")
	}

	// 输出日志验证
	nestedLogger.Info("Nested logger test")
}

// TestInitWithNilConfig 测试 nil 配置
func TestInitWithNilConfig(t *testing.T) {
	if err := Init(nil); err != nil {
		t.Fatalf("Init with nil config failed: %v", err)
	}
	defer Sync()

	log := Get()
	if log == nil {
		t.Error("Expected non-nil logger")
	}
}

// TestRotationConfig 测试轮转配置
func TestRotationConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	// 验证默认轮转配置
	if cfg.RotationConfig == nil {
		t.Error("RotationConfig should not be nil")
	}

	if cfg.RotationConfig.MaxSize != 100 {
		t.Errorf("Expected MaxSize 100, got %d", cfg.RotationConfig.MaxSize)
	}

	if cfg.RotationConfig.MaxBackups != 3 {
		t.Errorf("Expected MaxBackups 3, got %d", cfg.RotationConfig.MaxBackups)
	}

	if cfg.RotationConfig.MaxAge != 30 {
		t.Errorf("Expected MaxAge 30, got %d", cfg.RotationConfig.MaxAge)
	}

	if !cfg.RotationConfig.Compress {
		t.Error("Expected Compress to be true")
	}
}

// TestInitialFields 测试初始字段
func TestInitialFields(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.InitialFields = map[string]interface{}{
		"service": "test-service",
		"version": "1.0.0",
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	log := Get()
	log.Info("Test with initial fields")
}

// TestEnableCaller 测试调用者信息
func TestEnableCaller(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.EnableCaller = true

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	log := Get()
	log.Info("Test with caller info")
}

// TestEnableStacktrace 测试堆栈信息
func TestEnableStacktrace(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.EnableStacktrace = true

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	log := Get()
	log.Error("Test with stacktrace", "error", "test error")
}

// TestSync 测试 Sync 方法
func TestSync(t *testing.T) {
	// 测试未初始化时的 Sync
	err := Sync()
	if err != nil {
		t.Errorf("Sync should not error when logger is not initialized: %v", err)
	}

	// 测试初始化后的 Sync
	log := NewDefaultConsoleLogger()
	globalLogger = log
	err = Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

// TestGet 测试 Get 方法
func TestGet(t *testing.T) {
	// 测试未初始化时的 Get
	globalLogger = nil
	log := Get()
	if log == nil {
		t.Error("Get should return default console logger when not initialized")
	}

	// 测试初始化后的 Get
	cfg := NewDefaultConfig()
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	log = Get()
	if log == nil {
		t.Error("Get should return initialized logger")
	}
}

// TestFileOutput 测试文件输出
func TestFileOutput(t *testing.T) {
	// 创建临时日志文件
	tmpFile := "/tmp/test_logger.log"

	cfg := NewDefaultConfig()
	cfg.OutputPaths = []string{tmpFile}
	cfg.ErrorOutputPaths = []string{tmpFile}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()
	defer os.Remove(tmpFile)

	log := Get()
	log.Info("Test file output", "file", tmpFile)

	// 验证文件是否存在
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Log file should exist")
	}
}
