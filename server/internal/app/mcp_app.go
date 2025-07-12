package app

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/mcp"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// MCPApplication represents the MCP server application
type MCPApplication struct {
	config    *config.Config
	logger    *logger.Logger
	
	// MCP-specific components
	vectorDB  vectordb.VectorDB
	llmClient llm.LLM
	journal   journal.Journal
	mcpServer *mcp.Server
}

// NewMCPApplication creates a new MCP application instance
func NewMCPApplication(cfg *config.Config, logger *logger.Logger) *MCPApplication {
	return &MCPApplication{
		config: cfg,
		logger: logger,
	}
}

// Name returns the application name
func (a *MCPApplication) Name() string {
	return "mcp-server"
}

// Initialize sets up all MCP components
func (a *MCPApplication) Initialize() error {
	var err error

	// Initialize VectorDB
	vdbConfig := &vectordb.Config{
		Provider:        a.config.VectorDB.Provider,
		URL:            a.config.VectorDB.URL,
		Insecure:       a.config.VectorDB.Insecure,
		VectorDimension: a.config.VectorDB.VectorDimension,
		OnDiskPayload:  a.config.VectorDB.OnDiskPayload,
		CollectionNames: a.config.VectorDB.CollectionNames,
	}
	a.vectorDB, err = vectordb.NewVectorDB(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	// Initialize LLM
	llmConfig := &llm.Config{
		Provider:           a.config.LLM.Provider,
		URL:               a.config.LLM.URL,
		EmbeddingModel:    a.config.LLM.EmbeddingModel,
		ConsolidationModel: a.config.LLM.ConsolidationModel,
		CacheEnabled:      a.config.LLM.CacheEnabled,
		Timeout:           a.config.LLM.Timeout,
		MaxRetries:        a.config.LLM.MaxRetries,
	}
	a.llmClient, err = llm.NewLLM(llmConfig)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Initialize Journal
	journalDeps := &journal.Dependencies{
		VectorDB:  a.vectorDB,
		LLMClient: a.llmClient,
		Config:    &a.config.Journal,
	}
	a.journal = journal.NewJournal(journalDeps)

	// Initialize MCP Server
	a.mcpServer = mcp.NewServer(a.config.MCP, a.journal, a.logger)

	a.logger.Info("MCP application initialized successfully")
	return nil
}

// Start begins running the MCP server
func (a *MCPApplication) Start(ctx context.Context) error {
	// Initialize vector database
	if err := a.vectorDB.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize vector database: %w", err)
	}

	// Start MCP server (blocking call for stdio communication)
	a.logger.Info("Starting MCP server for stdio communication")
	return a.mcpServer.ServeStdio(ctx)
}

// Stop gracefully shuts down the MCP server
func (a *MCPApplication) Stop(ctx context.Context) error {
	a.logger.Info("Stopping MCP application")

	// Clear LLM cache
	a.llmClient.ClearCache()

	a.logger.Info("MCP application stopped")
	return nil
}

// HealthCheck verifies all MCP components are healthy
func (a *MCPApplication) HealthCheck(ctx context.Context) error {
	// Check vector database
	if err := a.vectorDB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vector database unhealthy: %w", err)
	}

	// Check LLM client
	if err := a.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("LLM client unhealthy: %w", err)
	}

	return nil
}