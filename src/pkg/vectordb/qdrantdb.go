package vectordb

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/models"
)

// QdrantDB implements vector database operations using Qdrant
type QdrantDB struct {
	client      *qdrant.Client
	config      *config.VectorDBConfig
	collections map[models.MemoryType]string
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
		client:      client,
		config:      config,
		collections: make(map[models.MemoryType]string),
	}

	// Map memory types to collection names
	for memType, collectionName := range config.CollectionNames {
		qc.collections[models.MemoryType(memType)] = collectionName
	}

	return qc, nil
}

// Initialize sets up collections and ensures they exist
func (qc *QdrantDB) Initialize(ctx context.Context) error {
	for memType, collectionName := range qc.collections {
		exists, err := qc.collectionExists(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to check collection %s: %w", collectionName, err)
		}

		if !exists {
			if err := qc.createCollection(ctx, collectionName); err != nil {
				return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
			}
			slog.Info("Created collection", "collection", collectionName, "type", memType)
			
			// Create payload index for created_at field to support GetRecent() ordering
			if err := qc.createPayloadIndex(ctx, collectionName); err != nil {
				return fmt.Errorf("failed to create payload index for collection %s: %w", collectionName, err)
			}
			slog.Info("Created payload index for created_at", "collection", collectionName)
		}
	}

	return nil
}

// Store stores a memory entry in the appropriate collection
func (qc *QdrantDB) Store(ctx context.Context, entry *models.MemoryEntry) error {
	collectionName, exists := qc.collections[entry.Type]
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
	collectionName, exists := qc.collections[memType]
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
		entry, err := qc.scoredPointToMemoryEntry(scoredPoint)
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
	collectionName, exists := qc.collections[memType]
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

	return qc.retrievedPointToMemoryEntry(response[0])
}

// GetRecent retrieves recent memories by creation time without similarity search
func (qc *QdrantDB) GetRecent(ctx context.Context, memType models.MemoryType, limit uint32) ([]*models.MemoryEntry, error) {
	collectionName, exists := qc.collections[memType]
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
		entry, err := qc.retrievedPointToMemoryEntry(point)
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
	collectionName, exists := qc.collections[memType]
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
	collectionName, exists := qc.collections[memType]
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
	collectionName, exists := qc.collections[memType]
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
		entry, err := qc.retrievedPointToMemoryEntry(point)
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

// collectionExists checks if a collection exists
func (qc *QdrantDB) collectionExists(ctx context.Context, name string) (bool, error) {
	response, err := qc.client.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, collection := range response {
		if collection == name {
			return true, nil
		}
	}
	return false, nil
}

// createCollection creates a new collection
func (qc *QdrantDB) createCollection(ctx context.Context, name string) error {
	err := qc.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(qc.config.VectorDimension),
					Distance: qdrant.Distance_Cosine,
					OnDisk:   &qc.config.OnDiskPayload,
				},
			},
		},
	})
	return err
}

// createPayloadIndex creates a payload index for the created_at field
func (qc *QdrantDB) createPayloadIndex(ctx context.Context, collectionName string) error {
	fieldType := qdrant.FieldType_FieldTypeInteger
	rangeEnabled := true
	
	_, err := qc.client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "created_at",
		FieldType:      &fieldType, // Unix timestamp
		FieldIndexParams: &qdrant.PayloadIndexParams{
			IndexParams: &qdrant.PayloadIndexParams_IntegerIndexParams{
				IntegerIndexParams: &qdrant.IntegerIndexParams{
					Range: &rangeEnabled, // Enable range queries for ordering
				},
			},
		},
	})
	return err
}

