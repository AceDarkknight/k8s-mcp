package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"k8s-mcp/internal/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: k8s-mcp-client <server-executable>")
		os.Exit(1)
	}

	serverPath := os.Args[1]

	// Start the MCP server as subprocess
	cmd := exec.Command(serverPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("Server: %s", scanner.Text())
		}
	}()

	// Initialize client
	client := &MCPClient{
		stdin:  stdin,
		stdout: stdout,
		id:     1,
	}

	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Initialize connection
	if err := client.Initialize(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Send initialized notification
	if err := client.SendNotification("notifications/initialized", nil); err != nil {
		log.Fatalf("Failed to send initialized notification: %v", err)
	}

	// Interactive loop
	fmt.Println("k8s MCP Client - Type 'help' for available commands, 'quit' to exit")

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

		if err := client.HandleInput(input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// MCPClient represents a simple MCP client
type MCPClient struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	id     int
}

// Initialize sends the initialize request
func (c *MCPClient) Initialize() error {
	initReq := mcp.InitializeRequest{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities:    mcp.ClientCapabilities{
			// No client capabilities for now
		},
		ClientInfo: mcp.Implementation{
			Name:    "k8s-mcp-client",
			Title:   "Kubernetes MCP Test Client",
			Version: "1.0.0",
		},
	}

	response, err := c.SendRequest("initialize", initReq)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("initialize error: %s", response.Error.Message)
	}

	var result mcp.InitializeResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	fmt.Printf("Connected to: %s v%s\n", result.ServerInfo.Title, result.ServerInfo.Version)
	if result.Instructions != "" {
		fmt.Printf("Instructions: %s\n", result.Instructions)
	}

	return nil
}

// SendRequest sends a JSON-RPC request and waits for response
func (c *MCPClient) SendRequest(method string, params interface{}) (*mcp.JSONRPCResponse, error) {
	request := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.id,
		Method:  method,
		Params:  params,
	}
	c.id++

	// Send request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := fmt.Fprintln(c.stdin, string(data)); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(c.stdout)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read response")
	}

	var response mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(scanner.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// SendNotification sends a JSON-RPC notification
func (c *MCPClient) SendNotification(method string, params interface{}) error {
	notification := mcp.JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if _, err := fmt.Fprintln(c.stdin, string(data)); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// HandleInput processes user input commands
func (c *MCPClient) HandleInput(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "help":
		c.showHelp()
		return nil
	case "tools":
		return c.listTools()
	case "resources":
		return c.listResources()
	case "prompts":
		return c.listPrompts()
	case "call":
		if len(parts) < 2 {
			fmt.Println("Usage: call <tool_name> [args...]")
			return nil
		}
		return c.callTool(parts[1], parts[2:])
	case "read":
		if len(parts) < 2 {
			fmt.Println("Usage: read <resource_uri>")
			return nil
		}
		return c.readResource(parts[1])
	case "prompt":
		if len(parts) < 2 {
			fmt.Println("Usage: prompt <prompt_name> [args...]")
			return nil
		}
		return c.getPrompt(parts[1], parts[2:])
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		return nil
	}
}

func (c *MCPClient) showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                     - Show this help")
	fmt.Println("  tools                    - List available tools")
	fmt.Println("  resources                - List available resources")
	fmt.Println("  prompts                  - List available prompts")
	fmt.Println("  call <tool> [args...]    - Call a tool")
	fmt.Println("  read <uri>               - Read a resource")
	fmt.Println("  prompt <name> [args...]  - Get a prompt")
	fmt.Println("  quit                     - Exit the client")
	fmt.Println()
	fmt.Println("Example tool calls:")
	fmt.Println("  call list_clusters")
	fmt.Println("  call list_namespaces")
	fmt.Println("  call list_resources pods default")
	fmt.Println("  call get_resource pods my-pod default")
}

func (c *MCPClient) listTools() error {
	response, err := c.SendRequest("tools/list", nil)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("tools/list error: %s", response.Error.Message)
	}

	var result mcp.ListToolsResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	fmt.Println("Available tools:")
	for _, tool := range result.Tools {
		fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
	}

	return nil
}

func (c *MCPClient) listResources() error {
	response, err := c.SendRequest("resources/list", nil)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("resources/list error: %s", response.Error.Message)
	}

	var result mcp.ListResourcesResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	fmt.Println("Available resources:")
	for _, resource := range result.Resources {
		fmt.Printf("  %s - %s\n", resource.URI, resource.Description)
	}

	return nil
}

func (c *MCPClient) listPrompts() error {
	response, err := c.SendRequest("prompts/list", nil)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("prompts/list error: %s", response.Error.Message)
	}

	var result mcp.ListPromptsResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	fmt.Println("Available prompts:")
	for _, prompt := range result.Prompts {
		fmt.Printf("  %s - %s\n", prompt.Name, prompt.Description)
		if len(prompt.Arguments) > 0 {
			fmt.Printf("    Arguments: ")
			for i, arg := range prompt.Arguments {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", arg.Name)
				if arg.Required {
					fmt.Printf("*")
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func (c *MCPClient) callTool(toolName string, args []string) error {
	// Parse simple arguments (key=value format)
	arguments := make(map[string]interface{})
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			arguments[parts[0]] = parts[1]
		}
	}

	toolReq := mcp.CallToolRequest{
		Name:      toolName,
		Arguments: arguments,
	}

	response, err := c.SendRequest("tools/call", toolReq)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("tool call error: %s", response.Error.Message)
	}

	var result mcp.CallToolResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	if result.IsError {
		fmt.Println("Tool execution error:")
	}

	for _, content := range result.Content {
		if textContent, ok := content.(map[string]interface{}); ok {
			if textContent["type"] == "text" {
				fmt.Println(textContent["text"])
			}
		}
	}

	return nil
}

func (c *MCPClient) readResource(uri string) error {
	readReq := mcp.ReadResourceRequest{
		URI: uri,
	}

	response, err := c.SendRequest("resources/read", readReq)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("resource read error: %s", response.Error.Message)
	}

	var result mcp.ReadResourceResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	for _, content := range result.Contents {
		fmt.Printf("Resource: %s\n", content.URI)
		if content.Text != "" {
			fmt.Println(content.Text)
		} else if content.Blob != "" {
			fmt.Printf("Binary content (%s)\n", content.MimeType)
		}
		fmt.Println()
	}

	return nil
}

func (c *MCPClient) getPrompt(promptName string, args []string) error {
	// Parse simple arguments (key=value format)
	arguments := make(map[string]string)
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			arguments[parts[0]] = parts[1]
		}
	}

	promptReq := mcp.GetPromptRequest{
		Name:      promptName,
		Arguments: arguments,
	}

	response, err := c.SendRequest("prompts/get", promptReq)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("prompt get error: %s", response.Error.Message)
	}

	var result mcp.GetPromptResult
	if err := c.unmarshalResult(response.Result, &result); err != nil {
		return err
	}

	fmt.Printf("Prompt: %s\n", promptName)
	if result.Description != "" {
		fmt.Printf("Description: %s\n", result.Description)
	}

	for _, message := range result.Messages {
		fmt.Printf("\n[%s]: ", message.Role)
		if textContent, ok := message.Content.(map[string]interface{}); ok {
			if textContent["type"] == "text" {
				fmt.Println(textContent["text"])
			}
		}
	}

	return nil
}

// unmarshalResult unmarshals a JSON result into target struct
func (c *MCPClient) unmarshalResult(result interface{}, target interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}
