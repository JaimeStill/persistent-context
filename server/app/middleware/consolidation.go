package middleware

import (
	"context"
	"fmt"
)

// ConsolidationMiddleware triggers consolidation events
func ConsolidationMiddleware(triggerFunc func(context.Context, *MemoryContext) error) MiddlewareFunc {
	return func(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error {
		memCtx.Stage = "consolidation"
		
		// Process through next middleware first
		if err := next(ctx, memCtx); err != nil {
			return err
		}
		
		// Then trigger consolidation if needed
		if triggerFunc != nil {
			if err := triggerFunc(ctx, memCtx); err != nil {
				return fmt.Errorf("consolidation trigger failed: %w", err)
			}
		}
		
		return nil
	}
}