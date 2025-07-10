package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// LifecycleManager handles application lifecycle events
type LifecycleManager struct {
	registry      *Registry
	shutdownFuncs []func(context.Context) error
	mu            sync.Mutex
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(registry *Registry) *LifecycleManager {
	return &LifecycleManager{
		registry:      registry,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}
}

// OnShutdown registers a function to be called during shutdown
func (m *LifecycleManager) OnShutdown(fn func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownFuncs = append(m.shutdownFuncs, fn)
}

// WaitForShutdown blocks until a shutdown signal is received
func (m *LifecycleManager) WaitForShutdown(ctx context.Context, timeout time.Duration) error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or context cancellation
	select {
	case <-sigChan:
		// Signal received, proceed with shutdown
	case <-ctx.Done():
		// Context cancelled
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute shutdown functions
	m.mu.Lock()
	funcs := make([]func(context.Context) error, len(m.shutdownFuncs))
	copy(funcs, m.shutdownFuncs)
	m.mu.Unlock()

	// Run shutdown functions in reverse order
	for i := len(funcs) - 1; i >= 0; i-- {
		if err := funcs[i](shutdownCtx); err != nil {
			// Log error but continue shutdown
			continue
		}
	}

	// Stop all services
	return m.registry.StopAll(shutdownCtx)
}