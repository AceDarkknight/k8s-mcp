package mcpclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// tokenAuthTransport 包装 http.RoundTripper 以添加授权头
// tokenAuthTransport wraps http.RoundTripper to add authorization header
type tokenAuthTransport struct {
	token         string
	customHeaders map[string]string
	transport     http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口
// RoundTrip implements http.RoundTripper interface
func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 添加授权头
	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))

	// 添加自定义头
	// Add custom headers
	for key, value := range t.customHeaders {
		req.Header.Set(key, value)
	}

	return t.transport.RoundTrip(req)
}

// createHTTPClient 创建带有 Token 认证和自定义头的 HTTP 客户端
// createHTTPClient creates an HTTP client with token authentication and custom headers
func createHTTPClient(config Config, customHeaders map[string]string) *http.Client {
	// 创建基础 HTTP 客户端
	// Create base HTTP client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
		},
	}

	// 注入 Token 和自定义头到请求中
	// Inject token and custom headers into requests
	tokenTransport := &tokenAuthTransport{
		token:         config.AuthToken,
		customHeaders: customHeaders,
		transport:     httpClient.Transport,
	}
	httpClient.Transport = tokenTransport

	return httpClient
}
