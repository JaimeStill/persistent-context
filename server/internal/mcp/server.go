package mcp

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents an MCP server instance that wraps the official SDK
type Server struct {
	mcpServer  *mcp.Server
	httpClient *http.Client
	config     config.MCPConfig
	logger     *logger.Logger
}

// NewServer creates a new MCP server using the official SDK
func NewServer(cfg config.MCPConfig, httpClient *http.Client, log *logger.Logger) *Server {
	// Create the official SDK server
	impl := &mcp.Implementation{
		Name:    cfg.Name,
		Version: cfg.Version,
	}
	
	mcpServer := mcp.NewServer(impl, nil)
	
	s := &Server{
		mcpServer:  mcpServer,
		httpClient: httpClient,
		config:     cfg,
		logger:     log,
	}

	// Register all our tools
	s.registerTools()

	return s
}

// ServeStdio starts the MCP server using stdio for communication
func (s *Server) ServeStdio(ctx context.Context) error {
	transport := mcp.NewStdioTransport()
	return s.mcpServer.Run(ctx, transport)
}

// Shutdown gracefully shuts down the MCP server
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down MCP server")
	s.logger.Info("MCP server shut down successfully")
	return nil
}

// registerTools registers all MCP tools using the official SDK
func (s *Server) registerTools() {
	// High-level tools (combine multiple operations)
	s.registerCaptureEventTool()
	s.registerGetStatsTool()
	s.registerQueryMemoryTool()
	s.registerTriggerConsolidationTool()
	
	// Direct HTTP API tools
	s.registerCaptureMemoryTool()
	s.registerGetMemoriesTool()
	s.registerGetMemoryByIDTool()
	s.registerSearchMemoriesTool()
	s.registerConsolidateMemoriesTool()
	s.registerGetMemoryStatsTool()
}

// Tool parameter types
type CaptureEventParams struct {
	Type     string         `json:"type" mcp:"Event type (file_read, file_write, command_output, search_results, etc.)"`
	Source   string         `json:"source" mcp:"Source identifier (file path, command name, etc.)"`
	Content  string         `json:"content" mcp:"The actual event content"`
	Metadata map[string]any `json:"metadata,omitempty" mcp:"Additional metadata about the event"`
}

type CaptureEventResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

