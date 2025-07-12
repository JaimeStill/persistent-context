package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// Server represents an MCP server instance
type Server struct {
	name        string
	version     string
	tools       map[string]*Tool
	handlers    map[string]ToolHandler
	journal     journal.Journal
	pipeline    *ProcessingPipeline
	filter      *FilterEngine
	config      config.MCPConfig
	logger      *logger.Logger
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
func NewServer(cfg config.MCPConfig, journal journal.Journal, log *logger.Logger) *Server {
	// Get active profile
	profile := cfg.Profiles[cfg.CaptureMode]
	if profile == nil {
		log.Warn("Profile not found, using balanced default", "profile", cfg.CaptureMode)
		profile = cfg.Profiles["balanced"]
	}
	
	// Create filter engine
	filter := NewFilterEngine(cfg.FilterRules, profile)
	
	// Create processing pipeline
	pipeline := NewProcessingPipeline(cfg, filter, log)
	
	s := &Server{
		name:     cfg.Name,
		version:  cfg.Version,
		tools:    make(map[string]*Tool),
		handlers: make(map[string]ToolHandler),
		journal:  journal,
		pipeline: pipeline,
		filter:   filter,
		config:   cfg,
		logger:   log,
	}
	
	// Register enhanced tools
	s.registerTools()
	
	return s
}

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	s.registerCaptureEventTool()
	s.registerGetStatsTool()
	s.registerQueryMemoryTool()
	s.registerTriggerConsolidationTool()
}

// registerCaptureEventTool adds the enhanced event capture tool
func (s *Server) registerCaptureEventTool() {
	tool := &Tool{
		Name:        "capture_event",
		Description: "Capture an event through the intelligent filtering and processing pipeline",
		Parameters: map[string]Parameter{
			"type": {
				Type:        "string",
				Description: "Event type (file_read, file_write, command_output, search_results, etc.)",
				Required:    true,
			},
			"source": {
				Type:        "string",
				Description: "Source identifier (file path, command name, etc.)",
				Required:    true,
			},
			"content": {
				Type:        "string",
				Description: "The actual event content",
				Required:    true,
			},
			"metadata": {
				Type:        "object",
				Description: "Additional metadata about the event",
				Required:    false,
			},
		},
	}
	
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		eventTypeStr, ok := params["type"].(string)
		if !ok {
			return nil, fmt.Errorf("type parameter must be a string")
		}
		
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
		
		// Create capture event
		event := &types.CaptureEvent{
			Type:      types.EventType(eventTypeStr),
			Source:    source,
			Content:   content,
			Metadata:  metadata,
			Timestamp: time.Now(),
		}
		
		// Process through pipeline
		err := s.pipeline.ProcessEvent(event)
		if err != nil {
			return nil, fmt.Errorf("failed to process event: %w", err)
		}
		
		return map[string]any{
			"success":   true,
			"message":   "Event processed successfully",
			"type":      event.Type,
			"timestamp": event.Timestamp,
		}, nil
	}
	
	s.AddTool(tool, handler)
}

// registerGetStatsTool adds the statistics tool
func (s *Server) registerGetStatsTool() {
	tool := &Tool{
		Name:        "get_stats",
		Description: "Get performance statistics for the MCP pipeline",
		Parameters:  map[string]Parameter{},
	}
	
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		pipelineMetrics := s.pipeline.GetMetrics()
		filterStats := s.filter.GetStats()
		
		return map[string]any{
			"pipeline": pipelineMetrics,
			"filter":   filterStats,
			"config": map[string]any{
				"capture_mode":     s.config.CaptureMode,
				"worker_count":     s.config.WorkerCount,
				"batch_window_ms":  s.config.BatchWindowMs,
				"max_batch_size":   s.config.MaxBatchSize,
			},
		}, nil
	}
	
	s.AddTool(tool, handler)
}

// registerQueryMemoryTool adds the memory query tool
func (s *Server) registerQueryMemoryTool() {
	tool := &Tool{
		Name:        "query_memory",
		Description: "Query memories using vector similarity search",
		Parameters: map[string]Parameter{
			"query": {
				Type:        "string",
				Description: "Query text for similarity search",
				Required:    true,
			},
			"limit": {
				Type:        "number",
				Description: "Maximum number of results to return",
				Required:    false,
			},
		},
	}
	
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		query, ok := params["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query parameter must be a string")
		}
		
		limit := uint64(10) // default
		if limitParam, ok := params["limit"].(float64); ok {
			limit = uint64(limitParam)
		}
		
		// Use journal's vector search
		results, err := s.journal.QuerySimilarMemories(ctx, query, types.TypeEpisodic, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search memories: %w", err)
		}
		
		return map[string]any{
			"success": true,
			"results": results,
			"count":   len(results),
		}, nil
	}
	
	s.AddTool(tool, handler)
}

// registerTriggerConsolidationTool adds the consolidation trigger tool
func (s *Server) registerTriggerConsolidationTool() {
	tool := &Tool{
		Name:        "trigger_consolidation",
		Description: "Manually trigger memory consolidation process",
		Parameters:  map[string]Parameter{},
	}
	
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		// Get recent memories for consolidation
		memories, err := s.journal.GetMemories(ctx, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get memories for consolidation: %w", err)
		}
		
		// Trigger consolidation through journal
		err = s.journal.ConsolidateMemories(ctx, memories)
		if err != nil {
			return nil, fmt.Errorf("failed to trigger consolidation: %w", err)
		}
		
		return map[string]any{
			"success": true,
			"message": "Consolidation triggered successfully",
			"memories_processed": len(memories),
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

// Shutdown gracefully shuts down the MCP server
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down MCP server")
	
	// Shutdown the processing pipeline
	if s.pipeline != nil {
		err := s.pipeline.Shutdown()
		if err != nil {
			s.logger.Error("Error shutting down pipeline", "error", err)
			return err
		}
	}
	
	s.logger.Info("MCP server shut down successfully")
	return nil
}