package middleware

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/types"
)

// RoutingMiddleware routes memories to different handlers based on type
func RoutingMiddleware(handlers map[types.MemoryType]func(context.Context, *MemoryContext) error) MiddlewareFunc {
	return func(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error {
		memCtx.Stage = "routing"
		
		// Check if there's a specific handler for this memory type
		if handler, exists := handlers[memCtx.Memory.Type]; exists {
			memCtx.Metadata["routed_to"] = string(memCtx.Memory.Type)
			if err := handler(ctx, memCtx); err != nil {
				return fmt.Errorf("routing handler failed: %w", err)
			}
		}
		
		return next(ctx, memCtx)
	}
}