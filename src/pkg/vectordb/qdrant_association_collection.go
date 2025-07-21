package vectordb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaimeStill/persistent-context/pkg/models"
	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// qdrantAssociationCollection implements AssociationCollection for Qdrant
type qdrantAssociationCollection struct {
	client         *qdrant.Client
	collectionName string
}

// newQdrantAssociationCollection creates a new Qdrant association collection
func newQdrantAssociationCollection(client *qdrant.Client, collectionName string) *qdrantAssociationCollection {
	return &qdrantAssociationCollection{
		client:         client,
		collectionName: collectionName,
	}
}

// Store saves a single association
func (qac *qdrantAssociationCollection) Store(ctx context.Context, association *models.MemoryAssociation) error {
	if association.ID == "" {
		association.ID = uuid.New().String()
	}

	payload := associationToQdrantPayload(association)
	
	// Associations don't need vector embeddings, use zero vector
	zeroVector := make([]float32, 1) // Minimal 1D vector
	
	points := []*qdrant.PointStruct{
		{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: association.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Data: zeroVector}}},
			Payload: payload,
		},
	}
	
	_, err := qac.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: qac.collectionName,
		Points:         points,
	})
	
	if err != nil {
		return fmt.Errorf("failed to store association: %w", err)
	}
	
	slog.Debug("Stored association", "id", association.ID, "source", association.SourceID, "target", association.TargetID)
	return nil
}

// BulkStore saves multiple associations efficiently
func (qac *qdrantAssociationCollection) BulkStore(ctx context.Context, associations []*models.MemoryAssociation) error {
	if len(associations) == 0 {
		return nil
	}
	
	points := make([]*qdrant.PointStruct, len(associations))
	zeroVector := make([]float32, 1) // Minimal 1D vector
	
	for i, association := range associations {
		if association.ID == "" {
			association.ID = uuid.New().String()
		}
		
		points[i] = &qdrant.PointStruct{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: association.ID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{Vector: &qdrant.Vector{Data: zeroVector}}},
			Payload: associationToQdrantPayload(association),
		}
	}
	
	_, err := qac.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: qac.collectionName,
		Points:         points,
	})
	
	if err != nil {
		return fmt.Errorf("failed to bulk store associations: %w", err)
	}
	
	slog.Debug("Bulk stored associations", "count", len(associations))
	return nil
}

// GetByMemoryID retrieves all associations for a specific memory
func (qac *qdrantAssociationCollection) GetByMemoryID(ctx context.Context, memoryID string) ([]*models.MemoryAssociation, error) {
	// Search for associations where source_id OR target_id matches
	filter := &qdrant.Filter{
		Should: []*qdrant.Condition{
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "source_id",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Keyword{
								Keyword: memoryID,
							},
						},
					},
				},
			},
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "target_id",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Keyword{
								Keyword: memoryID,
							},
						},
					},
				},
			},
		},
	}
	
	limit := uint32(1000) // Reasonable limit for associations
	response, err := qac.client.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: qac.collectionName,
		Filter:         filter,
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get associations for memory %s: %w", memoryID, err)
	}
	
	associations := make([]*models.MemoryAssociation, 0, len(response))
	for _, point := range response {
		association, err := qdrantPointToAssociation(point)
		if err != nil {
			slog.Warn("Failed to convert point to association", "error", err)
			continue
		}
		associations = append(associations, association)
	}
	
	return associations, nil
}

// GetByMemoryIDs retrieves associations for multiple memories
func (qac *qdrantAssociationCollection) GetByMemoryIDs(ctx context.Context, memoryIDs []string) (map[string][]*models.MemoryAssociation, error) {
	result := make(map[string][]*models.MemoryAssociation)
	
	// For simplicity, query each memory ID separately
	// This could be optimized with a single complex query if needed
	for _, memoryID := range memoryIDs {
		associations, err := qac.GetByMemoryID(ctx, memoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get associations for memory %s: %w", memoryID, err)
		}
		result[memoryID] = associations
	}
	
	return result, nil
}

// Delete removes specific associations by their IDs
func (qac *qdrantAssociationCollection) Delete(ctx context.Context, associationIDs []string) error {
	if len(associationIDs) == 0 {
		return nil
	}
	
	pointIds := make([]*qdrant.PointId, len(associationIDs))
	for i, id := range associationIDs {
		pointIds[i] = &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}}
	}
	
	_, err := qac.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: qac.collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIds,
				},
			},
		},
	})
	
	if err != nil {
		return fmt.Errorf("failed to delete associations: %w", err)
	}
	
	return nil
}

