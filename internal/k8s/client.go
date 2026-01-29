package k8s

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AceDarkknight/k8s-mcp/pkg/logger"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

// Options 定义 ClusterManager 的配置选项
type Options struct {
	// Logger 日志接口，如果为 nil 则使用默认的 console logger
	Logger logger.Logger
}

// ClusterManager manages multiple k8s clusters
type ClusterManager struct {
	clusters       map[string]*kubernetes.Clientset
	configs        map[string]*rest.Config
	currentCluster string
	logger         logger.Logger
}

// NewClusterManager creates a new cluster manager
// 如果 opts 为 nil 或 opts.Logger 为 nil，则使用默认的 console logger
func NewClusterManager(opts *Options) *ClusterManager {
	var log logger.Logger
	if opts != nil && opts.Logger != nil {
		log = opts.Logger
	} else {
		log = logger.NewDefaultConsoleLogger()
	}

	return &ClusterManager{
		clusters: make(map[string]*kubernetes.Clientset),
		configs:  make(map[string]*rest.Config),
		logger:   log,
	}
}

// LoadKubeConfigAndInitCluster loads kubeconfig and initializes clusters
// LoadKubeConfigAndInitCluster 加载 kubeconfig 并初始化集群
func (cm *ClusterManager) LoadKubeConfigAndInitCluster(configPath string) error {
	// Get the config file path
	// 获取配置文件路径
	configPath = cm.getKubeConfigPath(configPath)

	// Load the kubeconfig file
	// 加载 kubeconfig 文件
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clients for each cluster context
	// 为每个集群上下文创建客户端
	for contextName, context := range config.Contexts {
		err := cm.addContextCluster(config, contextName, context)
		if err != nil {
			return err
		}
	}

	return nil
}

// getKubeConfigPath returns the kubeconfig path, using default if not specified
// getKubeConfigPath 返回 kubeconfig 路径，如果未指定则使用默认值
func (cm *ClusterManager) getKubeConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}

	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}

// addContextCluster adds a cluster from a kubeconfig context
// addContextCluster 从 kubeconfig 上下文添加集群
func (cm *ClusterManager) addContextCluster(config *clientcmdapi.Config, contextName string, context *clientcmdapi.Context) error {
	clusterName := context.Cluster

	// Build config for this context
	// 为此上下文构建配置
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to create config for context %s: %w", contextName, err)
	}

	// Create kubernetes client
	// 创建 kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create client for context %s: %w", contextName, err)
	}

	cm.clusters[clusterName] = clientset
	cm.configs[clusterName] = restConfig

	// Set first cluster as current if none set
	// 如果未设置当前集群，则将第一个集群设置为当前集群
	if cm.currentCluster == "" {
		cm.currentCluster = clusterName
	}

	return nil
}

// AddCluster adds a cluster with direct configuration
func (cm *ClusterManager) AddCluster(name string, config *rest.Config) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create client for cluster %s: %w", name, err)
	}

	cm.clusters[name] = clientset
	cm.configs[name] = config

	// Set as current if none set
	if cm.currentCluster == "" {
		cm.currentCluster = name
	}

	return nil
}

// GetClusters returns list of available cluster names
func (cm *ClusterManager) GetClusters() []string {
	clusters := make([]string, 0, len(cm.clusters))
	for name := range cm.clusters {
		clusters = append(clusters, name)
	}
	return clusters
}

// GetCurrentCluster returns the current active cluster name
func (cm *ClusterManager) GetCurrentCluster() string {
	return cm.currentCluster
}

// SwitchCluster switches to a different cluster
func (cm *ClusterManager) SwitchCluster(clusterName string) error {
	if _, exists := cm.clusters[clusterName]; !exists {
		return fmt.Errorf("cluster %s not found", clusterName)
	}
	cm.currentCluster = clusterName
	return nil
}

// GetCurrentClient returns the kubernetes client for the current cluster
func (cm *ClusterManager) GetCurrentClient() (*kubernetes.Clientset, error) {
	if cm.currentCluster == "" {
		return nil, fmt.Errorf("no current cluster set")
	}

	client, exists := cm.clusters[cm.currentCluster]
	if !exists {
		return nil, fmt.Errorf("client for cluster %s not found", cm.currentCluster)
	}

	return client, nil
}

// GetClientForCluster returns the kubernetes client for a specific cluster
func (cm *ClusterManager) GetClientForCluster(clusterName string) (*kubernetes.Clientset, error) {
	client, exists := cm.clusters[clusterName]
	if !exists {
		return nil, fmt.Errorf("client for cluster %s not found", clusterName)
	}
	return client, nil
}

// HealthCheck checks if the current cluster is reachable
func (cm *ClusterManager) HealthCheck(ctx context.Context) error {
	client, err := cm.GetCurrentClient()
	if err != nil {
		return err
	}

	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster %s: %w", cm.currentCluster, err)
	}

	return nil
}

// HealthCheckCluster checks if a specific cluster is reachable
func (cm *ClusterManager) HealthCheckCluster(ctx context.Context, clusterName string) error {
	client, err := cm.GetClientForCluster(clusterName)
	if err != nil {
		return err
	}

	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster %s: %w", clusterName, err)
	}

	return nil
}
