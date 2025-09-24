package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// Transport interface defines how MCP messages are sent and received
type Transport interface {
	Send(message interface{}) error
	Receive() (*JSONRPCRequest, error)
	Close() error
}

// StdioTransport implements MCP transport over stdin/stdout
type StdioTransport struct {
	reader *bufio.Scanner
	writer io.Writer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewScanner(os.Stdin),
		writer: os.Stdout,
	}
}

// Send sends a message via stdout
func (t *StdioTransport) Send(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = fmt.Fprintln(t.writer, string(data))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Log to stderr for debugging (safe for stdio transport)
	log.Printf("Sent: %s", string(data))
	return nil
}

// Receive receives a message from stdin
func (t *StdioTransport) Receive() (*JSONRPCRequest, error) {
	if !t.reader.Scan() {
		if err := t.reader.Err(); err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		return nil, io.EOF
	}

	line := t.reader.Text()
	log.Printf("Received: %s", line)

	var request JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &request, nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	// Nothing to close for stdio
	return nil
}

// MessageHandler handles different types of MCP messages
type MessageHandler interface {
	HandleInitialize(req *InitializeRequest, id interface{}) (*InitializeResult, error)
	HandleListTools() (*ListToolsResult, error)
	HandleCallTool(req *CallToolRequest) (*CallToolResult, error)
	HandleListResources() (*ListResourcesResult, error)
	HandleReadResource(req *ReadResourceRequest) (*ReadResourceResult, error)
	HandleListPrompts() (*ListPromptsResult, error)
	HandleGetPrompt(req *GetPromptRequest) (*GetPromptResult, error)
}

// MessageDispatcher dispatches MCP messages to appropriate handlers
type MessageDispatcher struct {
	handler MessageHandler
}

// NewMessageDispatcher creates a new message dispatcher
func NewMessageDispatcher(handler MessageHandler) *MessageDispatcher {
	return &MessageDispatcher{
		handler: handler,
	}
}

// Dispatch processes a JSON-RPC request and returns a response
func (d *MessageDispatcher) Dispatch(request *JSONRPCRequest) interface{} {
	switch request.Method {
	case "initialize":
		return d.handleInitialize(request)
	case "tools/list":
		return d.handleListTools(request)
	case "tools/call":
		return d.handleCallTool(request)
	case "resources/list":
		return d.handleListResources(request)
	case "resources/read":
		return d.handleReadResource(request)
	case "prompts/list":
		return d.handleListPrompts(request)
	case "prompts/get":
		return d.handleGetPrompt(request)
	case "ping":
		return d.handlePing(request)
	case "notifications/initialized":
		// Handle initialized notification
		log.Println("Client initialized")
		return nil
	default:
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(MethodNotFound, "Method not found", request.Method))
	}
}

// handleInitialize processes initialize requests
func (d *MessageDispatcher) handleInitialize(request *JSONRPCRequest) interface{} {
	var initReq InitializeRequest
	if err := d.unmarshalParams(request.Params, &initReq); err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InvalidParams, "Invalid parameters", err.Error()))
	}

	result, err := d.handler.HandleInitialize(&initReq, request.ID)
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "Initialize failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleListTools processes tools/list requests
func (d *MessageDispatcher) handleListTools(request *JSONRPCRequest) interface{} {
	result, err := d.handler.HandleListTools()
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "List tools failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleCallTool processes tools/call requests
func (d *MessageDispatcher) handleCallTool(request *JSONRPCRequest) interface{} {
	var callReq CallToolRequest
	if err := d.unmarshalParams(request.Params, &callReq); err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InvalidParams, "Invalid parameters", err.Error()))
	}

	result, err := d.handler.HandleCallTool(&callReq)
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "Tool call failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleListResources processes resources/list requests
func (d *MessageDispatcher) handleListResources(request *JSONRPCRequest) interface{} {
	result, err := d.handler.HandleListResources()
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "List resources failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleReadResource processes resources/read requests
func (d *MessageDispatcher) handleReadResource(request *JSONRPCRequest) interface{} {
	var readReq ReadResourceRequest
	if err := d.unmarshalParams(request.Params, &readReq); err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InvalidParams, "Invalid parameters", err.Error()))
	}

	result, err := d.handler.HandleReadResource(&readReq)
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "Read resource failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleListPrompts processes prompts/list requests
func (d *MessageDispatcher) handleListPrompts(request *JSONRPCRequest) interface{} {
	result, err := d.handler.HandleListPrompts()
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "List prompts failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handleGetPrompt processes prompts/get requests
func (d *MessageDispatcher) handleGetPrompt(request *JSONRPCRequest) interface{} {
	var getReq GetPromptRequest
	if err := d.unmarshalParams(request.Params, &getReq); err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InvalidParams, "Invalid parameters", err.Error()))
	}

	result, err := d.handler.HandleGetPrompt(&getReq)
	if err != nil {
		return NewJSONRPCErrorResponse(request.ID, NewJSONRPCError(InternalError, "Get prompt failed", err.Error()))
	}

	return NewJSONRPCResponse(request.ID, result)
}

// handlePing processes ping requests
func (d *MessageDispatcher) handlePing(request *JSONRPCRequest) interface{} {
	// Return empty result for ping
	return NewJSONRPCResponse(request.ID, map[string]interface{}{})
}

// unmarshalParams unmarshals request parameters into a target struct
func (d *MessageDispatcher) unmarshalParams(params interface{}, target interface{}) error {
	if params == nil {
		return nil
	}

	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return nil
}
