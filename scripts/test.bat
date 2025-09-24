@echo off
REM Test script for k8s-mcp on Windows

echo Building k8s-mcp...
go build -o bin\k8s-mcp-server.exe .\cmd\server
go build -o bin\k8s-mcp-client.exe .\cmd\client

if not exist bin\k8s-mcp-server.exe (
    echo Failed to build server
    exit /b 1
)

if not exist bin\k8s-mcp-client.exe (
    echo Failed to build client
    exit /b 1
)

echo Build successful!
echo.

echo Starting interactive test...
echo Type 'help' for available commands in the client
echo.

bin\k8s-mcp-client.exe bin\k8s-mcp-server.exe
