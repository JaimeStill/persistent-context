package app

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/pkg/logger"
	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents an MCP server instance that wraps the official SDK
type Server struct {
	mcpServer  *mcp.Server
	httpClient *Client
	config     *MCPConfig
	logger     *logger.Logger
}

// NewServer creates a new MCP server using the official SDK
func NewServer(cfg *MCPConfig, httpClient *Client, log *logger.Logger) *Server {
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
	// Essential Core Loop tools (5 total for MVP)
	s.registerCaptureMemoryTool()
	s.registerGetMemoriesTool()
	s.registerSearchMemoriesTool()
	s.registerTriggerConsolidationTool()
	s.registerGetStatsTool()
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
	Memories []*models.MemoryEntry  `json:"memories"`
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


// SearchMemoriesParams represents the search memories parameters
type SearchMemoriesParams struct {
	Content    string  `json:"content" mcp:"Query text for similarity search"`
	MemoryType *string `json:"memory_type,omitempty" mcp:"Type of memory to search (episodic, semantic, procedural)"`
	Limit      *uint64 `json:"limit,omitempty" mcp:"Maximum number of results to return"`
}

// SearchMemoriesResult represents the search memories result
type SearchMemoriesResult struct {
	Success bool                  `json:"success"`
	Results []*models.MemoryEntry  `json:"results"`
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

		memoryType := models.TypeEpisodic // default
		if args.MemoryType != nil {
			memoryType = models.MemoryType(*args.MemoryType)
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


