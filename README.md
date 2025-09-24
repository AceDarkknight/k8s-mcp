# k8s-mcp

A Model Context Protocol (MCP) server for Kubernetes cluster management and resource viewing.

## Features

- 🔗 Connect to multiple Kubernetes clusters
- 🔄 Switch between clusters dynamically  
- 👀 View Kubernetes resources (read-only)
- 🛡️ Secure, read-only access by default
- 🌐 Support for both stdio and HTTP transports

## Architecture

- **MCP Server** (Golang): Provides k8s cluster connection and resource viewing capabilities
- **MCP Client** (Golang): Test client for validating server functionality

## Quick Start

### Prerequisites

- Go 1.21 or later
- kubectl configured with cluster access
- Valid kubeconfig file

### Building

```bash
# Build the MCP server
go build -o bin/k8s-mcp-server ./cmd/server

# Build the test client
go build -o bin/k8s-mcp-client ./cmd/client
```

### Running

```bash
# Start the MCP server (stdio mode)
./bin/k8s-mcp-server

# Test with the client
./bin/k8s-mcp-client ./bin/k8s-mcp-server
```

## Configuration

The server supports configuration via:
- Environment variables
- kubeconfig files
- Direct cluster configuration

## MCP Features

### Resources
- `k8s://clusters` - Available cluster list
- `k8s://cluster/{cluster-name}/namespaces` - Cluster namespaces
- `k8s://cluster/{cluster-name}/resources/{resource-type}` - Resource lists
- `k8s://cluster/{cluster-name}/resource/{resource-type}/{namespace}/{name}` - Resource details

### Tools
- `list_clusters` - List all available clusters
- `switch_cluster` - Switch active cluster
- `list_namespaces` - List cluster namespaces
- `list_resources` - List resources of specified type
- `get_resource` - Get specific resource details
- `describe_resource` - Get detailed resource description

### Prompts
- `analyze_cluster_health` - Cluster health analysis prompt
- `troubleshoot_pods` - Pod troubleshooting prompt
- `resource_summary` - Resource summary analysis prompt

## Security

- All operations are read-only by default
- Supports RBAC permission validation
- Secure kubeconfig handling
- Connection timeout and retry mechanisms
"# k8s-mcp" 
