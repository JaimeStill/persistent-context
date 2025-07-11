package middleware

import (
	"context"
	"log/slog"
)

// LoggingMiddleware logs memory processing
func LoggingMiddleware(logger *slog.Logger) MiddlewareFunc {
	return func(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error {
		logger.Info("Processing memory through pipeline",
			"memory_id", memCtx.Memory.ID,
			"source", memCtx.Source,
			"stage", memCtx.Stage,
			"content_length", len(memCtx.Memory.Content))
		
		err := next(ctx, memCtx)
		
		if err != nil {
			logger.Error("Memory processing failed",
				"memory_id", memCtx.Memory.ID,
				"source", memCtx.Source,
				"stage", memCtx.Stage,
				"error", err)
		} else {
			logger.Info("Memory processing completed",
				"memory_id", memCtx.Memory.ID,
				"source", memCtx.Source,
				"final_stage", memCtx.Stage)
		}
		
		return err
	}
}