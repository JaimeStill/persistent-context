package main

import (
	"log"

	"github.com/JaimeStill/persistent-context/persistent-context-svc/app"
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

	// Create and run the host
	host := app.NewHost(cfg, logger)
	if err := host.Run(); err != nil {
		logger.Error("Host failed", "error", err)
		log.Fatalf("Host failed: %v", err)
	}
}