package journal

import (
	"math"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// MemoryScorer provides enhanced memory scoring algorithms
type MemoryScorer struct {
	config *config.ConsolidationConfig
}

// NewMemoryScorer creates a new memory scorer with configurable parameters
func NewMemoryScorer(config *config.ConsolidationConfig) *MemoryScorer {
	return &MemoryScorer{
		config: config,
	}
}

// ScoreMemory calculates enhanced importance score for a memory by considering
// multiple factors: base importance, time decay, access frequency, and relevance
func (ms *MemoryScorer) ScoreMemory(memory *types.MemoryEntry) types.MemoryScore {
	now := time.Now()
	
	// Extract or initialize access frequency from existing data
	accessFreq := 1
	if memory.Score.AccessFrequency > 0 {
		accessFreq = memory.Score.AccessFrequency
	} else if freq, ok := memory.Metadata["access_count"].(int); ok {
		accessFreq = freq
	}
	
	// Calculate time-based decay factor (how fresh is this memory?)
	timeSinceAccess := now.Sub(memory.AccessedAt)
	decayFactor := ms.calculateDecayFactor(timeSinceAccess)
	
	// Calculate base importance from memory strength and type
	baseImportance := ms.calculateBaseImportance(memory)
	
	// Calculate relevance score (using existing strength as foundation)
	relevanceScore := float64(memory.Strength)
	
	// Calculate final composite score combining all factors
	compositeScore := ms.calculateCompositeScore(baseImportance, decayFactor, accessFreq, relevanceScore)
	
	return types.MemoryScore{
		BaseImportance:  baseImportance,
		DecayFactor:     decayFactor,
		AccessFrequency: accessFreq,
		LastAccessed:    memory.AccessedAt,
		RelevanceScore:  relevanceScore,
		CompositeScore:  compositeScore,
	}
}

// calculateDecayFactor computes time-based decay using exponential decay formula
// Recent memories maintain high importance, older memories fade naturally
func (ms *MemoryScorer) calculateDecayFactor(timeSinceAccess time.Duration) float64 {
	hours := timeSinceAccess.Hours()
	
	// Exponential decay: e^(-lambda * t) where lambda controls decay rate
	lambda := ms.config.DecayFactor
	decayFactor := math.Exp(-lambda * hours)
	
	// Ensure minimum decay factor to prevent complete obsolescence
	if decayFactor < 0.01 {
		decayFactor = 0.01
	}
	
	return decayFactor
}

// calculateBaseImportance determines inherent value based on memory type and content
// Different memory types have different baseline importance levels
func (ms *MemoryScorer) calculateBaseImportance(memory *types.MemoryEntry) float64 {
	baseImportance := float64(memory.Strength)
	
	// Apply type-based importance multipliers
	switch memory.Type {
	case types.TypeSemantic:
		// Facts and knowledge are highly valuable (consolidated understanding)
		baseImportance *= 1.5
	case types.TypeProcedural:
		// Skills and procedures are quite valuable (actionable knowledge)
		baseImportance *= 1.3
	case types.TypeMetacognitive:
		// Learning insights are very valuable (meta-knowledge)
		baseImportance *= 1.4
	case types.TypeEpisodic:
		// Raw experiences have baseline importance
		baseImportance *= 1.0
	}
	
	// Factor in content length (more comprehensive content gets slight bonus)
	contentLength := float64(len(memory.Content))
	lengthFactor := math.Log(1 + contentLength/1000) // Logarithmic scaling
	baseImportance *= (1.0 + lengthFactor*0.1)
	
	// Ensure importance stays within bounds
	if baseImportance > 1.0 {
		baseImportance = 1.0
	}
	
	return baseImportance
}

// calculateCompositeScore combines all scoring factors into final importance value
// Uses configurable weights to balance different aspects of memory importance
func (ms *MemoryScorer) calculateCompositeScore(baseImportance, decayFactor float64, accessFreq int, relevanceScore float64) float64 {
	// Normalize access frequency using logarithmic scaling to prevent outliers
	normalizedAccessFreq := math.Log(1 + float64(accessFreq))
	
	// Weight components according to configuration
	accessComponent := normalizedAccessFreq * ms.config.AccessWeight
	relevanceComponent := relevanceScore * ms.config.RelevanceWeight
	
	// Combine weighted components
	baseScore := baseImportance * (accessComponent + relevanceComponent)
	
	// Apply decay factor last (even important memories fade without access)
	compositeScore := baseScore * decayFactor
	
	return compositeScore
}

// UpdateMemoryAccess updates access tracking and recalculates score when memory is used
// Creates positive feedback loop where useful memories become more accessible
func (ms *MemoryScorer) UpdateMemoryAccess(memory *types.MemoryEntry) {
	now := time.Now()
	
	// Increment access frequency counter
	if memory.Score.AccessFrequency == 0 {
		// Initialize from metadata if migrating from old system
		if freq, ok := memory.Metadata["access_count"].(int); ok {
			memory.Score.AccessFrequency = freq + 1
		} else {
			memory.Score.AccessFrequency = 1
		}
	} else {
		memory.Score.AccessFrequency++
	}
	
	// Update access timestamps
	memory.AccessedAt = now
	memory.Score.LastAccessed = now
	
	// Maintain backward compatibility with metadata-based tracking
	if memory.Metadata == nil {
		memory.Metadata = make(map[string]any)
	}
	memory.Metadata["access_count"] = memory.Score.AccessFrequency
	memory.Metadata["last_access"] = now.Unix()
	
	// Recalculate score with updated access information
	memory.Score = ms.ScoreMemory(memory)
}

// ScoreMemories efficiently scores a batch of memories
func (ms *MemoryScorer) ScoreMemories(memories []*types.MemoryEntry) {
	for _, memory := range memories {
		memory.Score = ms.ScoreMemory(memory)
	}
}

// GetTopScoredMemories returns memories sorted by importance score up to specified limit
// Used by consolidation engine to select most valuable memories for processing
func (ms *MemoryScorer) GetTopScoredMemories(memories []*types.MemoryEntry, limit int) []*types.MemoryEntry {
	// Score all memories first
	ms.ScoreMemories(memories)
	
	// Create copy for sorting to avoid modifying original slice
	sortedMemories := make([]*types.MemoryEntry, len(memories))
	copy(sortedMemories, memories)
	
	// Sort by composite score in descending order (highest first)
	// Using bubble sort for simplicity - could optimize for large datasets
	for i := 0; i < len(sortedMemories)-1; i++ {
		for j := 0; j < len(sortedMemories)-i-1; j++ {
			if sortedMemories[j].Score.CompositeScore < sortedMemories[j+1].Score.CompositeScore {
				sortedMemories[j], sortedMemories[j+1] = sortedMemories[j+1], sortedMemories[j]
			}
		}
	}
	
	// Return top memories up to specified limit
	if limit > len(sortedMemories) {
		limit = len(sortedMemories)
	}
	
	return sortedMemories[:limit]
}