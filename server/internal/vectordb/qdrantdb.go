package vectordb

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// QdrantDB implements vector database operations using Qdrant
type QdrantDB struct {
	client      *qdrant.Client
	config      *Config
	collections map[types.MemoryType]string
}

// NewQdrantDB creates a new Qdrant database implementation
func NewQdrantDB(config *Config) (*QdrantDB, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: extractHost(config.URL),
		Port: extractPort(config.URL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	qc := &QdrantDB{
		client:      client,
		config:      config,
		collections: make(map[types.MemoryType]string),
	}

	// Map memory types to collection names
	for memType, collectionName := range config.CollectionNames {
		qc.collections[types.MemoryType(memType)] = collectionName
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
		}
	}

	return nil
}

// Store stores a memory entry in the appropriate collection
func (qc *QdrantDB) Store(ctx context.Context, entry *types.MemoryEntry) error {
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
				"created_at":  qdrant.NewValueString(entry.CreatedAt.Format(time.RFC3339)),
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
func (qc *QdrantDB) Query(ctx context.Context, memType types.MemoryType, vector []float32, limit uint64) ([]*types.MemoryEntry, error) {
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

	entries := make([]*types.MemoryEntry, 0, len(response))
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
func (qc *QdrantDB) Retrieve(ctx context.Context, memType types.MemoryType, id string) (*types.MemoryEntry, error) {
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

// retrievedPointToMemoryEntry converts a Qdrant RetrievedPoint to a memory entry
func (qc *QdrantDB) retrievedPointToMemoryEntry(point *qdrant.RetrievedPoint) (*types.MemoryEntry, error) {
	entry := &types.MemoryEntry{
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
			entry.Type = types.MemoryType(memType.GetStringValue())
		}
		if createdAt := payload["created_at"]; createdAt != nil {
			if t, err := time.Parse(time.RFC3339, createdAt.GetStringValue()); err == nil {
				entry.CreatedAt = t
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
func (qc *QdrantDB) scoredPointToMemoryEntry(scoredPoint *qdrant.ScoredPoint) (*types.MemoryEntry, error) {
	entry := &types.MemoryEntry{
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
			entry.Type = types.MemoryType(memType.GetStringValue())
		}
		if createdAt := payload["created_at"]; createdAt != nil {
			if t, err := time.Parse(time.RFC3339, createdAt.GetStringValue()); err == nil {
				entry.CreatedAt = t
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
func extractHost(url string) string {
	// Simple extraction - assumes format http://host:port
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}
	
	if idx := strings.LastIndex(url, ":"); idx != -1 {
		return url[:idx]
	}
	return url
}

func extractPort(url string) int {
	// Simple extraction - assumes format http://host:port
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}
	
	if idx := strings.LastIndex(url, ":"); idx != -1 {
		if port := url[idx+1:]; port != "" {
			if p, err := strconv.Atoi(port); err == nil {
				return p
			}
		}
	}
	return 6333 // Default Qdrant port
}