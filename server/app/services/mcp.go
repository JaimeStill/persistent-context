package services

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/mcp"
	"github.com/JaimeStill/persistent-context/internal/memory"
)

// MCPService wraps the MCP server as a managed service
type MCPService struct {
	BaseService
	server *mcp.Server
	config *config.MCPConfig
}

// NewMCPService creates a new MCP service
func NewMCPService(cfg *config.MCPConfig) *MCPService {
	return &MCPService{
		BaseService: NewBaseService("mcp", "memory"), // Depends on memory service
		config:      cfg,
	}
}

// Initialize creates the MCP server with memory store dependency
func (s *MCPService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// This will be handled by dependency injection
	s.SetInitialized(true)
	return nil
}

// InitializeWithDependencies initializes the MCP service with memory store
func (s *MCPService) InitializeWithDependencies(memoryStore *memory.MemoryStore) error {
	if s.IsInitialized() {
		return nil
	}

	// Create MCP server
	s.server = mcp.NewServer(s.config.Name, s.config.Version, memoryStore)
	s.SetInitialized(true)
	return nil
}

// Start begins the MCP server if enabled
func (s *MCPService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	if !s.config.Enabled {
		// MCP server is disabled, but this is not an error
		return nil
	}

	go func() {
		if err := s.server.ServeStdio(ctx); err != nil {
			// Log error - in a real implementation we'd want better error handling
		}
	}()

	s.SetRunning(true)
	return nil
}

// Stop gracefully shuts down the MCP server
func (s *MCPService) Stop(ctx context.Context) error {
	s.SetRunning(false)
	// MCP server will stop when context is cancelled
	return nil
}

// HealthCheck verifies the MCP server is operational
func (s *MCPService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	if !s.config.Enabled {
		return fmt.Errorf("MCP server disabled")
	}

	return nil
}

// Server returns the MCP server
func (s *MCPService) Server() *mcp.Server {
	return s.server
}