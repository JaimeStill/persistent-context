package app

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/app/lifecycle"
	"github.com/JaimeStill/persistent-context/app/services"
	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/pkg/logger"
)

// Orchestrator manages the application lifecycle and service coordination
type Orchestrator struct {
	registry *lifecycle.Registry
	config   *config.Config
	logger   *logger.Logger
	services map[string]services.Service
}

// NewOrchestrator creates a new application orchestrator
func NewOrchestrator(cfg *config.Config, logger *logger.Logger) *Orchestrator {
	return &Orchestrator{
		registry: lifecycle.NewRegistry(),
		config:   cfg,
		logger:   logger,
		services: make(map[string]services.Service),
	}
}

// RegisterServices registers all application services with the orchestrator
func (o *Orchestrator) RegisterServices(ctx context.Context) error {
	o.logger.Info("Registering application services")

	// Register VectorDB service
	vectordbService := services.NewVectorDBService(&o.config.VectorDB)
	if err := o.registry.Register(vectordbService); err != nil {
		return fmt.Errorf("failed to register vectordb service: %w", err)
	}
	o.services["vectordb"] = vectordbService

	// Register LLM service
	llmService := services.NewLLMService(&o.config.LLM)
	if err := o.registry.Register(llmService); err != nil {
		return fmt.Errorf("failed to register llm service: %w", err)
	}
	o.services["llm"] = llmService

	// Register Memory service (depends on vectordb and llm)
	memoryService := services.NewMemoryService(&o.config.Memory)
	if err := o.registry.Register(memoryService); err != nil {
		return fmt.Errorf("failed to register memory service: %w", err)
	}
	o.services["memory"] = memoryService

	// Register HTTP service
	httpService := services.NewHTTPService(&o.config.HTTP)
	if err := o.registry.Register(httpService); err != nil {
		return fmt.Errorf("failed to register http service: %w", err)
	}
	o.services["http"] = httpService

	// Register MCP service (depends on memory)
	mcpService := services.NewMCPService(&o.config.MCP)
	if err := o.registry.Register(mcpService); err != nil {
		return fmt.Errorf("failed to register mcp service: %w", err)
	}
	o.services["mcp"] = mcpService

	o.logger.Info("All services registered successfully", 
		"service_count", len(o.services))

	return nil
}

// Initialize initializes all registered services in dependency order
func (o *Orchestrator) Initialize(ctx context.Context) error {
	o.logger.Info("Initializing application services")
	
	// First, initialize base services that don't depend on others
	if err := o.registry.InitializeAll(ctx); err != nil {
		return fmt.Errorf("failed to initialize base services: %w", err)
	}
	
	// Then handle dependency injection for services that need it
	if err := o.injectDependencies(); err != nil {
		return fmt.Errorf("failed to inject dependencies: %w", err)
	}
	
	o.logger.Info("All services initialized successfully")
	return nil
}

// injectDependencies handles manual dependency injection for services that need it
func (o *Orchestrator) injectDependencies() error {
	// Get service instances
	vectordbService := o.services["vectordb"].(*services.VectorDBService)
	llmService := o.services["llm"].(*services.LLMService)
	memoryService := o.services["memory"].(*services.MemoryService)
	httpService := o.services["http"].(*services.HTTPService)
	mcpService := o.services["mcp"].(*services.MCPService)
	
	// Initialize memory service with its dependencies
	if err := memoryService.InitializeWithDependencies(vectordbService.DB(), llmService.LLM()); err != nil {
		return fmt.Errorf("failed to initialize memory service with dependencies: %w", err)
	}
	
	// Initialize HTTP service with its dependencies
	if err := httpService.InitializeWithDependencies(vectordbService, llmService); err != nil {
		return fmt.Errorf("failed to initialize HTTP service with dependencies: %w", err)
	}
	
	// Initialize MCP service with its dependencies
	if err := mcpService.InitializeWithDependencies(memoryService.Store()); err != nil {
		return fmt.Errorf("failed to initialize MCP service with dependencies: %w", err)
	}
	
	return nil
}

// Start starts all registered services
func (o *Orchestrator) Start(ctx context.Context) error {
	o.logger.Info("Starting application services")
	
	if err := o.registry.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	
	o.logger.Info("All services started successfully")
	return nil
}

// Stop gracefully stops all services
func (o *Orchestrator) Stop(ctx context.Context) error {
	o.logger.Info("Stopping application services")
	
	if err := o.registry.StopAll(ctx); err != nil {
		o.logger.Error("Failed to stop some services", "error", err)
		return fmt.Errorf("failed to stop services: %w", err)
	}
	
	o.logger.Info("All services stopped successfully")
	return nil
}

// HealthCheck performs health checks on all services
func (o *Orchestrator) HealthCheck(ctx context.Context) error {
	results := o.registry.HealthCheckAll(ctx)
	
	var unhealthy []string
	for serviceName, err := range results {
		if err != nil {
			unhealthy = append(unhealthy, serviceName)
			o.logger.Error("Service health check failed", 
				"service", serviceName, 
				"error", err)
		}
	}
	
	if len(unhealthy) > 0 {
		return fmt.Errorf("health check failed for services: %v", unhealthy)
	}
	
	o.logger.Info("All services are healthy")
	return nil
}

// GetService retrieves a service by name
func (o *Orchestrator) GetService(name string) services.Service {
	return o.services[name]
}

// GetMemoryService is a convenience method to get the memory service
func (o *Orchestrator) GetMemoryService() *services.MemoryService {
	if service, ok := o.services["memory"].(*services.MemoryService); ok {
		return service
	}
	return nil
}

// GetHTTPService is a convenience method to get the HTTP service
func (o *Orchestrator) GetHTTPService() *services.HTTPService {
	if service, ok := o.services["http"].(*services.HTTPService); ok {
		return service
	}
	return nil
}

// GetMCPService is a convenience method to get the MCP service
func (o *Orchestrator) GetMCPService() *services.MCPService {
	if service, ok := o.services["mcp"].(*services.MCPService); ok {
		return service
	}
	return nil
}