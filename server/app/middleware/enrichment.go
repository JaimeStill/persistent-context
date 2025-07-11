package middleware

import (
	"context"
)

// EnrichmentMiddleware enriches memory entries with additional metadata
func EnrichmentMiddleware(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error {
	memCtx.Stage = "enrichment"
	
	// Add enrichment metadata
	if memCtx.Memory.Metadata == nil {
		memCtx.Memory.Metadata = make(map[string]any)
	}
	
	// Add processing timestamp
	memCtx.Memory.Metadata["processed_at"] = memCtx.Memory.CreatedAt.Unix()
	memCtx.Memory.Metadata["pipeline_source"] = memCtx.Source
	
	// Add content analysis
	memCtx.Memory.Metadata["word_count"] = len(memCtx.Memory.Content) / 5 // Rough estimate
	
	// Copy middleware metadata to memory
	for k, v := range memCtx.Metadata {
		memCtx.Memory.Metadata["pipeline_"+k] = v
	}
	
	return next(ctx, memCtx)
}