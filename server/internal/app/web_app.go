package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/consolidation"
	httpserver "github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// WebApplication represents the web server application
type WebApplication struct {
	config    *config.Config
	logger    *logger.Logger
	
	// Web-specific components
	vectorDB      vectordb.VectorDB
	llmClient     llm.LLM
	journal       journal.Journal
	consolidation *consolidation.Engine
	httpServer    *httpserver.Server
}

// NewWebApplication creates a new web application instance
func NewWebApplication(cfg *config.Config, logger *logger.Logger) *WebApplication {
	return &WebApplication{
		config: cfg,
		logger: logger,
	}
}

// Name returns the application name
func (a *WebApplication) Name() string {
	return "web-server"
}

// Initialize sets up all web components
func (a *WebApplication) Initialize() error {
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
		VectorDB:            a.vectorDB,
		LLMClient:           a.llmClient,
		Config:              &a.config.Journal,
		ConsolidationConfig: &a.config.Consolidation,
	}
	
	// Validate dependencies before creating journal
	if err := journalDeps.Validate(); err != nil {
		return fmt.Errorf("journal dependencies validation failed: %w", err)
	}
	
	a.journal = journal.NewJournal(journalDeps)

	// Initialize Consolidation Engine
	a.consolidation = consolidation.NewEngine(a.journal, a.llmClient, &a.config.Consolidation)

	// Initialize HTTP Server
	httpDeps := &httpserver.Dependencies{
		VectorDBHealth: NewHealthChecker("vectordb", a.vectorDB),
		LLMHealth:      NewHealthChecker("llm", a.llmClient),
		Journal:        a.journal,
	}
	a.httpServer = httpserver.NewServer(&a.config.HTTP, httpDeps)

	a.logger.Info("Web application initialized successfully")
	return nil
}

// Start begins running the web server
func (a *WebApplication) Start(ctx context.Context) error {
	// Initialize vector database
	if err := a.vectorDB.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize vector database: %w", err)
	}

	// Start consolidation engine
	if err := a.consolidation.Start(ctx); err != nil {
		return fmt.Errorf("failed to start consolidation engine: %w", err)
	}
	a.logger.Info("Consolidation engine started")

	// Start HTTP server in background
	go func() {
		if err := a.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()
	
	a.logger.Info("HTTP server started", "port", a.config.HTTP.Port)
	return nil
}

// Stop gracefully shuts down the web server
func (a *WebApplication) Stop(ctx context.Context) error {
	a.logger.Info("Stopping web application")

	// Stop HTTP server
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("Error stopping HTTP server", "error", err)
	}

	// Stop consolidation engine
	a.consolidation.Stop()
	a.logger.Info("Consolidation engine stopped")

	// Clear LLM cache
	a.llmClient.ClearCache()

	a.logger.Info("Web application stopped")
	return nil
}

// HealthCheck verifies all web components are healthy
func (a *WebApplication) HealthCheck(ctx context.Context) error {
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