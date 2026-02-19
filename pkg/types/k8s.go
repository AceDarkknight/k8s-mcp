package types

// Namespace 命名空间信息
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

// NamespacesResult list_namespaces 命令的结果
type NamespacesResult struct {
	Namespaces string `json:"namespaces"`
}

// Pod Pod 信息
type Pod struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Ready     string            `json:"ready"`
	Restarts  int               `json:"restarts"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Service Service 信息
type Service struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     string            `json:"ports"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Deployment Deployment 信息
type Deployment struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	UpToDate  string            `json:"up_to_date"`
	Available string            `json:"available"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Node 节点信息
type Node struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Roles   string            `json:"roles"`
	Version string            `json:"version"`
	Age     string            `json:"age"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// Event 事件信息
type Event struct {
	Type      string            `json:"type"`
	Reason    string            `json:"reason"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Count     int               `json:"count"`
	FirstSeen string            `json:"first_seen"`
	LastSeen  string            `json:"last_seen"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// RBACPermission RBAC 权限检查结果
type RBACPermission struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// PodLogOptions Pod 日志选项
type PodLogOptions struct {
	ContainerName string `json:"container_name,omitempty"`
	TailLines     int    `json:"tail_lines,omitempty"`
	Previous      bool   `json:"previous,omitempty"`
	ClusterName   string `json:"cluster_name,omitempty"`
}

// ConfigMap 信息
type ConfigMap struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	DataCount int               `json:"data_count"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// StatefulSet 信息
type StatefulSet struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels,omitempty"`
}
