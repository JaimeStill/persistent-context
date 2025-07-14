package journal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/llm"
	"github.com/JaimeStill/persistent-context/pkg/vectordb"
	"github.com/JaimeStill/persistent-context/pkg/models"
)

// VectorJournal implements LLM memory storage using vectordb and llm interfaces
type VectorJournal struct {
	vectorDB     vectordb.VectorDB
	llmClient    llm.LLM
	config       *config.JournalConfig
	scorer       *MemoryScorer
	associations *AssociationTracker
	analyzer     *AssociationAnalyzer
	counter      int64
}

// NewVectorJournal creates a new vector-based journal implementation
func NewVectorJournal(deps *Dependencies) *VectorJournal {
	associations := NewAssociationTracker()
	
	return &VectorJournal{
		vectorDB:     deps.VectorDB,
		llmClient:    deps.LLMClient,
		config:       deps.Config,
		scorer:       NewMemoryScorer(deps.MemoryConfig),
		associations: associations,
		analyzer:     NewAssociationAnalyzer(associations),
		counter:      time.Now().UnixNano(), // Use timestamp as base counter
	}
}

// CaptureContext implements the MCP interface for capturing context
func (vj *VectorJournal) CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) (*models.MemoryEntry, error) {
	vj.counter++
	
	// Generate embedding for the content
	embedding, err := vj.llmClient.GenerateEmbedding(ctx, content)
	if err != nil {
		slog.Error("Failed to generate embedding", "error", err, "source", source)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create memory entry
	entry := &models.MemoryEntry{
		ID:            uuid.New().String(),
		Type:          models.TypeEpisodic,
		Content:       content,
		Embedding:     embedding,
		Metadata:      metadata,
		CreatedAt:     time.Now(),
		AccessedAt:    time.Now(),
		Strength:      1.0, // New memories start with full strength
		AssociationIDs: []string{}, // Initialize empty associations
	}
	
	// Add source to metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}
	entry.Metadata["source"] = source
	entry.Metadata["captured_at"] = time.Now().Unix()
	
	// Initialize memory scoring
	entry.Score = vj.scorer.ScoreMemory(entry)
	
	// Store in vector database
	if err := vj.vectorDB.Store(ctx, entry); err != nil {
		slog.Error("Failed to store memory in vector database", "error", err, "id", entry.ID)
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}
	
	slog.Info("Memory captured and stored",
		"source", source,
		"id", entry.ID,
		"content_length", len(content),
		"embedding_dim", len(embedding))
	
	// Analyze associations with recent memories
	go vj.analyzeNewMemoryAssociations(ctx, entry)
	
	return entry, nil
}

// GetMemories retrieves recent memories from episodic storage
func (vj *VectorJournal) GetMemories(ctx context.Context, limit uint64) ([]*models.MemoryEntry, error) {
	if limit == 0 {
		limit = vj.config.BatchSize
	}

	// Create a dummy vector for recent memories query (we'll improve this in Session 3)
	dummyVector := make([]float32, 1536) // Standard embedding dimension
	
	memories, err := vj.vectorDB.Query(ctx, models.TypeEpisodic, dummyVector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}

	return memories, nil
}

// GetMemoryByID retrieves a specific memory by ID
func (vj *VectorJournal) GetMemoryByID(ctx context.Context, id string) (*models.MemoryEntry, error) {
	entry, err := vj.vectorDB.Retrieve(ctx, models.TypeEpisodic, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memory %s: %w", id, err)
	}

	// Update access tracking using enhanced scoring system
	vj.scorer.UpdateMemoryAccess(entry)
	
	// Store updated memory with new score back to database
	if err := vj.vectorDB.Store(ctx, entry); err != nil {
		slog.Warn("Failed to update memory access tracking", "error", err, "id", entry.ID)
	}
	
	return entry, nil
}

// QuerySimilarMemories finds memories similar to the given content
func (vj *VectorJournal) QuerySimilarMemories(ctx context.Context, content string, memType models.MemoryType, limit uint64) ([]*models.MemoryEntry, error) {
	// Generate embedding for the query content
	embedding, err := vj.llmClient.GenerateEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if limit == 0 {
		limit = 10 // Default limit
	}

	// Query vector database
	memories, err := vj.vectorDB.Query(ctx, memType, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar memories: %w", err)
	}

	return memories, nil
}

// BatchStoreMemories stores multiple memories efficiently
func (vj *VectorJournal) BatchStoreMemories(ctx context.Context, entries []*models.MemoryEntry) error {
	batchSize := int(vj.config.BatchSize)
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
			if err := vj.vectorDB.Store(ctx, entry); err != nil {
				slog.Error("Failed to store memory in batch", "error", err, "id", entry.ID)
				return fmt.Errorf("failed to store memory %s: %w", entry.ID, err)
			}
		}

		slog.Debug("Stored memory batch", "count", len(batch), "total_processed", end)
	}

	return nil
}

