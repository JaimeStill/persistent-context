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
	logger.Info("Starting Persistent Context Web Server",
		"version", "1.0.0",
		"http_port", cfg.HTTP.Port,
		"vectordb_url", cfg.VectorDB.URL,
		"llm_url", cfg.LLM.URL,
	)

	// Create web application
	webApp := app.NewWebApplication(cfg, logger)

	// Create and run the application runner
	runner := app.NewRunner(webApp, cfg, logger)
	if err := runner.Run(); err != nil {
		logger.Error("Application failed", "error", err)
		log.Fatalf("Application failed: %v", err)
	}
}