package mcpclient

// Option 定义配置选项函数类型
// Option defines the function type for configuration options
type Option func(*Client)

// WithHeader 添加自定义 HTTP 头
// WithHeader adds a custom HTTP header
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.customHeaders[key] = value
	}
}

// WithUserAgent 设置自定义 User-Agent
// WithUserAgent sets a custom User-Agent
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.config.UserAgent = userAgent
	}
}
