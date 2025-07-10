package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/JaimeStill/persistent-context/app/services"
)

// Registry manages the lifecycle of all registered services
type Registry struct {
	mu          sync.RWMutex
	services    map[string]services.Service
	order       []string // Startup order based on dependencies
	initialized bool
}

// NewRegistry creates a new service registry
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]services.Service),
		order:    make([]string, 0),
	}
}

// Register adds a service to the registry
func (r *Registry) Register(service services.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return fmt.Errorf("cannot register service after initialization")
	}

	name := service.Name()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	slog.Info("Service registered", "name", name)
	return nil
}

// Get retrieves a service by name
func (r *Registry) Get(name string) services.Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services[name]
}

// InitializeAll initializes all services in dependency order
func (r *Registry) InitializeAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return nil
	}

	// Determine initialization order based on dependencies
	if err := r.computeInitializationOrder(); err != nil {
		return fmt.Errorf("failed to compute initialization order: %w", err)
	}

	// Initialize services in order
	for _, name := range r.order {
		service := r.services[name]
		slog.Info("Initializing service", "name", name)
		
		if err := service.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize service %s: %w", name, err)
		}
		
		slog.Info("Service initialized successfully", "name", name)
	}

	r.initialized = true
	return nil
}

// StartAll starts all services
func (r *Registry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return fmt.Errorf("services must be initialized before starting")
	}

	// Start services in initialization order
	for _, name := range r.order {
		service := r.services[name]
		slog.Info("Starting service", "name", name)
		
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
		
		slog.Info("Service started successfully", "name", name)
	}

	return nil
}

// StopAll stops all services in reverse order
func (r *Registry) StopAll(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Stop services in reverse order
	for i := len(r.order) - 1; i >= 0; i-- {
		name := r.order[i]
		service := r.services[name]
		slog.Info("Stopping service", "name", name)
		
		if err := service.Stop(ctx); err != nil {
			slog.Error("Failed to stop service", "name", name, "error", err)
			// Continue stopping other services
		} else {
			slog.Info("Service stopped successfully", "name", name)
		}
	}

	return nil
}

// HealthCheckAll performs health checks on all services
func (r *Registry) HealthCheckAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, service := range r.services {
		results[name] = service.HealthCheck(ctx)
	}
	return results
}

// computeInitializationOrder determines the order to initialize services based on dependencies
func (r *Registry) computeInitializationOrder() error {
	// For now, use a simple topological sort
	// This can be enhanced with proper dependency resolution
	visited := make(map[string]bool)
	tempMark := make(map[string]bool)
	order := make([]string, 0, len(r.services))

	var visit func(string) error
	visit = func(name string) error {
		if tempMark[name] {
			return fmt.Errorf("circular dependency detected at service %s", name)
		}
		if visited[name] {
			return nil
		}

		tempMark[name] = true
		service := r.services[name]
		
		// Check if service implements Dependencies interface
		if depService, ok := service.(services.Dependencies); ok {
			for _, dep := range depService.Require() {
				if _, exists := r.services[dep]; !exists {
					return fmt.Errorf("service %s depends on non-existent service %s", name, dep)
				}
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		tempMark[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	// Visit all services
	for name := range r.services {
		if err := visit(name); err != nil {
			return err
		}
	}

	r.order = order
	return nil
}