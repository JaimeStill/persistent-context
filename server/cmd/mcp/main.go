package main

import (
	"log"

	"github.com/JaimeStill/persistent-context/internal/app"
)

func main() {
	// Load configuration and logger
	cfg, logger, err := app.LoadConfigAndLogger()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log startup info
	logger.Info("Starting Persistent Context MCP Server",
		"version", "1.0.0",
		"name", cfg.MCP.Name,
		"capture_mode", cfg.MCP.CaptureMode,
	)

	// Create MCP application
	mcpApp := app.NewMCPApplication(cfg, logger)

	// Create and run the application runner
	runner := app.NewRunner(mcpApp, cfg, logger)
	if err := runner.Run(); err != nil {
		logger.Error("Application failed", "error", err)
		log.Fatalf("Application failed: %v", err)
	}
}