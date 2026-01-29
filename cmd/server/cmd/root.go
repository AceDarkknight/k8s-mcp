package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/AceDarkknight/k8s-mcp/internal/mcp"
	"github.com/AceDarkknight/k8s-mcp/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Configuration flags
	// 配置标志
	cfgPort       string
	cfgCertPath   string
	cfgKeyPath    string
	cfgInsecure   bool
	cfgAuthToken  string
	cfgConfigPath string

	// 日志配置
	logConfig = logger.NewDefaultConfig()
)

// initConfig initializes configuration from flags and environment variables
// initConfig 从标志和环境变量初始化配置
func initConfig() {
	// Bind environment variables
	// 绑定环境变量
	viper.BindEnv("port", "MCP_PORT")
	viper.BindEnv("cert", "MCP_CERT")
	viper.BindEnv("key", "MCP_KEY")
	viper.BindEnv("insecure", "MCP_INSECURE")
	viper.BindEnv("token", "MCP_TOKEN")
	viper.BindEnv("kubeconfig", "MCP_KUBECONFIG")
}

func init() {
	cobra.OnInitialize(initConfig)

	// Define flags directly on command
	// 直接在命令上定义标志
	rootCmd.Flags().StringVarP(&cfgPort, "port", "p", "8443", "Port to listen on")
	rootCmd.Flags().StringVarP(&cfgCertPath, "cert", "c", "", "Path to TLS certificate file (required for HTTPS)")
	rootCmd.Flags().StringVarP(&cfgKeyPath, "key", "k", "", "Path to TLS key file (required for HTTPS)")
	rootCmd.Flags().BoolVarP(&cfgInsecure, "insecure", "i", false, "Run in insecure HTTP mode (default is HTTPS)")
	rootCmd.Flags().StringVarP(&cfgAuthToken, "token", "t", "", "Authentication token (required)")
	rootCmd.Flags().StringVarP(&cfgConfigPath, "kubeconfig", "", "", "Path to kubeconfig file (optional)")

	// Bind flags to viper
	// 将标志绑定到 viper
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("cert", rootCmd.Flags().Lookup("cert"))
	viper.BindPFlag("key", rootCmd.Flags().Lookup("key"))
	viper.BindPFlag("insecure", rootCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))

	// Bind logger flags
	// 绑定日志标志（包括 log-to-file）
	logger.BindFlags(rootCmd.PersistentFlags(), logConfig)
}

// rootCmd represents the base command when called without any subcommands
// rootCmd 表示不带任何子命令时调用的基本命令
var rootCmd = &cobra.Command{
	Use:   "k8s-mcp-server",
	Short: "Kubernetes MCP Server",
	Long: `k8s-mcp-server 是一个用于 Kubernetes 集群管理的 MCP 服务器。
它通过 HTTP/SSE 提供对 Kubernetes 资源的只读访问，并支持 Token 认证。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 初始化日志系统
		// Server 端默认启用日志文件输出
		// 从 viper 获取 log-to-file 标志的值，如果没有设置则默认为 true
		logToFile := true
		if cmd.Flags().Changed("log-to-file") {
			logToFile = viper.GetBool("log-to-file")
		}
		logger.AdjustOutputPaths(logConfig, logToFile)
		if err := logger.Init(logConfig); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		executeServer()
	},
}

// Execute runs the server
// Execute 运行服务器
func Execute() error {
	return rootCmd.Execute()
}

// executeServer starts the MCP server
// executeServer 启动 MCP 服务器
func executeServer() {
	// 获取 logger 实例
	log := logger.Get()

	// Read configuration from viper (flags override env vars)
	// 从 viper 读取配置（标志覆盖环境变量）
	port := viper.GetString("port")
	certPath := viper.GetString("cert")
	keyPath := viper.GetString("key")
	insecure := viper.GetBool("insecure")
	authToken := viper.GetString("token")
	configPath := viper.GetString("kubeconfig")

	// Validate required parameters
	// 验证必需参数
	if authToken == "" {
		log.Error("--token is required")
		os.Exit(1)
	}

	if !insecure && (certPath == "" || keyPath == "") {
		log.Error("--cert and --key are required for HTTPS mode (default). Use --insecure for HTTP mode.")
		os.Exit(1)
	}

	// Create MCP server
	// 创建 MCP 服务器
	server := mcp.NewServer(authToken)

	// Register tools
	// 注册工具
	server.RegisterTools()

	// Load kubeconfig if provided or use default
	// 加载 kubeconfig（如果提供）或使用默认值
	if err := server.LoadKubeConfig(configPath); err != nil {
		log.Warn("Failed to load kubeconfig", "error", err)
		log.Info("Server will start but won't be able to connect to clusters until kubeconfig is properly configured")
	}

	// Create HTTP handler with authentication
	// 创建带有认证的 HTTP 处理器
	handler := server.CreateHTTPHandler()

	// Start server
	// 启动服务器
	addr := fmt.Sprintf(":%s", port)
	log.Info("Starting k8s MCP server", "address", addr)
	if insecure {
		log.Info("Running in INSECURE HTTP mode")
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Error("Server error", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Running in SECURE HTTPS mode")
		if err := http.ListenAndServeTLS(addr, certPath, keyPath, handler); err != nil {
			log.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
