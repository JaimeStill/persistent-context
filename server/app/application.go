package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/consolidation"
	httpserver "github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/mcp"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
	"github.com/JaimeStill/persistent-context/internal/logger"
)

// Application represents the complete application with all components
type Application struct {
	config *config.Config
	logger *logger.Logger

	// Core components
	vectorDB        vectordb.VectorDB
	llmClient       llm.LLM
	journal         journal.Journal
	consolidation   *consolidation.Engine
	httpServer      *httpserver.Server
	mcpServer       *mcp.Server

	// Lifecycle
	started bool
}

// NewApplication creates a new application with all components properly initialized
func NewApplication(cfg *config.Config, logger *logger.Logger) (*Application, error) {
	app := &Application{
		config: cfg,
		logger: logger,
	}

	if err := app.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return app, nil
}

// initializeComponents initializes all components with proper dependency injection
func (a *Application) initializeComponents() error {
	var err error

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

	journalDeps := &journal.Dependencies{
		VectorDB:  a.vectorDB,
		LLMClient: a.llmClient,
		Config:    &a.config.Journal,
	}
	a.journal = journal.NewJournal(journalDeps)

	if a.config.Consolidation.Enabled {
		a.consolidation = consolidation.NewEngine(a.journal, a.llmClient, &a.config.Consolidation)
	}

	httpDeps := &httpserver.Dependencies{
		VectorDBHealth: &healthChecker{name: "vectordb", checker: a.vectorDB},
		LLMHealth:      &healthChecker{name: "llm", checker: a.llmClient},
	}
	a.httpServer = httpserver.NewServer(&a.config.HTTP, httpDeps)

	if a.config.MCP.Enabled {
		a.mcpServer = mcp.NewServer(a.config.MCP.Name, a.config.MCP.Version, a.journal)
	}

	return nil
}

// Start starts all components in the correct order
func (a *Application) Start(ctx context.Context) error {
	if a.started {
		return nil
	}

	a.logger.Info("Starting application components")

	if err := a.vectorDB.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize vector database: %w", err)
	}

	if a.consolidation != nil {
		if err := a.consolidation.Start(ctx); err != nil {
			return fmt.Errorf("failed to start consolidation engine: %w", err)
		}
		a.logger.Info("Consolidation engine started")
	}

	go func() {
		if err := a.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()
	a.logger.Info("HTTP server started", "port", a.config.HTTP.Port)

	a.started = true
	a.logger.Info("All application components started successfully")
	return nil
}

// Stop gracefully stops all components
func (a *Application) Stop(ctx context.Context) error {
	if !a.started {
		return nil
	}

	a.logger.Info("Stopping application components")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("Error stopping HTTP server", "error", err)
	}

	if a.consolidation != nil {
		a.consolidation.Stop()
		a.logger.Info("Consolidation engine stopped")
	}

	a.llmClient.ClearCache()

	a.started = false
	a.logger.Info("All application components stopped")
	return nil
}

// HealthCheck returns the health status of all components
func (a *Application) HealthCheck(ctx context.Context) error {
	if !a.started {
		return fmt.Errorf("application not started")
	}

	// Check core components
	if err := a.vectorDB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vector database unhealthy: %w", err)
	}

	if err := a.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("LLM client unhealthy: %w", err)
	}

	return nil
}

// GetJournal returns the journal for external use
func (a *Application) GetJournal() journal.Journal {
	return a.journal
}

// GetConsolidationEngine returns the consolidation engine for external use
func (a *Application) GetConsolidationEngine() *consolidation.Engine {
	return a.consolidation
}

// healthChecker is a simple adapter for the HTTP server health check interface
type healthChecker struct {
	name    string
	checker interface{ HealthCheck(context.Context) error }
}

func (h *healthChecker) Name() string {
	return h.name
}

func (h *healthChecker) HealthCheck(ctx context.Context) error {
	return h.checker.HealthCheck(ctx)
}

func (h *healthChecker) IsInitialized() bool {
	return true
}

func (h *healthChecker) IsRunning() bool {
	return true
}