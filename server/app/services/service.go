package services

import (
	"context"
)

// Service defines the interface that all managed services must implement
type Service interface {
	// Name returns the unique identifier for this service
	Name() string
	
	// Initialize prepares the service for use (e.g., establish connections)
	Initialize(ctx context.Context) error
	
	// Start begins the service's main operations
	Start(ctx context.Context) error
	
	// Stop gracefully shuts down the service
	Stop(ctx context.Context) error
	
	// HealthCheck verifies the service is operating correctly
	HealthCheck(ctx context.Context) error
}

// Dependencies allows services to declare their dependencies
type Dependencies interface {
	// Require returns the names of services this service depends on
	Require() []string
}

// BaseService provides common functionality for all services
type BaseService struct {
	name         string
	dependencies []string
	initialized  bool
	running      bool
}

// NewBaseService creates a new base service
func NewBaseService(name string, deps ...string) BaseService {
	return BaseService{
		name:         name,
		dependencies: deps,
		initialized:  false,
		running:      false,
	}
}

// Name returns the service name
func (s *BaseService) Name() string {
	return s.name
}

// Require returns the service dependencies
func (s *BaseService) Require() []string {
	return s.dependencies
}

// IsInitialized returns whether the service has been initialized
func (s *BaseService) IsInitialized() bool {
	return s.initialized
}

// SetInitialized marks the service as initialized
func (s *BaseService) SetInitialized(initialized bool) {
	s.initialized = initialized
}

// IsRunning returns whether the service is running
func (s *BaseService) IsRunning() bool {
	return s.running
}

// SetRunning marks the service as running
func (s *BaseService) SetRunning(running bool) {
	s.running = running
}