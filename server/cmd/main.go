package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	httpserver "github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/mcp"
	"github.com/JaimeStill/persistent-context/internal/storage"
	"github.com/JaimeStill/persistent-context/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Setup logging
	logger := logger.Setup(cfg)
	logger.Info("Starting Persistent Context Server",
		"version", "1.0.0",
		"port", cfg.Server.Port,
		"qdrant_url", cfg.Qdrant.URL,
		"ollama_url", cfg.Ollama.URL,
	)
	
	// Create memory store
	memoryStore := storage.NewMemoryStore()
	
	// Create MCP server
	mcpServer := mcp.NewServer(cfg.MCP.Name, cfg.MCP.Version, memoryStore)
	
	// Create HTTP server dependencies
	deps := &httpserver.Dependencies{
		QdrantHealth: memoryStore, // Memory store implements HealthChecker
		OllamaHealth: nil,         // TODO: Implement Ollama health checker in Session 2
	}
	
	// Create HTTP server
	httpServer := httpserver.NewServer(cfg, deps)
	
	// Start servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := httpServer.Start(); err != nil {
			logger.Error("HTTP server failed", "error", err)
			cancel()
		}
	}()
	
	// Start MCP server if enabled
	if cfg.MCP.Enabled {
		go func() {
			logger.Info("Starting MCP server", "name", cfg.MCP.Name, "version", cfg.MCP.Version)
			if err := mcpServer.ServeStdio(ctx); err != nil {
				logger.Error("MCP server failed", "error", err)
			}
		}()
	} else {
		logger.Info("MCP server disabled (set APP_MCP_ENABLED=true to enable)")
	}
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-quit:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}
	
	// Graceful shutdown
	logger.Info("Shutting down servers...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 
		time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer shutdownCancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}
	
	logger.Info("Server stopped")
}