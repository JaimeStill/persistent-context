package middleware

import (
	"context"
	"fmt"
)

// ValidationMiddleware validates memory entries
func ValidationMiddleware(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error {
	memCtx.Stage = "validation"
	
	if memCtx.Memory == nil {
		return fmt.Errorf("memory entry is nil")
	}
	
	if memCtx.Memory.Content == "" {
		return fmt.Errorf("memory content is empty")
	}
	
	if memCtx.Memory.ID == "" {
		return fmt.Errorf("memory ID is empty")
	}
	
	// Add validation metadata
	memCtx.Metadata["validated"] = true
	memCtx.Metadata["content_length"] = len(memCtx.Memory.Content)
	
	return next(ctx, memCtx)
}