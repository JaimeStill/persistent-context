package vectordb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/qdrant/go-client/qdrant"
)

// QdrantDB implements vector database operations using Qdrant
type QdrantDB struct {
	client           *qdrant.Client
	config           *config.VectorDBConfig
	memoryCollections map[models.MemoryType]string
	memories         *qdrantMemoryCollection
	associations     *qdrantAssociationCollection
}

// NewQdrantDB creates a new Qdrant database implementation
func NewQdrantDB(config *config.VectorDBConfig) (*QdrantDB, error) {
	host, port := parseGRPCAddress(config.URL)

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   host,
		Port:   port,
		UseTLS: !config.Insecure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	qc := &QdrantDB{
		client:           client,
		config:           config,
		memoryCollections: make(map[models.MemoryType]string),
	}

	// Map memory types to collection names
	for memType, collectionName := range config.MemoryCollections {
		qc.memoryCollections[models.MemoryType(memType)] = collectionName
	}

	// Initialize collections
	qc.memories = newQdrantMemoryCollection(client, config)
	qc.associations = newQdrantAssociationCollection(client, config.AssociationsCollection)

	return qc, nil
}

// Initialize sets up collections and ensures they exist
func (qc *QdrantDB) Initialize(ctx context.Context) error {
	for memType, collectionName := range qc.memoryCollections {
		exists, err := collectionExists(ctx, qc.client, collectionName)
		if err != nil {
			return fmt.Errorf("failed to check collection %s: %w", collectionName, err)
		}

		if !exists {
			if err := createCollection(ctx, qc.client, collectionName, qc.config.VectorDimension, qc.config.OnDiskPayload); err != nil {
				return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
			}
			slog.Info("Created collection", "collection", collectionName, "type", memType)

			// Create payload index for created_at field to support GetRecent() ordering
			if err := createPayloadIndex(ctx, qc.client, collectionName); err != nil {
				return fmt.Errorf("failed to create payload index for collection %s: %w", collectionName, err)
			}
			slog.Info("Created payload index for created_at", "collection", collectionName)
		}
	}

	// Initialize association collection
	associationCollectionName := qc.config.AssociationsCollection
	exists, err := collectionExists(ctx, qc.client, associationCollectionName)
	if err != nil {
		return fmt.Errorf("failed to check association collection %s: %w", associationCollectionName, err)
	}

	if !exists {
		// Create association collection with minimal vector dimension (associations don't need semantic search)
		if err := createCollection(ctx, qc.client, associationCollectionName, 1, qc.config.OnDiskPayload); err != nil {
			return fmt.Errorf("failed to create association collection %s: %w", associationCollectionName, err)
		}
		slog.Info("Created association collection", "collection", associationCollectionName)
	}

	return nil
}

// Store stores a memory entry in the appropriate collection
func (qc *QdrantDB) Store(ctx context.Context, entry *models.MemoryEntry) error {
	collectionName, exists := qc.memoryCollections[entry.Type]
	if !exists {
		return fmt.Errorf("no collection configured for memory type: %s", entry.Type)
	}

	points := []*qdrant.PointStruct{
		{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: entry.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Data: entry.Embedding}}},
			Payload: map[string]*qdrant.Value{
				"content":     qdrant.NewValueString(entry.Content),
				"type":        qdrant.NewValueString(string(entry.Type)),
				"created_at":  qdrant.NewValueInt(entry.CreatedAt.Unix()),
				"accessed_at": qdrant.NewValueString(entry.AccessedAt.Format(time.RFC3339)),
				"strength":    qdrant.NewValueDouble(float64(entry.Strength)),
			},
		},
	}

	// Add metadata to payload
	for key, value := range entry.Metadata {
		switch v := value.(type) {
		case string:
			points[0].Payload[key] = qdrant.NewValueString(v)
		case int:
			points[0].Payload[key] = qdrant.NewValueInt(int64(v))
		case int64:
			points[0].Payload[key] = qdrant.NewValueInt(v)
		case float64:
			points[0].Payload[key] = qdrant.NewValueDouble(v)
		case bool:
			points[0].Payload[key] = qdrant.NewValueBool(v)
		default:
			points[0].Payload[key] = qdrant.NewValueString(fmt.Sprintf("%v", v))
		}
	}

	_, err := qc.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	return err
}

// Query performs a vector similarity search
func (qc *QdrantDB) Query(ctx context.Context, memType models.MemoryType, vector []float32, limit uint64) ([]*models.MemoryEntry, error) {
	collectionName, exists := qc.memoryCollections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qc.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
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
func (qc *QdrantDB) Retrieve(ctx context.Context, memType models.MemoryType, id string) (*models.MemoryEntry, error) {
	collectionName, exists := qc.memoryCollections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qc.client.Get(ctx, &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids:            []*qdrant.PointId{{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}}},
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve point %s: %w", id, err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("memory entry not found: %s", id)
	}

	return retrievedPointToMemoryEntry(response[0])
}

// GetRecent retrieves recent memories by creation time without similarity search
func (qc *QdrantDB) GetRecent(ctx context.Context, memType models.MemoryType, limit uint32) ([]*models.MemoryEntry, error) {
	collectionName, exists := qc.memoryCollections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	direction := qdrant.Direction_Desc
	response, err := qc.client.Scroll(ctx, &qdrant.ScrollPoints{
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
func (qc *QdrantDB) Count(ctx context.Context, memType models.MemoryType) (uint64, error) {
	collectionName, exists := qc.memoryCollections[memType]
	if !exists {
		return 0, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qc.client.Count(ctx, &qdrant.CountPoints{
		CollectionName: collectionName,
		Exact:          &[]bool{true}[0], // Use exact count
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count collection %s: %w", collectionName, err)
	}

	return response, nil
}

// Delete removes memories by their IDs
func (qc *QdrantDB) Delete(ctx context.Context, memType models.MemoryType, ids []string) error {
	collectionName, exists := qc.memoryCollections[memType]
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

	_, err := qc.client.Delete(ctx, &qdrant.DeletePoints{
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
func (qc *QdrantDB) GetAll(ctx context.Context, memType models.MemoryType, cursor string, limit uint32) (entries []*models.MemoryEntry, nextCursor string, err error) {
	collectionName, exists := qc.memoryCollections[memType]
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

	response, err := qc.client.Scroll(ctx, scrollRequest)
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

	// Set nextCursor for pagination
	// If we got fewer results than requested, we're at the end
	if len(entries) < int(limit) {
		nextCursor = ""
	} else if len(entries) > 0 {
		// Use the last entry's ID as the cursor for the next page
		nextCursor = entries[len(entries)-1].ID
	}

	return entries, nextCursor, nil
}

// HealthCheck checks if the Qdrant server is healthy
func (qc *QdrantDB) HealthCheck(ctx context.Context) error {
	_, err := qc.client.HealthCheck(ctx)
	return err
}

// Memories returns the memory collection interface
func (qc *QdrantDB) Memories() MemoryCollection {
	return qc.memories
}

// Associations returns the association collection interface
func (qc *QdrantDB) Associations() AssociationCollection {
	return qc.associations
}
