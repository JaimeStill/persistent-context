package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/logger"
)

// Application defines the interface that all application types must implement
type Application interface {
	// Initialize sets up all components
	Initialize() error
	
	// Start begins running all components
	Start(ctx context.Context) error
	
	// Stop gracefully shuts down all components
	Stop(ctx context.Context) error
	
	// HealthCheck verifies all components are healthy
	HealthCheck(ctx context.Context) error
	
	// Name returns the application name for logging
	Name() string
}

// Runner provides the common process lifecycle management for any Application
type Runner struct {
	app    Application
	config *config.Config
	logger *logger.Logger
}

// NewRunner creates a new application runner
func NewRunner(app Application, cfg *config.Config, logger *logger.Logger) *Runner {
	return &Runner{
		app:    app,
		config: cfg,
		logger: logger,
	}
}

// Run executes the application with proper lifecycle management
func (r *Runner) Run() error {
	// Initialize the application
	r.logger.Info("Initializing application", "name", r.app.Name())
	if err := r.app.Initialize(); err != nil {
		r.logger.Error("Failed to initialize application", "error", err)
		return err
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the application
	r.logger.Info("Starting application", "name", r.app.Name())
	if err := r.app.Start(ctx); err != nil {
		r.logger.Error("Failed to start application", "error", err)
		return err
	}

	r.logger.Info("Application started successfully", "name", r.app.Name())

	// Perform initial health check
	if err := r.app.HealthCheck(ctx); err != nil {
		r.logger.Warn("Initial health check failed", "error", err)
	} else {
		r.logger.Info("All components are healthy")
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		r.logger.Info("Shutdown signal received")
	case <-ctx.Done():
		r.logger.Info("Context cancelled, shutting down")
	}

	// Graceful shutdown
	r.logger.Info("Shutting down application...")
	
	shutdownTimeout := time.Duration(r.config.HTTP.ShutdownTimeout) * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := r.app.Stop(shutdownCtx); err != nil {
		r.logger.Error("Error during shutdown", "error", err)
		return err
	}

	r.logger.Info("Application stopped", "name", r.app.Name())
	return nil
}

// LoadConfigAndLogger is a utility function to load configuration and setup logging
func LoadConfigAndLogger() (*config.Config, *logger.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	logger := logger.Setup(cfg)
	return cfg, logger, nil
}

// HealthChecker is a simple adapter for components that need health checking
type HealthChecker struct {
	name    string
	checker interface{ HealthCheck(context.Context) error }
}

// NewHealthChecker creates a new health checker adapter
func NewHealthChecker(name string, checker interface{ HealthCheck(context.Context) error }) *HealthChecker {
	return &HealthChecker{
		name:    name,
		checker: checker,
	}
}

func (h *HealthChecker) Name() string {
	return h.name
}

func (h *HealthChecker) HealthCheck(ctx context.Context) error {
	return h.checker.HealthCheck(ctx)
}

func (h *HealthChecker) IsInitialized() bool {
	return true
}

func (h *HealthChecker) IsRunning() bool {
	return true
}