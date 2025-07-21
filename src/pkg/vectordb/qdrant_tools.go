package vectordb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/models"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// retrievedPointToMemoryEntry converts a Qdrant RetrievedPoint to a memory entry
func retrievedPointToMemoryEntry(retrievedPoint *qdrant.RetrievedPoint) (*models.MemoryEntry, error) {
	entry := &models.MemoryEntry{
		ID:       retrievedPoint.Id.GetUuid(),
		Metadata: make(map[string]any),
	}

	// Extract vector
	if vectors := retrievedPoint.Vectors; vectors != nil {
		if vector := vectors.GetVector(); vector != nil {
			entry.Embedding = vector.Data
		}
	}

	// Extract payload
	if payload := retrievedPoint.Payload; payload != nil {
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
		if associationIDs := payload["association_ids"]; associationIDs != nil {
			if listVal := associationIDs.GetListValue(); listVal != nil {
				entry.AssociationIDs = make([]string, 0, len(listVal.GetValues()))
				for _, val := range listVal.GetValues() {
					if strVal := val.GetStringValue(); strVal != "" {
						entry.AssociationIDs = append(entry.AssociationIDs, strVal)
					}
				}
			}
		}

		// Extract metadata
		for key, value := range payload {
			if key == "content" || key == "type" || key == "created_at" || key == "accessed_at" || key == "strength" || key == "association_ids" {
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
func scoredPointToMemoryEntry(scoredPoint *qdrant.ScoredPoint) (*models.MemoryEntry, error) {
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
		if associationIDs := payload["association_ids"]; associationIDs != nil {
			if listVal := associationIDs.GetListValue(); listVal != nil {
				entry.AssociationIDs = make([]string, 0, len(listVal.GetValues()))
				for _, val := range listVal.GetValues() {
					if strVal := val.GetStringValue(); strVal != "" {
						entry.AssociationIDs = append(entry.AssociationIDs, strVal)
					}
				}
			}
		}

		// Extract metadata
		for key, value := range payload {
			if key == "content" || key == "type" || key == "created_at" || key == "accessed_at" || key == "strength" || key == "association_ids" {
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

// memoryEntryToQdrantPayload converts a MemoryEntry to Qdrant payload
func memoryEntryToQdrantPayload(entry *models.MemoryEntry) map[string]*qdrant.Value {
	payload := map[string]*qdrant.Value{
		"content":    {Kind: &qdrant.Value_StringValue{StringValue: entry.Content}},
		"type":       {Kind: &qdrant.Value_StringValue{StringValue: string(entry.Type)}},
		"created_at": {Kind: &qdrant.Value_IntegerValue{IntegerValue: entry.CreatedAt.Unix()}},
		"strength":   {Kind: &qdrant.Value_DoubleValue{DoubleValue: float64(entry.Strength)}},
	}

	// Add accessed_at if not zero
	if !entry.AccessedAt.IsZero() {
		payload["accessed_at"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.AccessedAt.Format(time.RFC3339)}}
	}

	// Add association IDs if present
	if len(entry.AssociationIDs) > 0 {
		values := make([]*qdrant.Value, len(entry.AssociationIDs))
		for i, id := range entry.AssociationIDs {
			values[i] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: id}}
		}
		payload["association_ids"] = &qdrant.Value{Kind: &qdrant.Value_ListValue{ListValue: &qdrant.ListValue{Values: values}}}
	}

	// Add metadata
	for key, value := range entry.Metadata {
		payload[key] = anyToQdrantValue(value)
	}

	return payload
}

// anyToQdrantValue converts a Go value to Qdrant Value
func anyToQdrantValue(v any) *qdrant.Value {
	switch val := v.(type) {
	case string:
		return &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: val}}
	case int:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(val)}}
	case int64:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: val}}
	case float32:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: float64(val)}}
	case float64:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: val}}
	case bool:
		return &qdrant.Value{Kind: &qdrant.Value_BoolValue{BoolValue: val}}
	default:
		// Fallback: convert to string
		return &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: fmt.Sprintf("%v", val)}}
	}
}

// collectionExists checks if a collection exists
func collectionExists(ctx context.Context, client *qdrant.Client, name string) (bool, error) {
	response, err := client.ListCollections(ctx)
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
func createCollection(ctx context.Context, client *qdrant.Client, name string, vectorDimension int, onDiskPayload bool) error {
	err := client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(vectorDimension),
					Distance: qdrant.Distance_Cosine,
					OnDisk:   &onDiskPayload,
				},
			},
		},
	})
	return err
}

// createPayloadIndex creates a payload index for the created_at field
func createPayloadIndex(ctx context.Context, client *qdrant.Client, collectionName string) error {
	fieldType := qdrant.FieldType_FieldTypeInteger
	rangeEnabled := true
	
	_, err := client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
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

// timeFromUnix converts Unix timestamp to time.Time
func timeFromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}