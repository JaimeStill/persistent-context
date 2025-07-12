package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/JaimeStill/persistent-context/internal/journal"
)

// Server represents an MCP server instance
type Server struct {
	name        string
	version     string
	tools       map[string]*Tool
	handlers    map[string]ToolHandler
	journal journal.Journal
}

// ToolHandler is a function that handles tool invocations
type ToolHandler func(ctx context.Context, params map[string]any) (any, error)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]Parameter   `json:"parameters"`
}

// Parameter represents a tool parameter
type Parameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// NewServer creates a new MCP server
func NewServer(name, version string, journal journal.Journal) *Server {
	s := &Server{
		name:        name,
		version:     version,
		tools:       make(map[string]*Tool),
		handlers:    make(map[string]ToolHandler),
		journal: journal,
	}
	
	// Register default capture_context tool
	s.registerCaptureContextTool()
	
	return s
}

// registerCaptureContextTool adds the context capture tool
func (s *Server) registerCaptureContextTool() {
	tool := &Tool{
		Name:        "capture_context",
		Description: "Capture context from the environment for memory consolidation",
		Parameters: map[string]Parameter{
			"source": {
				Type:        "string",
				Description: "Source of the context (e.g., 'file_edit', 'command_output')",
				Required:    true,
			},
			"content": {
				Type:        "string",
				Description: "The actual context content to capture",
				Required:    true,
			},
			"metadata": {
				Type:        "object",
				Description: "Additional metadata about the context",
				Required:    false,
			},
		},
	}
	
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		source, ok := params["source"].(string)
		if !ok {
			return nil, fmt.Errorf("source parameter must be a string")
		}
		
		content, ok := params["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content parameter must be a string")
		}
		
		metadata, _ := params["metadata"].(map[string]any)
		if metadata == nil {
			metadata = make(map[string]any)
		}
		
		err := s.journal.CaptureContext(ctx, source, content, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to capture context: %w", err)
		}
		
		return map[string]any{
			"success": true,
			"message": "Context captured successfully",
		}, nil
	}
	
	s.AddTool(tool, handler)
}

// AddTool registers a new tool with its handler
func (s *Server) AddTool(tool *Tool, handler ToolHandler) {
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

// ServeStdio starts the MCP server using stdio for communication
func (s *Server) ServeStdio(ctx context.Context) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var request Request
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to decode request: %w", err)
			}
			
			response, err := s.handleRequest(ctx, &request)
			if err != nil {
				response = &Response{
					ID:    request.ID,
					Error: &Error{Message: err.Error()},
				}
			}
			
			if err := encoder.Encode(response); err != nil {
				return fmt.Errorf("failed to encode response: %w", err)
			}
		}
	}
}

// handleRequest processes incoming MCP requests
func (s *Server) handleRequest(ctx context.Context, req *Request) (*Response, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(ctx, req)
	default:
		return nil, fmt.Errorf("unknown method: %s", req.Method)
	}
}

// Request represents an MCP request
type Request struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// Response represents an MCP response
type Response struct {
	ID     string `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *Request) (*Response, error) {
	return &Response{
		ID: req.ID,
		Result: map[string]any{
			"name":    s.name,
			"version": s.version,
		},
	}, nil
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList(req *Request) (*Response, error) {
	tools := make([]any, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	
	return &Response{
		ID: req.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}, nil
}

// handleToolCall executes a tool
func (s *Server) handleToolCall(ctx context.Context, req *Request) (*Response, error) {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name not provided")
	}
	
	handler, exists := s.handlers[toolName]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	
	arguments, _ := req.Params["arguments"].(map[string]any)
	if arguments == nil {
		arguments = make(map[string]any)
	}
	
	result, err := handler(ctx, arguments)
	if err != nil {
		return nil, err
	}
	
	return &Response{
		ID:     req.ID,
		Result: result,
	}, nil
}