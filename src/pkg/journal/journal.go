package journal

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/llm"
	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/JaimeStill/persistent-context/pkg/vectordb"
)

// Journal defines the interface for LLM memory journal storage and retrieval operations
type Journal interface {
	// CaptureContext captures and stores a new memory from context
	CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) (*models.MemoryEntry, error)
	
	// GetMemories retrieves memories with pagination
	GetMemories(ctx context.Context, limit uint32) ([]*models.MemoryEntry, error)
	
	// GetMemoryByID retrieves a specific memory by ID
	GetMemoryByID(ctx context.Context, id string) (*models.MemoryEntry, error)
	
	// QuerySimilarMemories finds similar memories using vector similarity
	QuerySimilarMemories(ctx context.Context, content string, memType models.MemoryType, limit uint64) ([]*models.MemoryEntry, error)
	
	// ConsolidateMemories consolidates episodic memories into semantic knowledge
	ConsolidateMemories(ctx context.Context, memories []*models.MemoryEntry) error
	
	// GetMemoryStats returns statistics about stored memories
	GetMemoryStats(ctx context.Context) (map[string]any, error)
	
	// HealthCheck verifies the journal is accessible
	HealthCheck(ctx context.Context) error
}

// Dependencies holds the dependencies for Journal implementations
type Dependencies struct {
	VectorDB            vectordb.VectorDB
	LLMClient           llm.LLM
	Config              *config.JournalConfig
	MemoryConfig        *config.MemoryConfig
	VectorDBConfig      *config.VectorDBConfig
}

// Validate ensures all required dependencies are present
func (deps *Dependencies) Validate() error {
	if deps.VectorDB == nil {
		return fmt.Errorf("vectorDB is required")
	}
	if deps.LLMClient == nil {
		return fmt.Errorf("llmClient is required")
	}
	if deps.Config == nil {
		return fmt.Errorf("journal config is required")
	}
	if deps.MemoryConfig == nil {
		return fmt.Errorf("memory config is required for memory scoring functionality")
	}
	if deps.VectorDBConfig == nil {
		return fmt.Errorf("vectordb config is required for vector dimension configuration")
	}
	return nil
}

// NewJournal creates a new Journal implementation
func NewJournal(deps *Dependencies) Journal {
	return NewVectorJournal(deps)
}