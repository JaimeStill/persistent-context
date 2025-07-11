package vectordb

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/types"
)

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	// Initialize sets up the vector database (collections, etc.)
	Initialize(ctx context.Context) error
	
	// Store saves a memory entry to the vector database
	Store(ctx context.Context, entry *types.MemoryEntry) error
	
	// Query performs vector similarity search
	Query(ctx context.Context, memType types.MemoryType, vector []float32, limit uint64) ([]*types.MemoryEntry, error)
	
	// Retrieve gets a specific memory entry by ID
	Retrieve(ctx context.Context, memType types.MemoryType, id string) (*types.MemoryEntry, error)
	
	// HealthCheck verifies the database is accessible
	HealthCheck(ctx context.Context) error
}

// Config holds configuration for vector database implementations
type Config struct {
	Provider        string            `mapstructure:"provider"`
	URL             string            `mapstructure:"url"`
	APIKey          string            `mapstructure:"api_key"`
	CollectionNames map[string]string `mapstructure:"collection_names"`
	VectorDimension int               `mapstructure:"vector_dimension"`
	OnDiskPayload   bool              `mapstructure:"on_disk_payload"`
	Insecure        bool              `mapstructure:"insecure"`
}

// NewVectorDB creates a new VectorDB implementation based on the provider
func NewVectorDB(config *Config) (VectorDB, error) {
	switch config.Provider {
	case "qdrant":
		return NewQdrantDB(config)
	default:
		return nil, fmt.Errorf("unsupported vector database provider: %s", config.Provider)
	}
}