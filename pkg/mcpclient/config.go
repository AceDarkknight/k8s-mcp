package mcpclient

import (
	"os"
	"strings"
)

// Config 定义客户端配置
// Config defines client configuration
type Config struct {
	ServerURL          string // MCP 服务器地址
	AuthToken          string // 认证 Token
	InsecureSkipVerify bool   // 是否跳过 TLS 证书验证
	UserAgent          string // 可选：标识客户端身份
}

// LoadConfig 从环境变量加载配置
// LoadConfig loads configuration from environment variables
func LoadConfig() (Config, error) {
	cfg := Config{
		ServerURL:          getEnvWithDefault("MCP_CLIENT_SERVER", "https://localhost:8443"),
		AuthToken:          os.Getenv("MCP_CLIENT_TOKEN"),
		InsecureSkipVerify: strings.ToLower(getEnvWithDefault("MCP_CLIENT_INSECURE_SKIP_VERIFY", "false")) == "true",
		UserAgent:          getEnvWithDefault("MCP_CLIENT_USER_AGENT", "k8s-mcp-client/1.0.0"),
	}
	return cfg, nil
}

// getEnvWithDefault 获取环境变量，如果不存在则返回默认值
// getEnvWithDefault gets environment variable or returns default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
