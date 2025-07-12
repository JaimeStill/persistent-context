package app

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/http"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/mcp"
)

// MCPApplication represents the MCP server application
type MCPApplication struct {
	config    *config.Config
	logger    *logger.Logger
	
	// MCP-specific components
	httpClient *http.Client
	mcpServer  *mcp.Server
}

// NewMCPApplication creates a new MCP application instance
func NewMCPApplication(cfg *config.Config, logger *logger.Logger) *MCPApplication {
	return &MCPApplication{
		config: cfg,
		logger: logger,
	}
}

// Name returns the application name
func (a *MCPApplication) Name() string {
	return "mcp-server"
}

// Initialize sets up all MCP components
func (a *MCPApplication) Initialize() error {
	// Initialize HTTP client to communicate with web server
	a.httpClient = http.NewClient(a.config.MCP.WebAPIURL, a.config.MCP.Timeout)

	// Initialize MCP Server with HTTP client
	a.mcpServer = mcp.NewServer(a.config.MCP, a.httpClient, a.logger)

	a.logger.Info("MCP application initialized successfully")
	return nil
}

// Start begins running the MCP server
func (a *MCPApplication) Start(ctx context.Context) error {
	// Start MCP server (blocking call for stdio communication)
	a.logger.Info("Starting MCP server for stdio communication")
	return a.mcpServer.ServeStdio(ctx)
}

// Stop gracefully shuts down the MCP server
func (a *MCPApplication) Stop(ctx context.Context) error {
	a.logger.Info("Stopping MCP application")
	a.logger.Info("MCP application stopped")
	return nil
}

// HealthCheck verifies all MCP components are healthy
func (a *MCPApplication) HealthCheck(ctx context.Context) error {
	// Check web server health via HTTP client
	if err := a.httpClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("web server unhealthy: %w", err)
	}

	return nil
}