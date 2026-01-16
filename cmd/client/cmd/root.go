package cmd

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Configuration flags
	// 配置标志
	cfgServerURL          string
	cfgAuthToken          string
	cfgInsecureSkipVerify bool
)

// rootCmd represents the base command when called without any subcommands
// rootCmd 表示不带任何子命令时调用的基本命令
var rootCmd = &cobra.Command{
	Use:   "k8s-mcp-client",
	Short: "Kubernetes MCP Client",
	Long: `k8s-mcp-client 是一个用于连接到 k8s-mcp 服务器的测试客户端。
它支持通过 HTTP/SSE 连接，并带有 Token 认证。`,
	Run: func(cmd *cobra.Command, args []string) {
		executeClient()
	},
}

// Execute runs the client
// Execute 运行客户端
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Define flags directly on command
	// 直接在命令上定义标志
	rootCmd.Flags().StringVarP(&cfgServerURL, "server", "s", "https://localhost:8443", "MCP server URL")
	rootCmd.Flags().StringVarP(&cfgAuthToken, "token", "t", "", "Authentication token (required)")
	rootCmd.Flags().BoolVarP(&cfgInsecureSkipVerify, "insecure-skip-verify", "i", false, "Skip TLS certificate verification")

	// Bind flags to viper
	// 将标志绑定到 viper
	viper.BindPFlag("server", rootCmd.Flags().Lookup("server"))
	viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
	viper.BindPFlag("insecure-skip-verify", rootCmd.Flags().Lookup("insecure-skip-verify"))
}

// initConfig initializes configuration from flags and environment variables
// initConfig 从标志和环境变量初始化配置
func initConfig() {
	// Bind environment variables
	// 绑定环境变量
	viper.BindEnv("server", "MCP_CLIENT_SERVER")
	viper.BindEnv("token", "MCP_CLIENT_TOKEN")
	viper.BindEnv("insecure-skip-verify", "MCP_CLIENT_INSECURE_SKIP_VERIFY")
}

// tokenAuthTransport wraps http.RoundTripper to add authorization header
// tokenAuthTransport 包装 http.RoundTripper 以添加授权头
type tokenAuthTransport struct {
	token     string
	transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper interface
// RoundTrip 实现 http.RoundTripper 接口
func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add authorization header
	// 添加授权头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return t.transport.RoundTrip(req)
}

// executeClient starts the MCP client
// executeClient 启动 MCP 客户端
func executeClient() {
	// Read configuration from viper (flags override env vars)
	// 从 viper 读取配置（标志覆盖环境变量）
	serverURL := viper.GetString("server")
	authToken := viper.GetString("token")
	insecureSkipVerify := viper.GetBool("insecure-skip-verify")

	// Validate required parameters
	// 验证必需参数
	if authToken == "" {
		log.Fatal("Error: --token is required")
	}

	// Create HTTP client with token authentication
	// 创建带有 Token 认证的 HTTP 客户端
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
		},
	}

	// Inject token into requests using a custom transport wrapper
	// 使用自定义传输包装器在请求中注入 Token
	tokenTransport := &tokenAuthTransport{
		token:     authToken,
		transport: httpClient.Transport,
	}
	httpClient.Transport = tokenTransport

	// Create MCP client
	// 创建 MCP 客户端
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "k8s-mcp-client",
		Version: "1.0.0",
	}, nil)

	// Create streamable transport with custom HTTP client
	// 创建带有自定义 HTTP 客户端的可流式传输
	transport := &mcp.StreamableClientTransport{
		Endpoint:   serverURL,
		HTTPClient: httpClient,
	}

	// Connect to server
	// 连接到服务器
	ctx := context.Background()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer session.Close()

	fmt.Printf("Connected to: %s\n", serverURL)
	fmt.Println("Type 'help' for available commands, 'quit' to exit")

	// Interactive loop
	// 交互式循环
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			break
		}

		if err := handleCommand(ctx, session, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// handleCommand processes user commands
// handleCommand 处理用户命令
func handleCommand(ctx context.Context, session *mcp.ClientSession, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "help":
		showHelp()
		return nil
	case "tools":
		return listTools(ctx, session)
	case "call":
		if len(parts) < 2 {
			fmt.Println("Usage: call <tool_name> [args...]")
			return nil
		}
		return callTool(ctx, session, parts[1], parts[2:])
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		return nil
	}
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                     - Show this help")
	fmt.Println("  tools                    - List available tools")
	fmt.Println("  call <tool> [args...]    - Call a tool")
	fmt.Println("  quit                     - Exit the client")
	fmt.Println()
	fmt.Println("Example tool calls:")
	fmt.Println("  call get_cluster_status")
	fmt.Println("  call list_pods namespace=default")
	fmt.Println("  call get_events namespace=default")
	fmt.Println("  call get_pod_logs pod_name=my-pod namespace=default")
}

func listTools(ctx context.Context, session *mcp.ClientSession) error {
	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Println("Available tools:")
	for _, tool := range result.Tools {
		fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
	}

	return nil
}

func callTool(ctx context.Context, session *mcp.ClientSession, toolName string, args []string) error {
	// Parse simple arguments (key=value format)
	// 解析简单参数（key=value 格式）
	arguments := make(map[string]interface{})
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			arguments[parts[0]] = parts[1]
		}
	}

	// Call tool
	// 调用工具
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}

	// Display result
	// 显示结果
	if result.IsError {
		fmt.Println("Tool execution error:")
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
	}

	return nil
}
