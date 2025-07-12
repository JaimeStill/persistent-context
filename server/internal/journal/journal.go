package journal

import (
	"context"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// Journal defines the interface for LLM memory journal storage and retrieval operations
type Journal interface {
	// CaptureContext captures and stores a new memory from context
	CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) (*types.MemoryEntry, error)
	
	// GetMemories retrieves memories with pagination
	GetMemories(ctx context.Context, limit uint64) ([]*types.MemoryEntry, error)
	
	// GetMemoryByID retrieves a specific memory by ID
	GetMemoryByID(ctx context.Context, id string) (*types.MemoryEntry, error)
	
	// QuerySimilarMemories finds similar memories using vector similarity
	QuerySimilarMemories(ctx context.Context, content string, memType types.MemoryType, limit uint64) ([]*types.MemoryEntry, error)
	
	// BatchStoreMemories stores multiple memories efficiently
	BatchStoreMemories(ctx context.Context, entries []*types.MemoryEntry) error
	
	// ConsolidateMemories consolidates episodic memories into semantic knowledge
	ConsolidateMemories(ctx context.Context, memories []*types.MemoryEntry) error
	
	// GetMemoryStats returns statistics about stored memories
	GetMemoryStats(ctx context.Context) (map[string]any, error)
	
	// HealthCheck verifies the journal is accessible
	HealthCheck(ctx context.Context) error
}

// Dependencies holds the dependencies for Journal implementations
type Dependencies struct {
	VectorDB  vectordb.VectorDB
	LLMClient llm.LLM
	Config    *config.JournalConfig
}

// NewJournal creates a new Journal implementation
func NewJournal(deps *Dependencies) Journal {
	return NewVectorJournal(deps)
}