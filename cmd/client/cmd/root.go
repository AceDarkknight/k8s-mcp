package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AceDarkknight/k8s-mcp/pkg/mcpclient"

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

	// Create client configuration
	// 创建客户端配置
	config := mcpclient.Config{
		ServerURL:          serverURL,
		AuthToken:          authToken,
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Create client instance
	// 创建客户端实例
	client, err := mcpclient.NewClient(config, mcpclient.WithUserAgent("k8s-mcp-client/1.0.0"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Connect to server
	// 连接到服务器
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

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

		if err := handleCommand(ctx, client, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// handleCommand processes user commands
// handleCommand 处理用户命令
func handleCommand(ctx context.Context, client *mcpclient.Client, input string) error {
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
		return listTools(ctx, client)
	case "call":
		if len(parts) < 2 {
			fmt.Println("Usage: call <tool_name> [args...]")
			return nil
		}
		return callTool(ctx, client, parts[1], parts[2:])
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

func listTools(ctx context.Context, client *mcpclient.Client) error {
	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Println("Available tools:")
	for _, tool := range tools {
		fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
	}

	return nil
}

func callTool(ctx context.Context, client *mcpclient.Client, toolName string, args []string) error {
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
	result, err := client.CallTool(ctx, toolName, arguments)
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
