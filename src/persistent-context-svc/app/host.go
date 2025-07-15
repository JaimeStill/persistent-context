package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/journal"
	"github.com/JaimeStill/persistent-context/pkg/llm"
	"github.com/JaimeStill/persistent-context/pkg/logger"
	"github.com/JaimeStill/persistent-context/pkg/memory"
	"github.com/JaimeStill/persistent-context/pkg/vectordb"
)

// Host represents the web service host that manages all components
type Host struct {
	config *Config
	logger *logger.Logger

	// Service components
	vectorDB        vectordb.VectorDB
	llmClient       llm.LLM
	journal         journal.Journal
	memoryProcessor *memory.Processor
	httpServer      *http.Server
}

// NewHost creates a new host instance
func NewHost(cfg *Config, logger *logger.Logger) *Host {
	return &Host{
		config: cfg,
		logger: logger,
	}
}

// Start initializes and starts all components
func (h *Host) Start(ctx context.Context) error {
	// Initialize all components
	if err := h.initialize(); err != nil {
		return fmt.Errorf("failed to initialize host: %w", err)
	}

	// Start HTTP server
	if err := h.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start memory processor
	if err := h.memoryProcessor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start memory processor: %w", err)
	}

	h.logger.Info("Host started successfully")
	return nil
}

// Stop gracefully shuts down all components
func (h *Host) Stop(ctx context.Context) error {
	h.logger.Info("Shutting down host...")

	// Stop memory processor
	if h.memoryProcessor != nil {
		h.memoryProcessor.Stop()
	}

	// Stop HTTP server
	if h.httpServer != nil {
		if err := h.httpServer.Shutdown(ctx); err != nil {
			h.logger.Error("Error shutting down HTTP server", "error", err)
		}
	}

	h.logger.Info("Host stopped successfully")
	return nil
}

// HealthCheck verifies all components are healthy
func (h *Host) HealthCheck(ctx context.Context) error {
	// Check VectorDB health
	if err := h.vectorDB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vectordb health check failed: %w", err)
	}

	// Check LLM health
	if err := h.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("llm health check failed: %w", err)
	}

	return nil
}

// initialize sets up all components
func (h *Host) initialize() error {
	var err error

	// Initialize VectorDB
	h.vectorDB, err = vectordb.NewVectorDB(&h.config.VectorDB)
	if err != nil {
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	// Initialize VectorDB collections
	if err := h.vectorDB.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize vector database: %w", err)
	}

	// Initialize LLM client
	h.llmClient, err = llm.NewLLM(&h.config.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Initialize journal
	journalDeps := &journal.Dependencies{
		VectorDB:       h.vectorDB,
		LLMClient:      h.llmClient,
		Config:         &h.config.Journal,
		MemoryConfig:   &h.config.Memory,
		VectorDBConfig: &h.config.VectorDB,
	}

	if err := journalDeps.Validate(); err != nil {
		return fmt.Errorf("invalid journal dependencies: %w", err)
	}

	h.journal = journal.NewJournal(journalDeps)

	// Initialize memory processor
	h.memoryProcessor = memory.NewProcessor(h.journal, h.llmClient, &h.config.Memory)

	return nil
}

// startHTTPServer starts the HTTP server
func (h *Host) startHTTPServer() error {
	// Create server dependencies
	deps := &Dependencies{
		VectorDBHealth: h.vectorDB,
		LLMHealth:      h.llmClient,
		Journal:        h.journal,
		VectorDB:       h.vectorDB,
	}

	// Create HTTP server using the server.go implementation
	server := NewServer(h.config, deps)

	// Start server in goroutine
	go func() {
		h.logger.Info("Starting HTTP server", "port", h.config.HTTP.Port)
		if err := server.Start(); err != nil {
			h.logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}

// Run provides the main execution loop with graceful shutdown
func (h *Host) Run() error {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the host
	if err := h.Start(ctx); err != nil {
		return fmt.Errorf("failed to start host: %w", err)
	}

	// Perform initial health check
	if err := h.HealthCheck(ctx); err != nil {
		h.logger.Warn("Initial health check failed", "error", err)
	} else {
		h.logger.Info("All components are healthy")
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		h.logger.Info("Shutdown signal received")
	case <-ctx.Done():
		h.logger.Info("Context cancelled, shutting down")
	}

	// Graceful shutdown
	shutdownTimeout := time.Duration(h.config.HTTP.ShutdownTimeout) * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	return h.Stop(shutdownCtx)
}

// LoadConfigAndLogger is a utility function to load configuration and setup logging
func LoadConfigAndLogger() (*Config, *logger.Logger, error) {
	cfg, err := Load()
	if err != nil {
		return nil, nil, err
	}

	logger := logger.Setup(&cfg.Logging)
	return cfg, logger, nil
}