// retrievedPointToMemoryEntry converts a Qdrant RetrievedPoint to a memory entry
func (qc *QdrantDB) retrievedPointToMemoryEntry(point *qdrant.RetrievedPoint) (*models.MemoryEntry, error) {
	entry := &models.MemoryEntry{
		ID:       point.Id.GetUuid(),
		Metadata: make(map[string]any),
	}

	// Extract vector
	if vectors := point.Vectors; vectors != nil {
		if vector := vectors.GetVector(); vector != nil {
			entry.Embedding = vector.Data
		}
	}

	// Extract payload
	if payload := point.Payload; payload != nil {
		if content := payload["content"]; content != nil {
			entry.Content = content.GetStringValue()
		}
		if memType := payload["type"]; memType != nil {
			entry.Type = models.MemoryType(memType.GetStringValue())
		}
		if createdAt := payload["created_at"]; createdAt != nil {
			if timestamp := createdAt.GetIntegerValue(); timestamp != 0 {
				entry.CreatedAt = time.Unix(timestamp, 0)
			}
		}
		if accessedAt := payload["accessed_at"]; accessedAt != nil {
			if t, err := time.Parse(time.RFC3339, accessedAt.GetStringValue()); err == nil {
				entry.AccessedAt = t
			}
		}
		if strength := payload["strength"]; strength != nil {
			entry.Strength = float32(strength.GetDoubleValue())
		}

		// Extract metadata
		for key, value := range payload {
			if key == "content" || key == "type" || key == "created_at" || key == "accessed_at" || key == "strength" {
				continue
			}
			switch v := value.Kind.(type) {
			case *qdrant.Value_StringValue:
				entry.Metadata[key] = v.StringValue
			case *qdrant.Value_IntegerValue:
				entry.Metadata[key] = v.IntegerValue
			case *qdrant.Value_DoubleValue:
				entry.Metadata[key] = v.DoubleValue
			case *qdrant.Value_BoolValue:
				entry.Metadata[key] = v.BoolValue
			}
		}
	}

	return entry, nil
}

// scoredPointToMemoryEntry converts a Qdrant ScoredPoint to a memory entry
func (qc *QdrantDB) scoredPointToMemoryEntry(scoredPoint *qdrant.ScoredPoint) (*models.MemoryEntry, error) {
	entry := &models.MemoryEntry{
		ID:       scoredPoint.Id.GetUuid(),
		Metadata: make(map[string]any),
	}

	// Extract vector
	if vectors := scoredPoint.Vectors; vectors != nil {
		if vector := vectors.GetVector(); vector != nil {
			entry.Embedding = vector.Data
		}
	}

	// Extract payload
	if payload := scoredPoint.Payload; payload != nil {
		if content := payload["content"]; content != nil {
			entry.Content = content.GetStringValue()
		}
		if memType := payload["type"]; memType != nil {
			entry.Type = models.MemoryType(memType.GetStringValue())
		}
		if createdAt := payload["created_at"]; createdAt != nil {
			if timestamp := createdAt.GetIntegerValue(); timestamp != 0 {
				entry.CreatedAt = time.Unix(timestamp, 0)
			}
		}
		if accessedAt := payload["accessed_at"]; accessedAt != nil {
			if t, err := time.Parse(time.RFC3339, accessedAt.GetStringValue()); err == nil {
				entry.AccessedAt = t
			}
		}
		if strength := payload["strength"]; strength != nil {
			entry.Strength = float32(strength.GetDoubleValue())
		}

		// Extract metadata
		for key, value := range payload {
			if key == "content" || key == "type" || key == "created_at" || key == "accessed_at" || key == "strength" {
				continue
			}
			switch v := value.Kind.(type) {
			case *qdrant.Value_StringValue:
				entry.Metadata[key] = v.StringValue
			case *qdrant.Value_IntegerValue:
				entry.Metadata[key] = v.IntegerValue
			case *qdrant.Value_DoubleValue:
				entry.Metadata[key] = v.DoubleValue
			case *qdrant.Value_BoolValue:
				entry.Metadata[key] = v.BoolValue
			}
		}
	}

	return entry, nil
}

// Helper functions to extract host and port from URL
// parseGRPCAddress parses a gRPC address in host:port format
// Supports both "host:port" and "host" (defaults to port 6334)
func parseGRPCAddress(address string) (host string, port int) {
	if idx := strings.LastIndex(address, ":"); idx != -1 {
		host = address[:idx]
		if portStr := address[idx+1:]; portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				return host, p
			}
		}
	}
	// No port specified, use host as-is and default gRPC port
	return address, 6334
}