package services

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	httpserver "github.com/JaimeStill/persistent-context/internal/http"
)

// HTTPService wraps the HTTP server as a managed service
type HTTPService struct {
	BaseService
	server *httpserver.Server
	config *config.HTTPConfig
}

// NewHTTPService creates a new HTTP service
func NewHTTPService(cfg *config.HTTPConfig) *HTTPService {
	return &HTTPService{
		BaseService: NewBaseService("http", "vectordb", "llm"), // Depends on vectordb and llm for health checks
		config:      cfg,
	}
}

// Initialize creates the HTTP server with its dependencies
func (s *HTTPService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// This will be handled by dependency injection
	s.SetInitialized(true)
	return nil
}

// InitializeWithDependencies initializes the HTTP service with health check dependencies
func (s *HTTPService) InitializeWithDependencies(vdbService *VectorDBService, llmService *LLMService) error {
	if s.IsInitialized() {
		return nil
	}

	// Create HTTP server dependencies
	deps := &httpserver.Dependencies{
		VectorDBHealth: vdbService,
		LLMHealth:      llmService,
	}

	// Create HTTP server
	s.server = httpserver.NewServer(s.config, deps)
	s.SetInitialized(true)
	return nil
}

// Start begins the HTTP server
func (s *HTTPService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	go func() {
		if err := s.server.Start(); err != nil {
			// Log error - in a real implementation we'd want better error handling
		}
	}()

	s.SetRunning(true)
	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *HTTPService) Stop(ctx context.Context) error {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
	}
	s.SetRunning(false)
	return nil
}

// HealthCheck verifies the HTTP server is operational
func (s *HTTPService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	if !s.IsRunning() {
		return fmt.Errorf("HTTP server not running")
	}

	return nil
}

// Server returns the HTTP server
func (s *HTTPService) Server() *httpserver.Server {
	return s.server
}