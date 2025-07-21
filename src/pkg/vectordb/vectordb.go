package vectordb

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/pkg/config"
)

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	// Initialize sets up the vector database (collections, etc.)
	Initialize(ctx context.Context) error
	
	// HealthCheck verifies the database is accessible
	HealthCheck(ctx context.Context) error
	
	// Memories returns the memory collection interface
	Memories() MemoryCollection
	
	// Associations returns the association collection interface
	Associations() AssociationCollection
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