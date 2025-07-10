package vectordb

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/JaimeStill/persistent-context/internal/memory"
)

// QdrantClient implements vector database operations using Qdrant
type QdrantClient struct {
	client      *qdrant.Client
	config      *Config
	collections map[memory.MemoryType]string
}

// NewQdrantClient creates a new Qdrant client
func NewQdrantClient(config *Config) (*QdrantClient, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: extractHost(config.URL),
		Port: extractPort(config.URL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	qc := &QdrantClient{
		client:      client,
		config:      config,
		collections: make(map[memory.MemoryType]string),
	}

	// Map memory types to collection names
	for memType, collectionName := range config.CollectionNames {
		qc.collections[memory.MemoryType(memType)] = collectionName
	}

	return qc, nil
}

// Initialize sets up collections and ensures they exist
func (qc *QdrantClient) Initialize(ctx context.Context) error {
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
func (qc *QdrantClient) Store(ctx context.Context, entry *memory.MemoryEntry) error {
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
			points[0].Payload[key] = qdrant.NewValueInteger(int64(v))
		case int64:
			points[0].Payload[key] = qdrant.NewValueInteger(v)
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
func (qc *QdrantClient) Query(ctx context.Context, memType memory.MemoryType, vector []float32, limit int) ([]*memory.MemoryEntry, error) {
	collectionName, exists := qc.collections[memType]
	if !exists {
		return nil, fmt.Errorf("no collection configured for memory type: %s", memType)
	}

	response, err := qc.client.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Limit:          uint64(limit),
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search collection %s: %w", collectionName, err)
	}

	entries := make([]*memory.MemoryEntry, 0, len(response.Result))
	for _, point := range response.Result {
		entry, err := qc.pointToMemoryEntry(point)
		if err != nil {
			slog.Warn("Failed to convert point to memory entry", "error", err)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Retrieve gets a specific memory entry by ID
func (qc *QdrantClient) Retrieve(ctx context.Context, memType memory.MemoryType, id string) (*memory.MemoryEntry, error) {
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

	if len(response.Result) == 0 {
		return nil, fmt.Errorf("memory entry not found: %s", id)
	}

	return qc.pointToMemoryEntry(response.Result[0])
}

// HealthCheck checks if the Qdrant server is healthy
func (qc *QdrantClient) HealthCheck(ctx context.Context) error {
	_, err := qc.client.HealthCheck(ctx)
	return err
}

// collectionExists checks if a collection exists
func (qc *QdrantClient) collectionExists(ctx context.Context, name string) (bool, error) {
	response, err := qc.client.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, collection := range response.Collections {
		if collection.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// createCollection creates a new collection
func (qc *QdrantClient) createCollection(ctx context.Context, name string) error {
	_, err := qc.client.CreateCollection(ctx, &qdrant.CreateCollection{
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

// pointToMemoryEntry converts a Qdrant point to a memory entry
func (qc *QdrantClient) pointToMemoryEntry(point *qdrant.RetrievedPoint) (*memory.MemoryEntry, error) {
	entry := &memory.MemoryEntry{
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
			entry.Type = memory.MemoryType(memType.GetStringValue())
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