package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	logger := logger.Setup(cfg)
	logger.Info("Starting Persistent Context Server",
		"version", "1.0.0",
		"http_port", cfg.HTTP.Port,
		"vectordb_url", cfg.VectorDB.URL,
		"llm_url", cfg.LLM.URL,
	)
	
	application, err := NewApplication(cfg, logger)
	if err != nil {
		logger.Error("Failed to create application", "error", err)
		os.Exit(1)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := application.Start(ctx); err != nil {
		logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
	
	logger.Info("All services started successfully")
	
	if err := application.HealthCheck(ctx); err != nil {
		logger.Warn("Initial health check failed", "error", err)
	} else {
		logger.Info("All services are healthy")
	}
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-quit:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}
	
	logger.Info("Shutting down all services...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 
		time.Duration(cfg.HTTP.ShutdownTimeout)*time.Second)
	defer shutdownCancel()
	
	if err := application.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
	}
	
	logger.Info("Server stopped")
}