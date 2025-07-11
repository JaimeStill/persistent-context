package middleware

import (
	"context"
	"log/slog"

	"github.com/JaimeStill/persistent-context/internal/types"
)

// MemoryContext holds the context for memory processing
type MemoryContext struct {
	Memory   *types.MemoryEntry
	Metadata map[string]any
	Source   string
	Stage    string
}

// MiddlewareFunc represents a middleware function for memory processing
type MiddlewareFunc func(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error

// Pipeline represents a memory processing pipeline with middleware
type Pipeline struct {
	middleware []MiddlewareFunc
	logger     *slog.Logger
}

// NewPipeline creates a new memory processing pipeline
func NewPipeline(logger *slog.Logger) *Pipeline {
	return &Pipeline{
		middleware: make([]MiddlewareFunc, 0),
		logger:     logger,
	}
}

// Use adds middleware to the pipeline
func (p *Pipeline) Use(middleware MiddlewareFunc) {
	p.middleware = append(p.middleware, middleware)
}

// Process processes a memory through the pipeline
func (p *Pipeline) Process(ctx context.Context, memory *types.MemoryEntry, source string) error {
	memCtx := &MemoryContext{
		Memory:   memory,
		Metadata: make(map[string]any),
		Source:   source,
		Stage:    "start",
	}

	// Chain middleware functions
	var next func(context.Context, *MemoryContext) error
	next = func(ctx context.Context, memCtx *MemoryContext) error {
		// Final stage - nothing to do
		memCtx.Stage = "complete"
		return nil
	}

	// Build the middleware chain in reverse order
	for i := len(p.middleware) - 1; i >= 0; i-- {
		middleware := p.middleware[i]
		currentNext := next
		next = func(ctx context.Context, memCtx *MemoryContext) error {
			return middleware(ctx, memCtx, currentNext)
		}
	}

	// Execute the chain
	return next(ctx, memCtx)
}