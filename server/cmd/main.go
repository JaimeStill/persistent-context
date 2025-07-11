package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JaimeStill/persistent-context/app"
	"github.com/JaimeStill/persistent-context/internal/config"
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
		"http_port", cfg.HTTP.Port,
		"vectordb_url", cfg.VectorDB.URL,
		"llm_url", cfg.LLM.URL,
	)
	
	// Create application orchestrator
	orchestrator := app.NewOrchestrator(cfg, logger)
	
	// Application lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Register all services
	if err := orchestrator.RegisterServices(ctx); err != nil {
		logger.Error("Failed to register services", "error", err)
		os.Exit(1)
	}
	
	// Initialize all services
	if err := orchestrator.Initialize(ctx); err != nil {
		logger.Error("Failed to initialize services", "error", err)
		os.Exit(1)
	}
	
	// Start all services
	if err := orchestrator.Start(ctx); err != nil {
		logger.Error("Failed to start services", "error", err)
		os.Exit(1)
	}
	
	logger.Info("All services started successfully")
	
	// Perform initial health check
	if err := orchestrator.HealthCheck(ctx); err != nil {
		logger.Warn("Initial health check failed", "error", err)
	} else {
		logger.Info("All services are healthy")
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
	logger.Info("Shutting down all services...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 
		time.Duration(cfg.HTTP.ShutdownTimeout)*time.Second)
	defer shutdownCancel()
	
	if err := orchestrator.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
	}
	
	logger.Info("Server stopped")
}