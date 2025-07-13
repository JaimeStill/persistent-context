package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/mcp"
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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	log := logger.Setup(cfg)

	// Log startup info
	log.Info("Starting Persistent Context MCP Server",
		"version", cfg.MCP.Version,
		"name", cfg.MCP.Name,
		"web_api_url", cfg.MCP.WebAPIURL,
	)

	// Create HTTP client to communicate with web server
	httpClient := http.NewClient(cfg.MCP.WebAPIURL, cfg.MCP.Timeout)

	// Create MCP server
	mcpServer := mcp.NewServer(cfg.MCP, httpClient, log)

	// Start stdio communication (blocking)
	log.Info("Starting MCP server for stdio communication")
	ctx := context.Background()
	if err := mcpServer.ServeStdio(ctx); err != nil {
		log.Error("MCP server failed", "error", err)
	}

	log.Info("MCP server stopped")
}