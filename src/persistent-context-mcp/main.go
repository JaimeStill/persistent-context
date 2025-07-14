package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/logger"
	"github.com/JaimeStill/persistent-context/persistent-context-mcp/app"
)

func main() {
	// Parse command-line flags
	var (
		stdio = flag.Bool("stdio", false, "Start MCP server for stdio communication")
		help  = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Printf("Persistent Context MCP Server\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s [flags]\n\n", os.Args[0])
		fmt.Printf("Flags:\n")
		flag.PrintDefaults()
		return
	}

	// Require --stdio flag to start server
	if !*stdio {
		fmt.Printf("Error: --stdio flag is required to start the MCP server\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s --stdio\n", os.Args[0])
		fmt.Printf("  %s --help\n", os.Args[0])
		os.Exit(1)
	}

	// Load MCP configuration
	mcpConfig, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	loggingConfig := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.Setup(loggingConfig)

	// Log startup info
	logger.Info("Starting Persistent Context MCP Server",
		"version", mcpConfig.Version,
		"name", mcpConfig.Name,
		"web_api_url", mcpConfig.WebAPIURL,
	)

	// Create HTTP client to communicate with web server
	httpClient := app.NewClient(mcpConfig.WebAPIURL, mcpConfig.Timeout)

	// Create MCP server
	mcpServer := app.NewServer(mcpConfig, httpClient, logger)

	// Start stdio communication (blocking)
	logger.Info("Starting MCP server for stdio communication")
	ctx := context.Background()
	if err := mcpServer.ServeStdio(ctx); err != nil {
		logger.Error("MCP server failed", "error", err)
	}

	logger.Info("MCP server stopped")
}