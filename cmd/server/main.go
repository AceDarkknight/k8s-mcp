package main

import (
	"flag"
	"log"

	"k8s-mcp/internal/mcp"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "kubeconfig", "", "Path to kubeconfig file (optional)")
	flag.Parse()

	// Create MCP server
	server := mcp.NewServer()

	// Set up stdio transport
	transport := mcp.NewStdioTransport()
	server.SetTransport(transport)

	// Load kubeconfig if provided or use default
	if err := server.LoadKubeConfig(configPath); err != nil {
		log.Printf("Warning: Failed to load kubeconfig: %v", err)
		log.Println("Server will start but won't be able to connect to clusters until kubeconfig is properly configured")
	}

	// Run the server
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
