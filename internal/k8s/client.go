package k8s

import (
	"context"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClusterManager manages multiple k8s clusters
type ClusterManager struct {
	clusters       map[string]*kubernetes.Clientset
	configs        map[string]*rest.Config
	currentCluster string
}

// NewClusterManager creates a new cluster manager
func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*kubernetes.Clientset),
		configs:  make(map[string]*rest.Config),
	}
}

// LoadKubeConfig loads kubeconfig and initializes clusters
func (cm *ClusterManager) LoadKubeConfig(configPath string) error {
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clients for each cluster context
	for contextName, context := range config.Contexts {
		clusterName := context.Cluster

		// Build config for this context
		clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		})

		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return fmt.Errorf("failed to create config for context %s: %w", contextName, err)
		}

		// Create kubernetes client
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return fmt.Errorf("failed to create client for context %s: %w", contextName, err)
		}

		cm.clusters[clusterName] = clientset
		cm.configs[clusterName] = restConfig

		// Set first cluster as current if none set
		if cm.currentCluster == "" {
			cm.currentCluster = clusterName
		}
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

// GetClient returns the kubernetes client for the current cluster
func (cm *ClusterManager) GetClient() (*kubernetes.Clientset, error) {
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
	client, err := cm.GetClient()
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
