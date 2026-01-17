#!/bin/bash
# Test script for k8s-mcp on Linux

echo "Building k8s-mcp..."
go build -o bin/k8s-mcp-server ./cmd/server
go build -o bin/k8s-mcp-client ./cmd/client

if [ ! -f bin/k8s-mcp-server ]; then
    echo "Failed to build server"
    exit 1
fi

if [ ! -f bin/k8s-mcp-client ]; then
    echo "Failed to build client"
    exit 1
fi

echo "Build successful!"
echo

echo "Starting interactive test..."
echo "Type 'help' for available commands in the client"
echo

./bin/k8s-mcp-client ./bin/k8s-mcp-server