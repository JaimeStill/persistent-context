package journal

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/JaimeStill/persistent-context/pkg/models"
)

// AssociationTracker manages relationships between memories
type AssociationTracker struct {
	// In-memory association storage (could be moved to persistent storage later)
	associations map[string]*models.MemoryAssociation
	// Index for quick lookups by source memory
	sourceIndex map[string][]*models.MemoryAssociation
	// Index for quick lookups by target memory
	targetIndex map[string][]*models.MemoryAssociation
}

// NewAssociationTracker creates a new association tracker
func NewAssociationTracker() *AssociationTracker {
	return &AssociationTracker{
		associations: make(map[string]*models.MemoryAssociation),
		sourceIndex:  make(map[string][]*models.MemoryAssociation),
		targetIndex:  make(map[string][]*models.MemoryAssociation),
	}
}

// CreateAssociation creates a new association between two memories
func (at *AssociationTracker) CreateAssociation(sourceID, targetID string, associationType models.AssociationType, strength float64, metadata map[string]any) *models.MemoryAssociation {
	association := &models.MemoryAssociation{
		ID:        uuid.New().String(),
		SourceID:  sourceID,
		TargetID:  targetID,
		Type:      associationType,
		Strength:  strength,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  metadata,
	}
	
	// Store association
	at.associations[association.ID] = association
	
	// Update indexes
	at.sourceIndex[sourceID] = append(at.sourceIndex[sourceID], association)
	at.targetIndex[targetID] = append(at.targetIndex[targetID], association)
	
	return association
}

// GetAssociationsForMemory returns all associations for a given memory ID
func (at *AssociationTracker) GetAssociationsForMemory(memoryID string) []*models.MemoryAssociation {
	var associations []*models.MemoryAssociation
	
	// Get associations where this memory is the source
	if sourceAssocs, exists := at.sourceIndex[memoryID]; exists {
		associations = append(associations, sourceAssocs...)
	}
	
	// Get associations where this memory is the target
	if targetAssocs, exists := at.targetIndex[memoryID]; exists {
		associations = append(associations, targetAssocs...)
	}
	
	return associations
}

// GetRelatedMemoryIDs returns IDs of memories related to the given memory
func (at *AssociationTracker) GetRelatedMemoryIDs(memoryID string) []string {
	associations := at.GetAssociationsForMemory(memoryID)
	relatedIDs := make([]string, 0, len(associations))
	
	for _, assoc := range associations {
		if assoc.SourceID == memoryID {
			relatedIDs = append(relatedIDs, assoc.TargetID)
		} else {
			relatedIDs = append(relatedIDs, assoc.SourceID)
		}
	}
	
	return relatedIDs
}

// UpdateAssociationStrength updates the strength of an existing association
func (at *AssociationTracker) UpdateAssociationStrength(associationID string, newStrength float64) bool {
	if association, exists := at.associations[associationID]; exists {
		association.Strength = newStrength
		association.UpdatedAt = time.Now()
		return true
	}
	return false
}

// RemoveAssociation removes an association and updates indexes
func (at *AssociationTracker) RemoveAssociation(associationID string) bool {
	association, exists := at.associations[associationID]
	if !exists {
		return false
	}
	
	// Remove from main storage
	delete(at.associations, associationID)
	
	// Remove from source index
	if sourceAssocs, exists := at.sourceIndex[association.SourceID]; exists {
		at.sourceIndex[association.SourceID] = removeAssociationFromSlice(sourceAssocs, associationID)
	}
	
	// Remove from target index
	if targetAssocs, exists := at.targetIndex[association.TargetID]; exists {
		at.targetIndex[association.TargetID] = removeAssociationFromSlice(targetAssocs, associationID)
	}
	
	return true
}

// removeAssociationFromSlice removes an association from a slice by ID
func removeAssociationFromSlice(associations []*models.MemoryAssociation, associationID string) []*models.MemoryAssociation {
	for i, assoc := range associations {
		if assoc.ID == associationID {
			return append(associations[:i], associations[i+1:]...)
		}
	}
	return associations
}

// AssociationAnalyzer provides algorithms for creating associations automatically
type AssociationAnalyzer struct {
	tracker *AssociationTracker
}

// NewAssociationAnalyzer creates a new association analyzer
func NewAssociationAnalyzer(tracker *AssociationTracker) *AssociationAnalyzer {
	return &AssociationAnalyzer{
		tracker: tracker,
	}
}

