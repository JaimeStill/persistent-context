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
	
	// Query performs vector similarity search
	Query(ctx context.Context, memType models.MemoryType, vector []float32, limit uint64) ([]*models.MemoryEntry, error)
	
	// Retrieve gets a specific memory entry by ID
	Retrieve(ctx context.Context, memType models.MemoryType, id string) (*models.MemoryEntry, error)
	
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