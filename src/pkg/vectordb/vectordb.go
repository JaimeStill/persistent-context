package vectordb

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/models"
)

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	// Initialize sets up the vector database (collections, etc.)
	Initialize(ctx context.Context) error
	
	// Store saves a memory entry to the vector database
	Store(ctx context.Context, entry *models.MemoryEntry) error
	
	// Query performs vector similarity search (uses uint64 to match QueryPoints)
	Query(ctx context.Context, memType models.MemoryType, vector []float32, limit uint64) ([]*models.MemoryEntry, error)
	
	// Retrieve gets a specific memory entry by ID
	Retrieve(ctx context.Context, memType models.MemoryType, id string) (*models.MemoryEntry, error)
	
	// GetRecent retrieves recent memories by creation time without similarity search (uses uint32 to match ScrollPoints)
	GetRecent(ctx context.Context, memType models.MemoryType, limit uint32) ([]*models.MemoryEntry, error)
	
	// Count returns the number of memories of a specific type (returns uint64 to match CountResult)
	Count(ctx context.Context, memType models.MemoryType) (uint64, error)
	
	// Delete removes memories by their IDs
	Delete(ctx context.Context, memType models.MemoryType, ids []string) error
	
	// GetAll retrieves all memories with cursor-based pagination (efficient for large datasets)
	// cursor: empty string for first page, or value from previous response's nextCursor
	// Returns: entries, nextCursor (empty if no more results), error
	GetAll(ctx context.Context, memType models.MemoryType, cursor string, limit uint32) (entries []*models.MemoryEntry, nextCursor string, err error)
	
	// HealthCheck verifies the database is accessible
	HealthCheck(ctx context.Context) error
}

// NewVectorDB creates a new VectorDB implementation based on the provider
func NewVectorDB(config *config.VectorDBConfig) (VectorDB, error) {
	switch config.Provider {
	case "qdrant":
		return NewQdrantDB(config)
	default:
		return nil, fmt.Errorf("unsupported vector database provider: %s", config.Provider)
	}
}