// registerCaptureEventTool adds the enhanced event capture tool
func (s *Server) registerCaptureEventTool() {
	tool := &mcp.Tool{
		Name:        "capture_event",
		Description: "Capture an event through the intelligent filtering and processing pipeline",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CaptureEventParams]) (*mcp.CallToolResultFor[CaptureEventResult], error) {
		args := params.Arguments

		metadata := args.Metadata
		if metadata == nil {
			metadata = make(map[string]any)
		}

		// Add event type to metadata for server-side processing
		metadata["event_type"] = args.Type

		// Capture via HTTP client
		entry, err := s.httpClient.CaptureContext(ctx, args.Source, args.Content, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to capture memory: %w", err)
		}

		result := CaptureEventResult{
			Success: true,
			Message: "Memory captured successfully",
			ID:      entry.ID,
		}

		return &mcp.CallToolResultFor[CaptureEventResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully captured event: %s", args.Type),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// GetStatsResult represents the statistics result
type GetStatsResult struct {
	Success bool           `json:"success"`
	Stats   map[string]any `json:"stats"`
}

// registerGetStatsTool adds the statistics tool
func (s *Server) registerGetStatsTool() {
	tool := &mcp.Tool{
		Name:        "get_stats",
		Description: "Get memory statistics from the persistent context service",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[GetStatsResult], error) {
		stats, err := s.httpClient.GetMemoryStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get memory stats: %w", err)
		}

		result := GetStatsResult{
			Success: true,
			Stats:   stats,
		}

		totalMemories := 0
		if total, ok := stats["total_memories"].(float64); ok {
			totalMemories = int(total)
		}

		return &mcp.CallToolResultFor[GetStatsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Memory stats: %d total memories", totalMemories),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// QueryMemoryParams represents the query memory parameters
type QueryMemoryParams struct {
	Query string  `json:"query" mcp:"Query text for similarity search"`
	Limit *uint64 `json:"limit,omitempty" mcp:"Maximum number of results to return"`
}

// QueryMemoryResult represents the query memory result
type QueryMemoryResult struct {
	Success bool                  `json:"success"`
	Results []*types.MemoryEntry  `json:"results"`
	Count   int                   `json:"count"`
}

// registerQueryMemoryTool adds the memory query tool
func (s *Server) registerQueryMemoryTool() {
	tool := &mcp.Tool{
		Name:        "query_memory",
		Description: "Query memories using vector similarity search",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[QueryMemoryParams]) (*mcp.CallToolResultFor[QueryMemoryResult], error) {
		args := params.Arguments

		limit := uint64(10) // default
		if args.Limit != nil {
			limit = *args.Limit
		}

		// Use HTTP client's vector search
		results, err := s.httpClient.QuerySimilarMemories(ctx, args.Query, types.TypeEpisodic, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search memories: %w", err)
		}

		result := QueryMemoryResult{
			Success: true,
			Results: results,
			Count:   len(results),
		}

		return &mcp.CallToolResultFor[QueryMemoryResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Found %d memories matching query", len(results)),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// TriggerConsolidationResult represents the consolidation result
type TriggerConsolidationResult struct {
	Success           bool `json:"success"`
	Message           string `json:"message"`
	MemoriesProcessed int  `json:"memories_processed"`
}

// registerTriggerConsolidationTool adds the consolidation trigger tool
func (s *Server) registerTriggerConsolidationTool() {
	tool := &mcp.Tool{
		Name:        "trigger_consolidation",
		Description: "Manually trigger memory consolidation process",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[TriggerConsolidationResult], error) {
		// Get recent memories for consolidation
		memories, err := s.httpClient.GetMemories(ctx, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get memories for consolidation: %w", err)
		}

		// Trigger consolidation through HTTP client
		err = s.httpClient.ConsolidateMemories(ctx, memories)
		if err != nil {
			return nil, fmt.Errorf("failed to trigger consolidation: %w", err)
		}

		result := TriggerConsolidationResult{
			Success:           true,
			Message:           "Consolidation triggered successfully",
			MemoriesProcessed: len(memories),
		}

		return &mcp.CallToolResultFor[TriggerConsolidationResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully triggered consolidation for %d memories", len(memories)),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// CaptureMemoryParams represents the capture memory parameters
type CaptureMemoryParams struct {
	Source   string         `json:"source" mcp:"Source identifier for the memory"`
	Content  string         `json:"content" mcp:"The content to store in memory"`
	Metadata map[string]any `json:"metadata,omitempty" mcp:"Additional metadata for the memory"`
}

// CaptureMemoryResult represents the capture memory result
type CaptureMemoryResult struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Message string `json:"message"`
}

// registerCaptureMemoryTool adds the memory capture tool via HTTP API
func (s *Server) registerCaptureMemoryTool() {
	tool := &mcp.Tool{
		Name:        "capture_memory",
		Description: "Capture a memory entry via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CaptureMemoryParams]) (*mcp.CallToolResultFor[CaptureMemoryResult], error) {
		args := params.Arguments

		// Capture memory via HTTP API
		entry, err := s.httpClient.CaptureContext(ctx, args.Source, args.Content, args.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to capture memory: %w", err)
		}

		result := CaptureMemoryResult{
			Success: true,
			ID:      entry.ID,
			Message: "Memory captured successfully",
		}

		return &mcp.CallToolResultFor[CaptureMemoryResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "Memory captured successfully",
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// GetMemoriesParams represents the get memories parameters
type GetMemoriesParams struct {
	Limit *uint64 `json:"limit,omitempty" mcp:"Maximum number of memories to retrieve"`
}

// GetMemoriesResult represents the get memories result
type GetMemoriesResult struct {
	Success  bool                  `json:"success"`
	Memories []*types.MemoryEntry  `json:"memories"`
	Count    int                   `json:"count"`
}

// registerGetMemoriesTool adds the get memories tool via HTTP API
func (s *Server) registerGetMemoriesTool() {
	tool := &mcp.Tool{
		Name:        "get_memories",
		Description: "Retrieve recent memories via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetMemoriesParams]) (*mcp.CallToolResultFor[GetMemoriesResult], error) {
		args := params.Arguments

		limit := uint64(100) // default
		if args.Limit != nil {
			limit = *args.Limit
		}

		// Get memories via HTTP API
		memories, err := s.httpClient.GetMemories(ctx, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get memories: %w", err)
		}

		result := GetMemoriesResult{
			Success:  true,
			Memories: memories,
			Count:    len(memories),
		}

		return &mcp.CallToolResultFor[GetMemoriesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Retrieved %d memories", len(memories)),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// GetMemoryByIDParams represents the get memory by ID parameters
type GetMemoryByIDParams struct {
	ID string `json:"id" mcp:"The ID of the memory to retrieve"`
}

// GetMemoryByIDResult represents the get memory by ID result
type GetMemoryByIDResult struct {
	Success bool                `json:"success"`
	Memory  *types.MemoryEntry  `json:"memory"`
}

// registerGetMemoryByIDTool adds the get memory by ID tool via HTTP API
func (s *Server) registerGetMemoryByIDTool() {
	tool := &mcp.Tool{
		Name:        "get_memory_by_id",
		Description: "Retrieve a specific memory by ID via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetMemoryByIDParams]) (*mcp.CallToolResultFor[GetMemoryByIDResult], error) {
		args := params.Arguments

		// Get memory via HTTP API
		memory, err := s.httpClient.GetMemoryByID(ctx, args.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get memory: %w", err)
		}

		result := GetMemoryByIDResult{
			Success: true,
			Memory:  memory,
		}

		return &mcp.CallToolResultFor[GetMemoryByIDResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Retrieved memory: %s", memory.ID),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// SearchMemoriesParams represents the search memories parameters
type SearchMemoriesParams struct {
	Content    string  `json:"content" mcp:"Query text for similarity search"`
	MemoryType *string `json:"memory_type,omitempty" mcp:"Type of memory to search (episodic, semantic, procedural)"`
	Limit      *uint64 `json:"limit,omitempty" mcp:"Maximum number of results to return"`
}

// SearchMemoriesResult represents the search memories result
type SearchMemoriesResult struct {
	Success bool                  `json:"success"`
	Results []*types.MemoryEntry  `json:"results"`
	Count   int                   `json:"count"`
}

// registerSearchMemoriesTool adds the search memories tool via HTTP API
func (s *Server) registerSearchMemoriesTool() {
	tool := &mcp.Tool{
		Name:        "search_memories",
		Description: "Search memories by content similarity via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchMemoriesParams]) (*mcp.CallToolResultFor[SearchMemoriesResult], error) {
		args := params.Arguments

		memoryType := types.TypeEpisodic // default
		if args.MemoryType != nil {
			memoryType = types.MemoryType(*args.MemoryType)
		}

		limit := uint64(10) // default
		if args.Limit != nil {
			limit = *args.Limit
		}

		// Search memories via HTTP API
		results, err := s.httpClient.QuerySimilarMemories(ctx, args.Content, memoryType, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search memories: %w", err)
		}

		result := SearchMemoriesResult{
			Success: true,
			Results: results,
			Count:   len(results),
		}

		return &mcp.CallToolResultFor[SearchMemoriesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Found %d memories", len(results)),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// ConsolidateMemoriesParams represents the consolidate memories parameters
type ConsolidateMemoriesParams struct {
	MemoryIDs []string `json:"memory_ids" mcp:"Array of memory IDs to consolidate"`
}

// ConsolidateMemoriesResult represents the consolidate memories result
type ConsolidateMemoriesResult struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	MemoriesProcessed int    `json:"memories_processed"`
}

// registerConsolidateMemoriesTool adds the consolidate memories tool via HTTP API
func (s *Server) registerConsolidateMemoriesTool() {
	tool := &mcp.Tool{
		Name:        "consolidate_memories",
		Description: "Consolidate specified memories via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ConsolidateMemoriesParams]) (*mcp.CallToolResultFor[ConsolidateMemoriesResult], error) {
		args := params.Arguments

		if len(args.MemoryIDs) == 0 {
			return nil, fmt.Errorf("memory_ids cannot be empty")
		}

		// Get memories by IDs
		var memories []*types.MemoryEntry
		for _, id := range args.MemoryIDs {
			memory, err := s.httpClient.GetMemoryByID(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("failed to get memory %s: %w", id, err)
			}
			memories = append(memories, memory)
		}

		// Consolidate via HTTP API
		err := s.httpClient.ConsolidateMemories(ctx, memories)
		if err != nil {
			return nil, fmt.Errorf("failed to consolidate memories: %w", err)
		}

		result := ConsolidateMemoriesResult{
			Success:           true,
			Message:           "Memories consolidated successfully",
			MemoriesProcessed: len(memories),
		}

		return &mcp.CallToolResultFor[ConsolidateMemoriesResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully consolidated %d memories", len(memories)),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}

// registerGetMemoryStatsTool adds the get memory stats tool via HTTP API
func (s *Server) registerGetMemoryStatsTool() {
	tool := &mcp.Tool{
		Name:        "get_memory_stats",
		Description: "Get memory statistics via the HTTP API",
	}

	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[GetStatsResult], error) {
		// Get stats via HTTP API
		stats, err := s.httpClient.GetMemoryStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get memory stats: %w", err)
		}

		result := GetStatsResult{
			Success: true,
			Stats:   stats,
		}

		totalMemories := 0
		if total, ok := stats["total_memories"].(float64); ok {
			totalMemories = int(total)
		}

		return &mcp.CallToolResultFor[GetStatsResult]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Memory statistics retrieved: %d total memories", totalMemories),
			}},
			StructuredContent: result,
		}, nil
	}

	mcp.AddTool(s.mcpServer, tool, handler)
}