package vectordb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// qdrantMemoryCollection implements MemoryCollection for Qdrant
type qdrantMemoryCollection struct {
	client      *qdrant.Client
	config      *config.VectorDBConfig
	collections map[models.MemoryType]string
}

// newQdrantMemoryCollection creates a new Qdrant memory collection
func newQdrantMemoryCollection(client *qdrant.Client, config *config.VectorDBConfig) *qdrantMemoryCollection {
	qmc := &qdrantMemoryCollection{
		client:      client,
		config:      config,
		collections: make(map[models.MemoryType]string),
	}

	// Map memory types to collection names
	for memType, collectionName := range config.MemoryCollections {
		qmc.collections[models.MemoryType(memType)] = collectionName
	}

	return qmc
}

// Store stores a memory entry in the appropriate collection
func (qmc *qdrantMemoryCollection) Store(ctx context.Context, entry *models.MemoryEntry) error {
	collectionName, exists := qmc.collections[entry.Type]
	if !exists {
		return fmt.Errorf("no collection configured for memory type: %s", entry.Type)
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Create Qdrant point
	points := []*qdrant.PointStruct{
		{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: entry.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Data: entry.Embedding}}},
			Payload: memoryEntryToQdrantPayload(entry),
		},
	}

	_, err := qmc.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	slog.Debug("Stored memory", "id", entry.ID, "type", entry.Type, "collection", collectionName)
	return nil
}

// Query performs a vector similarity search
func (qmc *qdrantMemoryCollection) Query(ctx context.Context, memType models.MemoryType, vector []float32, limit uint64) ([]*models.MemoryEntry, error) {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qmc.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
		WithVectors: &qdrant.WithVectorsSelector{
			SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search collection %s: %w", collectionName, err)
	}

	entries := make([]*models.MemoryEntry, 0, len(response))
	for _, scoredPoint := range response {
		entry, err := scoredPointToMemoryEntry(scoredPoint)
		if err != nil {
			slog.Warn("Failed to convert scored point to memory entry", "error", err)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Retrieve gets a specific memory entry by ID
func (qmc *qdrantMemoryCollection) Retrieve(ctx context.Context, memType models.MemoryType, id string) (*models.MemoryEntry, error) {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qmc.client.Get(ctx, &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids:            []*qdrant.PointId{{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}}},
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memory: %w", err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	return retrievedPointToMemoryEntry(response[0])
}

// GetRecent retrieves recent memories by creation time without similarity search
func (qmc *qdrantMemoryCollection) GetRecent(ctx context.Context, memType models.MemoryType, limit uint32) ([]*models.MemoryEntry, error) {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	direction := qdrant.Direction_Desc
	response, err := qmc.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: collectionName,
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
		OrderBy: &qdrant.OrderBy{
			Key:       "created_at",
			Direction: &direction,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scroll collection %s: %w", collectionName, err)
	}

	entries := make([]*models.MemoryEntry, 0, len(response))
	for _, point := range response {
		entry, err := retrievedPointToMemoryEntry(point)
		if err != nil {
			slog.Warn("Failed to convert retrieved point to memory entry", "error", err)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Count returns the number of memories of a specific type
func (qmc *qdrantMemoryCollection) Count(ctx context.Context, memType models.MemoryType) (uint64, error) {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return 0, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qmc.client.Count(ctx, &qdrant.CountPoints{
		CollectionName: collectionName,
		Exact:          &[]bool{true}[0], // Use exact count
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count collection %s: %w", collectionName, err)
	}

	return response, nil
}

// Delete removes memories by their IDs
func (qmc *qdrantMemoryCollection) Delete(ctx context.Context, memType models.MemoryType, ids []string) error {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	if len(ids) == 0 {
		return nil // Nothing to delete
	}

	// Convert string IDs to PointId structures
	pointIds := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		pointIds[i] = &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}}
	}

	_, err := qmc.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIds,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete points from collection %s: %w", collectionName, err)
	}

	return nil
}

// GetAll retrieves all memories with cursor-based pagination
func (qmc *qdrantMemoryCollection) GetAll(ctx context.Context, memType models.MemoryType, cursor string, limit uint32) (entries []*models.MemoryEntry, nextCursor string, err error) {
	collectionName, exists := qmc.collections[memType]
	if !exists {
		return nil, "", fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	// Build scroll request
	scrollRequest := &qdrant.ScrollPoints{
		CollectionName: collectionName,
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	// If cursor is provided, use it as the offset point
	if cursor != "" {
		scrollRequest.Offset = &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: cursor}}
	}

	// Add ordering for consistent pagination
	direction := qdrant.Direction_Desc
	scrollRequest.OrderBy = &qdrant.OrderBy{
		Key:       "created_at",
		Direction: &direction,
	}

	response, err := qmc.client.Scroll(ctx, scrollRequest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to scroll collection %s: %w", collectionName, err)
	}

	// Convert points to memory entries
	entries = make([]*models.MemoryEntry, 0, len(response))
	for _, point := range response {
		entry, err := retrievedPointToMemoryEntry(point)
		if err != nil {
			slog.Warn("Failed to convert retrieved point to memory entry", "error", err)
			continue
		}
		entries = append(entries, entry)
	}

	// Determine next cursor - use the last entry's ID if there might be more results
	if len(entries) == int(limit) && len(entries) > 0 {
		nextCursor = entries[len(entries)-1].ID
	}

	return entries, nextCursor, nil
}