// AnalyzeTemporalAssociations finds memories that occurred close in time
func (aa *AssociationAnalyzer) AnalyzeTemporalAssociations(ctx context.Context, memory *models.MemoryEntry, recentMemories []*models.MemoryEntry, timeWindow time.Duration) {
	for _, otherMemory := range recentMemories {
		if otherMemory.ID == memory.ID {
			continue // Skip self
		}
		
		// Calculate time difference
		timeDiff := math.Abs(float64(memory.CreatedAt.Sub(otherMemory.CreatedAt)))
		
		// If within time window, create temporal association
		if time.Duration(timeDiff) <= timeWindow {
			strength := aa.calculateTemporalStrength(time.Duration(timeDiff), timeWindow)
			metadata := map[string]any{
				"time_diff_minutes": timeDiff / float64(time.Minute),
				"created_at":        time.Now().Unix(),
			}
			
			aa.tracker.CreateAssociation(
				memory.ID,
				otherMemory.ID,
				models.AssociationTemporal,
				strength,
				metadata,
			)
		}
	}
}

// AnalyzeSemanticAssociations finds memories with similar content using embeddings
func (aa *AssociationAnalyzer) AnalyzeSemanticAssociations(ctx context.Context, memory *models.MemoryEntry, candidateMemories []*models.MemoryEntry, similarityThreshold float64) {
	if memory.Embedding == nil || len(memory.Embedding) == 0 {
		return // Cannot analyze without embeddings
	}
	
	for _, otherMemory := range candidateMemories {
		if otherMemory.ID == memory.ID {
			continue // Skip self
		}
		
		if otherMemory.Embedding == nil || len(otherMemory.Embedding) == 0 {
			continue // Skip memories without embeddings
		}
		
		// Calculate cosine similarity between embeddings
		similarity := aa.calculateCosineSimilarity(memory.Embedding, otherMemory.Embedding)
		
		// If similarity is above threshold, create semantic association
		if similarity >= similarityThreshold {
			metadata := map[string]any{
				"similarity_score": similarity,
				"created_at":       time.Now().Unix(),
			}
			
			aa.tracker.CreateAssociation(
				memory.ID,
				otherMemory.ID,
				models.AssociationSemantic,
				similarity,
				metadata,
			)
		}
	}
}

// AnalyzeContextualAssociations finds memories from similar contexts (same source, session, etc.)
func (aa *AssociationAnalyzer) AnalyzeContextualAssociations(ctx context.Context, memory *models.MemoryEntry, contextMemories []*models.MemoryEntry) {
	memorySource := ""
	if memory.Metadata != nil {
		if source, ok := memory.Metadata["source"].(string); ok {
			memorySource = source
		}
	}
	
	if memorySource == "" {
		return // Cannot analyze without source context
	}
	
	for _, otherMemory := range contextMemories {
		if otherMemory.ID == memory.ID {
			continue // Skip self
		}
		
		otherSource := ""
		if otherMemory.Metadata != nil {
			if source, ok := otherMemory.Metadata["source"].(string); ok {
				otherSource = source
			}
		}
		
		// If from same source/context, create contextual association
		if otherSource != "" && otherSource == memorySource {
			strength := 0.7 // Moderate strength for contextual associations
			metadata := map[string]any{
				"shared_context": memorySource,
				"created_at":     time.Now().Unix(),
			}
			
			aa.tracker.CreateAssociation(
				memory.ID,
				otherMemory.ID,
				models.AssociationContextual,
				strength,
				metadata,
			)
		}
	}
}

// calculateTemporalStrength calculates association strength based on time proximity
func (aa *AssociationAnalyzer) calculateTemporalStrength(timeDiff, maxWindow time.Duration) float64 {
	// Stronger association for closer times (inverse relationship)
	ratio := float64(timeDiff) / float64(maxWindow)
	return math.Max(0.1, 1.0-ratio) // Minimum strength of 0.1
}

// calculateCosineSimilarity calculates cosine similarity between two embedding vectors
func (aa *AssociationAnalyzer) calculateCosineSimilarity(embedding1, embedding2 []float32) float64 {
	if len(embedding1) != len(embedding2) {
		return 0.0 // Cannot compare vectors of different dimensions
	}
	
	var dotProduct, norm1, norm2 float64
	
	for i := 0; i < len(embedding1); i++ {
		dotProduct += float64(embedding1[i]) * float64(embedding2[i])
		norm1 += float64(embedding1[i]) * float64(embedding1[i])
		norm2 += float64(embedding2[i]) * float64(embedding2[i])
	}
	
	// Avoid division by zero
	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}