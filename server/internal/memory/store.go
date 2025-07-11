package memory

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// MemoryStore implements memory storage using vectordb and llm interfaces
type MemoryStore struct {
	vectorDB  vectordb.VectorDB
	llmClient llm.LLM
	config    *config.MemoryConfig
	counter   int64
}

// Dependencies holds the dependencies for memory store
type Dependencies struct {
	VectorDB  vectordb.VectorDB
	LLMClient llm.LLM
	Config    *config.MemoryConfig
}

// NewMemoryStore creates a new memory store with Qdrant and Ollama integration
func NewMemoryStore(deps *Dependencies) *MemoryStore {
	return &MemoryStore{
		vectorDB:  deps.VectorDB,
		llmClient: deps.LLMClient,
		config:    deps.Config,
		counter:   time.Now().UnixNano(), // Use timestamp as base counter
	}
}

// CaptureContext implements the MCP interface for capturing context
func (ms *MemoryStore) CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) error {
	ms.counter++
	
	// Generate embedding for the content
	embedding, err := ms.llmClient.GenerateEmbedding(ctx, content)
	if err != nil {
		slog.Error("Failed to generate embedding", "error", err, "source", source)
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create memory entry
	entry := &types.MemoryEntry{
		ID:         fmt.Sprintf("mem_%d_%d", time.Now().Unix(), ms.counter),
		Type:       types.TypeEpisodic,
		Content:    content,
		Embedding:  embedding,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Strength:   1.0, // New memories start with full strength
	}
	
	// Add source to metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}
	entry.Metadata["source"] = source
	entry.Metadata["captured_at"] = time.Now().Unix()
	
	// Store in vector database
	if err := ms.vectorDB.Store(ctx, entry); err != nil {
		slog.Error("Failed to store memory in vector database", "error", err, "id", entry.ID)
		return fmt.Errorf("failed to store memory: %w", err)
	}
	
	slog.Info("Memory captured and stored",
		"source", source,
		"id", entry.ID,
		"content_length", len(content),
		"embedding_dim", len(embedding))
	
	return nil
}

// GetMemories retrieves recent memories from episodic storage
func (ms *MemoryStore) GetMemories(ctx context.Context, limit uint64) ([]*types.MemoryEntry, error) {
	if limit == 0 {
		limit = ms.config.BatchSize
	}

	// Create a dummy vector for recent memories query (we'll improve this in Session 3)
	dummyVector := make([]float32, 1536) // Standard embedding dimension
	
	memories, err := ms.vectorDB.Query(ctx, types.TypeEpisodic, dummyVector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}

	return memories, nil
}

// GetMemoryByID retrieves a specific memory by ID
func (ms *MemoryStore) GetMemoryByID(ctx context.Context, id string) (*types.MemoryEntry, error) {
	entry, err := ms.vectorDB.Retrieve(ctx, types.TypeEpisodic, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memory %s: %w", id, err)
	}

	// Update access time (we'll implement this properly in Session 3)
	entry.AccessedAt = time.Now()
	
	return entry, nil
}

// QuerySimilarMemories finds memories similar to the given content
func (ms *MemoryStore) QuerySimilarMemories(ctx context.Context, content string, memType types.MemoryType, limit uint64) ([]*types.MemoryEntry, error) {
	// Generate embedding for the query content
	embedding, err := ms.llmClient.GenerateEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if limit == 0 {
		limit = 10 // Default limit
	}

	// Query vector database
	memories, err := ms.vectorDB.Query(ctx, memType, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar memories: %w", err)
	}

	return memories, nil
}

// BatchStoreMemories stores multiple memories efficiently
func (ms *MemoryStore) BatchStoreMemories(ctx context.Context, entries []*types.MemoryEntry) error {
	batchSize := int(ms.config.BatchSize)
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(entries); i += batchSize {
		end := i + batchSize
		if end > len(entries) {
			end = len(entries)
		}

		batch := entries[i:end]
		for _, entry := range batch {
			if err := ms.vectorDB.Store(ctx, entry); err != nil {
				slog.Error("Failed to store memory in batch", "error", err, "id", entry.ID)
				return fmt.Errorf("failed to store memory %s: %w", entry.ID, err)
			}
		}

		slog.Debug("Stored memory batch", "count", len(batch), "total_processed", end)
	}

	return nil
}

// ConsolidateMemories processes episodic memories into semantic knowledge
func (ms *MemoryStore) ConsolidateMemories(ctx context.Context, memories []*types.MemoryEntry) error {
	if len(memories) == 0 {
		return nil
	}

	// Extract content for consolidation
	memoryTexts := make([]string, len(memories))
	for i, mem := range memories {
		memoryTexts[i] = mem.Content
	}

	// Use LLM to consolidate memories
	consolidatedContent, err := ms.llmClient.ConsolidateMemories(ctx, memoryTexts)
	if err != nil {
		return fmt.Errorf("failed to consolidate memories: %w", err)
	}

	// Generate embedding for consolidated content
	embedding, err := ms.llmClient.GenerateEmbedding(ctx, consolidatedContent)
	if err != nil {
		return fmt.Errorf("failed to generate embedding for consolidated memory: %w", err)
	}

	// Create semantic memory entry
	ms.counter++
	semanticEntry := &types.MemoryEntry{
		ID:        fmt.Sprintf("semantic_%d_%d", time.Now().Unix(), ms.counter),
		Type:      types.TypeSemantic,
		Content:   consolidatedContent,
		Embedding: embedding,
		Metadata: map[string]any{
			"source_memories": len(memories),
			"consolidation_timestamp": time.Now().Unix(),
			"consolidated_from": extractMemoryIDs(memories),
		},
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Strength:   1.0,
	}

	// Store semantic memory
	if err := ms.vectorDB.Store(ctx, semanticEntry); err != nil {
		return fmt.Errorf("failed to store semantic memory: %w", err)
	}

	slog.Info("Consolidated memories into semantic knowledge",
		"semantic_id", semanticEntry.ID,
		"source_count", len(memories),
		"content_length", len(consolidatedContent))

	return nil
}

// GetMemoryStats returns statistics about stored memories
func (ms *MemoryStore) GetMemoryStats(ctx context.Context) (map[string]any, error) {
	stats := map[string]any{
		"episodic_memories":      0,
		"semantic_memories":      0,
		"procedural_memories":    0,
		"metacognitive_memories": 0,
		"total_memories":         0,
	}

	// For now, return basic stats (we'll enhance this in Session 3)
	// This would require additional Qdrant queries to get accurate counts
	
	return stats, nil
}

// HealthCheck implements the HealthChecker interface
func (ms *MemoryStore) HealthCheck(ctx context.Context) error {
	// Check Qdrant connectivity
	if err := ms.vectorDB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vector database health check failed: %w", err)
	}

	// Check Ollama connectivity
	if err := ms.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("LLM client health check failed: %w", err)
	}

	return nil
}

// extractMemoryIDs extracts IDs from memory entries for metadata
func extractMemoryIDs(memories []*types.MemoryEntry) []string {
	ids := make([]string, len(memories))
	for i, mem := range memories {
		ids[i] = mem.ID
	}
	return ids
}