// DeleteByMemoryID removes all associations for a specific memory
func (qac *qdrantAssociationCollection) DeleteByMemoryID(ctx context.Context, memoryID string) error {
	// First get all association IDs for this memory
	associations, err := qac.GetByMemoryID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("failed to get associations for deletion: %w", err)
	}
	
	if len(associations) == 0 {
		return nil // Nothing to delete
	}
	
	// Extract IDs and delete
	ids := make([]string, len(associations))
	for i, assoc := range associations {
		ids[i] = assoc.ID
	}
	
	return qac.Delete(ctx, ids)
}

// Count returns the total number of associations
func (qac *qdrantAssociationCollection) Count(ctx context.Context) (uint64, error) {
	response, err := qac.client.Count(ctx, &qdrant.CountPoints{
		CollectionName: qac.collectionName,
		Exact:          &[]bool{true}[0],
	})
	
	if err != nil {
		return 0, fmt.Errorf("failed to count associations: %w", err)
	}
	
	return response, nil
}

// GetAll retrieves all associations with pagination
func (qac *qdrantAssociationCollection) GetAll(ctx context.Context, cursor string, limit uint32) (associations []*models.MemoryAssociation, nextCursor string, err error) {
	scrollRequest := &qdrant.ScrollPoints{
		CollectionName: qac.collectionName,
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	}
	
	if cursor != "" {
		scrollRequest.Offset = &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: cursor}}
	}
	
	response, err := qac.client.Scroll(ctx, scrollRequest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to scroll associations: %w", err)
	}
	
	associations = make([]*models.MemoryAssociation, 0, len(response))
	for _, point := range response {
		association, err := qdrantPointToAssociation(point)
		if err != nil {
			slog.Warn("Failed to convert point to association", "error", err)
			continue
		}
		associations = append(associations, association)
	}
	
	// Set next cursor if there might be more results
	if len(associations) == int(limit) && len(associations) > 0 {
		nextCursor = associations[len(associations)-1].ID
	}
	
	return associations, nextCursor, nil
}

// Helper functions

// associationToQdrantPayload converts a MemoryAssociation to Qdrant payload
func associationToQdrantPayload(association *models.MemoryAssociation) map[string]*qdrant.Value {
	payload := map[string]*qdrant.Value{
		"source_id":  {Kind: &qdrant.Value_StringValue{StringValue: association.SourceID}},
		"target_id":  {Kind: &qdrant.Value_StringValue{StringValue: association.TargetID}},
		"type":       {Kind: &qdrant.Value_StringValue{StringValue: string(association.Type)}},
		"strength":   {Kind: &qdrant.Value_DoubleValue{DoubleValue: association.Strength}},
		"created_at": {Kind: &qdrant.Value_IntegerValue{IntegerValue: association.CreatedAt.Unix()}},
		"updated_at": {Kind: &qdrant.Value_IntegerValue{IntegerValue: association.UpdatedAt.Unix()}},
	}
	
	// Add metadata
	for key, value := range association.Metadata {
		payload[key] = anyToQdrantValue(value)
	}
	
	return payload
}

// qdrantPointToAssociation converts a Qdrant point to MemoryAssociation
func qdrantPointToAssociation(point *qdrant.RetrievedPoint) (*models.MemoryAssociation, error) {
	association := &models.MemoryAssociation{
		ID:       point.Id.GetUuid(),
		Metadata: make(map[string]any),
	}
	
	if payload := point.Payload; payload != nil {
		if sourceID := payload["source_id"]; sourceID != nil {
			association.SourceID = sourceID.GetStringValue()
		}
		if targetID := payload["target_id"]; targetID != nil {
			association.TargetID = targetID.GetStringValue()
		}
		if assocType := payload["type"]; assocType != nil {
			association.Type = models.AssociationType(assocType.GetStringValue())
		}
		if strength := payload["strength"]; strength != nil {
			association.Strength = strength.GetDoubleValue()
		}
		if createdAt := payload["created_at"]; createdAt != nil {
			if timestamp := createdAt.GetIntegerValue(); timestamp != 0 {
				association.CreatedAt = timeFromUnix(timestamp)
			}
		}
		if updatedAt := payload["updated_at"]; updatedAt != nil {
			if timestamp := updatedAt.GetIntegerValue(); timestamp != 0 {
				association.UpdatedAt = timeFromUnix(timestamp)
			}
		}
		
		// Extract metadata
		for key, value := range payload {
			if key == "source_id" || key == "target_id" || key == "type" || key == "strength" || key == "created_at" || key == "updated_at" {
				continue
			}
			switch v := value.Kind.(type) {
			case *qdrant.Value_StringValue:
				association.Metadata[key] = v.StringValue
			case *qdrant.Value_IntegerValue:
				association.Metadata[key] = v.IntegerValue
			case *qdrant.Value_DoubleValue:
				association.Metadata[key] = v.DoubleValue
			case *qdrant.Value_BoolValue:
				association.Metadata[key] = v.BoolValue
			}
		}
	}
	
	return association, nil
}