// ConsolidateMemories processes episodic memories into semantic knowledge
func (vj *VectorJournal) ConsolidateMemories(ctx context.Context, memories []*models.MemoryEntry) error {
	if len(memories) == 0 {
		return nil
	}

	// Extract content for consolidation
	memoryTexts := make([]string, len(memories))
	for i, mem := range memories {
		memoryTexts[i] = mem.Content
	}

	// Use LLM to consolidate memories
	consolidatedContent, err := vj.llmClient.ConsolidateMemories(ctx, memoryTexts)
	if err != nil {
		return fmt.Errorf("failed to consolidate memories: %w", err)
	}

	// Generate embedding for consolidated content
	embedding, err := vj.llmClient.GenerateEmbedding(ctx, consolidatedContent)
	if err != nil {
		return fmt.Errorf("failed to generate embedding for consolidated memory: %w", err)
	}

	// Create semantic memory entry
	semanticEntry := &models.MemoryEntry{
		ID:        uuid.New().String(),
		Type:      models.TypeSemantic,
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
	if err := vj.vectorDB.Store(ctx, semanticEntry); err != nil {
		return fmt.Errorf("failed to store semantic memory: %w", err)
	}

	slog.Info("Consolidated memories into semantic knowledge",
		"semantic_id", semanticEntry.ID,
		"source_count", len(memories),
		"content_length", len(consolidatedContent))

	return nil
}

// GetMemoryStats returns statistics about stored memories
func (vj *VectorJournal) GetMemoryStats(ctx context.Context) (map[string]any, error) {
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
func (vj *VectorJournal) HealthCheck(ctx context.Context) error {
	// Check Qdrant connectivity
	if err := vj.vectorDB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("vector database health check failed: %w", err)
	}

	// Check Ollama connectivity
	if err := vj.llmClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("LLM client health check failed: %w", err)
	}

	return nil
}

// extractMemoryIDs extracts IDs from memory entries for metadata
func extractMemoryIDs(memories []*models.MemoryEntry) []string {
	ids := make([]string, len(memories))
	for i, mem := range memories {
		ids[i] = mem.ID
	}
	return ids
}

// analyzeNewMemoryAssociations runs association analysis for a newly created memory
func (vj *VectorJournal) analyzeNewMemoryAssociations(ctx context.Context, newMemory *models.MemoryEntry) {
	// Get recent memories for association analysis
	recentMemories, err := vj.GetMemories(ctx, 100) // Get last 100 memories
	if err != nil {
		slog.Warn("Failed to get recent memories for association analysis", "error", err)
		return
	}
	
	// Analyze temporal associations (memories within 1 hour)
	vj.analyzer.AnalyzeTemporalAssociations(ctx, newMemory, recentMemories, time.Hour)
	
	// Analyze semantic associations (similarity threshold 0.8)
	vj.analyzer.AnalyzeSemanticAssociations(ctx, newMemory, recentMemories, 0.8)
	
	// Analyze contextual associations (same source)
	vj.analyzer.AnalyzeContextualAssociations(ctx, newMemory, recentMemories)
	
	// Update memory with association IDs
	associationIDs := vj.associations.GetRelatedMemoryIDs(newMemory.ID)
	if len(associationIDs) > 0 {
		newMemory.AssociationIDs = associationIDs
		// Store updated memory (fire and forget, don't block on errors)
		if err := vj.vectorDB.Store(ctx, newMemory); err != nil {
			slog.Warn("Failed to update memory with associations", "error", err, "id", newMemory.ID)
		}
	}
	
	slog.Info("Association analysis complete",
		"memory_id", newMemory.ID,
		"associations_found", len(associationIDs))
}

// GetMemoryWithAssociations retrieves a memory and its associated memories
func (vj *VectorJournal) GetMemoryWithAssociations(ctx context.Context, id string) (*models.MemoryEntry, []*models.MemoryEntry, error) {
	// Get the main memory
	memory, err := vj.GetMemoryByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	
	// Get associated memory IDs
	associatedIDs := vj.associations.GetRelatedMemoryIDs(id)
	
	// Retrieve associated memories
	associatedMemories := make([]*models.MemoryEntry, 0, len(associatedIDs))
	for _, assocID := range associatedIDs {
		// Try to retrieve each associated memory (could be different types)
		for _, memType := range []models.MemoryType{models.TypeEpisodic, models.TypeSemantic, models.TypeProcedural, models.TypeMetacognitive} {
			assocMemory, err := vj.vectorDB.Retrieve(ctx, memType, assocID)
			if err == nil {
				associatedMemories = append(associatedMemories, assocMemory)
				break // Found it, move to next ID
			}
		}
	}
	
	slog.Info("Retrieved memory with associations",
		"memory_id", id,
		"associated_count", len(associatedMemories))
	
	return memory, associatedMemories, nil
}