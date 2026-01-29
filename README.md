# k8s-mcp

A Model Context Protocol (MCP) server for Kubernetes cluster management and resource viewing.

## Features

- 🔗 Connect to Kubernetes clusters via HTTP/SSE
- 👀 View Kubernetes resources (read-only)
- 🛡️ Secure access with Token authentication
- 📊 Comprehensive tool set for cluster management
- 🔒 Secret data redaction for security
- 🚀 CLI powered by Cobra and Viper

## Architecture

- **MCP Server** (Golang): Provides k8s cluster connection and resource viewing capabilities via HTTP/SSE
- **MCP Client** (Golang): Test client for validating server functionality
- **pkg/mcpclient** (Golang): Reusable client library for integrating MCP functionality into other Go applications

## Quick Start

### Prerequisites

- Go 1.23 or later
- kubectl configured with cluster access
- Valid kubeconfig file
- TLS certificate and key (for HTTPS mode, default)

### Building

```bash
# Build MCP server
go build -o bin/k8s-mcp-server ./cmd/server

# Build test client
go build -o bin/k8s-mcp-client ./cmd/client
```

### Running

#### 1. Generate TLS Certificate (for HTTPS mode)

```bash
# Generate a self-signed certificate for testing
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout key.pem -out cert.pem -subj "/CN=localhost"
```

#### 2. Start MCP Server

```bash
# Using command line flags
./bin/k8s-mcp-server --token my-secret-token --cert cert.pem --key key.pem

# Using environment variables (flags override env vars)
export MCP_TOKEN=my-secret-token
export MCP_CERT=cert.pem
export MCP_KEY=key.pem
./bin/k8s-mcp-server
```

#### 3. Test with Client

```bash
# Connect to HTTPS server
./bin/k8s-mcp-client --server https://localhost:8443 --token my-secret-token

# Connect to HTTP server
./bin/k8s-mcp-client --server http://localhost:8443 --token my-secret-token --insecure-skip-verify
```

## Configuration

The server supports configuration via command-line flags and environment variables.

### Server Configuration

| Flag | Environment Variable | Default | Description |
|-------|---------------------|---------|-------------|
| `--port` | `MCP_PORT` | 8443 | Port to listen on |
| `--cert` | `MCP_CERT` | | Path to TLS certificate file (required for HTTPS) |
| `--key` | `MCP_KEY` | | Path to TLS key file (required for HTTPS) |
| `--insecure` | `MCP_INSECURE` | false | Run in insecure HTTP mode (default is HTTPS) |
| `--token` | `MCP_TOKEN` | | Authentication token (required) |
| `--kubeconfig` | `MCP_KUBECONFIG` | | Path to kubeconfig file (optional) |

### Logging Configuration

The server provides a comprehensive logging system based on Uber Zap and Lumberjack.

| Flag | Environment Variable | Default | Description |
|-------|---------------------|---------|-------------|
| `--log-level` | | info | Log level (debug, info, warn, error) |
| `--log-format` | | text | Log format (json, text) |
| `--log-to-file` | | false (Server: true) | Enable logging to file (Server defaults to true) |
| `--log-file` | | logs/app.log | Log file path |
| `--log-max-size` | | 100 | Maximum size of each log file (MB) |
| `--log-max-backups`| | 3 | Maximum number of old log files to retain |
| `--log-max-age` | | 30 | Maximum number of days to retain old log files |
| `--log-compress` | | true | Whether to compress old log files |
| `--log-caller` | | true | Whether to include caller information (file and line) |
| `--log-stacktrace` | | false | Whether to include stacktrace on error level |

When `--log-to-file` is enabled, logs are written to both stdout/stderr and the specified log file. The logging system automatically handles log rotation based on size, age, and number of backups.

### Client Configuration

| Flag | Environment Variable | Default | Description |
|-------|---------------------|---------|-------------|
| `--server` | `MCP_CLIENT_SERVER` | https://localhost:8443 | MCP server URL |
| `--token` | `MCP_CLIENT_TOKEN` | | Authentication token (required) |
| `--insecure-skip-verify` | `MCP_CLIENT_INSECURE_SKIP_VERIFY` | false | Skip TLS certificate verification |

## MCP Tools

The server provides the following tools:

### Cluster Management

- `get_cluster_status`: Get cluster status information (version, node count, namespace count)
- `list_nodes`: List all nodes in cluster

### Resource Management

- `list_pods`: List pods in a namespace
- `list_services`: List services in a namespace
- `list_deployments`: List deployments in a namespace

- `get_resource`: Get detailed information about a specific resource (JSON format). Secrets will be redacted.
- `get_resource_yaml`: Get full YAML definition of a resource. Secrets will be redacted.

### Observability & Debugging

- `get_events`: Get cluster events
- `get_pod_logs`: Get pod logs. Default tail_lines=100, max_bytes=1MB

### Security

- `check_rbac_permission`: Check if the current user has permission to perform an action (kubectl auth can-i)

## Security

- All operations are read-only by default
- Token-based authentication is required for all connections
- Secret data is automatically redacted when retrieved
- Supports RBAC permission validation
- Secure kubeconfig handling
- Connection timeout and retry mechanisms

## Integration

### Using as a Library

The `pkg/mcpclient` package can be imported and used in other Go applications:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AceDarkknight/k8s-mcp/pkg/mcpclient"
)

func main() {
    // Create configuration
    config := mcpclient.Config{
        ServerURL:          "https://localhost:8443",
        AuthToken:          "your-token",
        InsecureSkipVerify: true,
    }

    // Create client with optional custom headers
    client, err := mcpclient.NewClient(config,
        mcpclient.WithHeader("X-Custom-Header", "value"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Connect to server
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // List tools
    tools, err := client.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, tool := range tools {
        fmt.Printf("Tool: %s\n", tool.Name)
    }

    // Call a tool
    result, err := client.CallTool(ctx, "get_cluster_status", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Result: %v\n", result)
}
```

For more details, see [`pkg/mcpclient/README.md`](pkg/mcpclient/README.md).

### MCP Protocol Integration

k8s-mcp follows the standard MCP protocol and can be integrated into any MCP-compatible application:

1. **Claude Desktop**: AI assistant can view and analyze Kubernetes resources
2. **VS Code**: Get Kubernetes context via MCP extensions
3. **Custom Applications**: Use MCP client libraries to integrate

## Development

To add new features:

1. **New Tools**: Add new tool definitions and handler functions in `internal/mcp/server.go`
2. **New Resources**: Add new resource types in `internal/k8s/resources.go`

The project uses a modular design for easy extension and maintenance.
