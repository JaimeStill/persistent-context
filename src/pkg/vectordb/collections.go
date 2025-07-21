package vectordb

import (
	"context"

	"github.com/JaimeStill/persistent-context/pkg/models"
)

// Collection represents generic vector database collection operations
type Collection[T any] interface {
	// Store saves an item to the collection
	Store(ctx context.Context, item T) error
	
	// Retrieve gets a specific item by ID
	Retrieve(ctx context.Context, id string) (T, error)
	
	// Query performs vector similarity search
	Query(ctx context.Context, vector []float32, limit uint64) ([]T, error)
	
	// Delete removes items by their IDs
	Delete(ctx context.Context, ids []string) error
	
	// Count returns the number of items in the collection
	Count(ctx context.Context) (uint64, error)
	
	// GetAll retrieves all items with cursor-based pagination
	GetAll(ctx context.Context, cursor string, limit uint32) (items []T, nextCursor string, err error)
}

// MemoryCollection handles memory-specific operations
type MemoryCollection interface {
	// Store saves a memory entry to the appropriate collection based on its type
	Store(ctx context.Context, entry *models.MemoryEntry) error
	
	// Query performs vector similarity search for a specific memory type
	Query(ctx context.Context, memType models.MemoryType, vector []float32, limit uint64) ([]*models.MemoryEntry, error)
	
	// Retrieve gets a specific memory entry by ID and type
	Retrieve(ctx context.Context, memType models.MemoryType, id string) (*models.MemoryEntry, error)
	
	// GetRecent retrieves recent memories by creation time without similarity search
	GetRecent(ctx context.Context, memType models.MemoryType, limit uint32) ([]*models.MemoryEntry, error)
	
	// Count returns the number of memories of a specific type
	Count(ctx context.Context, memType models.MemoryType) (uint64, error)
	
	// Delete removes memories by their IDs from a specific type collection
	Delete(ctx context.Context, memType models.MemoryType, ids []string) error
	
	// GetAll retrieves all memories of a type with cursor-based pagination
	GetAll(ctx context.Context, memType models.MemoryType, cursor string, limit uint32) (entries []*models.MemoryEntry, nextCursor string, err error)
}

// AssociationCollection handles association-specific operations
type AssociationCollection interface {
	// Store saves a single association
	Store(ctx context.Context, association *models.MemoryAssociation) error
	
	// BulkStore saves multiple associations efficiently
	BulkStore(ctx context.Context, associations []*models.MemoryAssociation) error
	
	// GetByMemoryID retrieves all associations for a specific memory
	GetByMemoryID(ctx context.Context, memoryID string) ([]*models.MemoryAssociation, error)
	
	// GetByMemoryIDs retrieves associations for multiple memories
	GetByMemoryIDs(ctx context.Context, memoryIDs []string) (map[string][]*models.MemoryAssociation, error)
	
	// Delete removes specific associations by their IDs
	Delete(ctx context.Context, associationIDs []string) error
	
	// DeleteByMemoryID removes all associations for a specific memory
	DeleteByMemoryID(ctx context.Context, memoryID string) error
	
	// Count returns the total number of associations
	Count(ctx context.Context) (uint64, error)
	
	// GetAll retrieves all associations with pagination
	GetAll(ctx context.Context, cursor string, limit uint32) (associations []*models.MemoryAssociation, nextCursor string, err error)
}