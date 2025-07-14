# Source Documentation 001: Enhanced Memory Scoring System

## Overview

The Enhanced Memory Scoring System is like a **smart librarian** for AI memories. Instead of treating all memories equally, it creates an intelligent ranking system that helps the AI decide which memories are most valuable to keep, access, and prioritize during decision-making.

### The Problem It Solves

Imagine if your brain kept every single detail with equal importance - every random conversation, every important lesson, every mundane task. You'd be overwhelmed! The human brain naturally prioritizes:

- Recent experiences (recency bias)
- Frequently accessed memories (familiarity)
- Emotionally significant events (importance)
- Useful skills and knowledge (utility)

Our memory scoring system mimics this natural prioritization for AI systems.

## Architecture Context

This scoring system integrates with the existing memory consolidation engine:

```
Memory Flow:
1. Raw experience captured → MemoryEntry created
2. Scoring system evaluates importance → MemoryScore calculated  
3. Consolidation engine uses scores → Decides what to keep/transform
4. Access tracking updates scores → Feedback loop improves decisions
```

## Function-by-Function Breakdown

### 1. MemoryScorer (The Judge)

```go
type MemoryScorer struct {
    config *config.ConsolidationConfig
}

func NewMemoryScorer(config *config.ConsolidationConfig) *MemoryScorer {
    return &MemoryScorer{
        config: config,
    }
}
```

**What it does**: Creates a "judge" that evaluates memory importance using configurable rules.

**Analogy**: Like hiring a professional appraiser who uses established criteria to value items. The config contains the "appraisal guidelines" (how much to weight recency vs frequency vs content type).

**Why it's designed this way**: Separating the scorer from individual scoring functions allows us to easily adjust the "judgment criteria" without rewriting the core logic.

### 2. ScoreMemory (The Main Evaluation)

```go
func (ms *MemoryScorer) ScoreMemory(memory *types.MemoryEntry) types.MemoryScore {
    now := time.Now()
    
    // Extract or initialize access frequency
    accessFreq := 1
    if memory.Score.AccessFrequency > 0 {
        accessFreq = memory.Score.AccessFrequency
    } else if freq, ok := memory.Metadata["access_count"].(int); ok {
        accessFreq = freq
    }
    
    // Calculate time-based decay factor
    timeSinceAccess := now.Sub(memory.AccessedAt)
    decayFactor := ms.calculateDecayFactor(timeSinceAccess)
    
    // Calculate base importance from memory strength and type
    baseImportance := ms.calculateBaseImportance(memory)
    
    // Calculate relevance score
    relevanceScore := float64(memory.Strength)
    
    // Calculate composite score
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
```

**What it does**: Examines one memory and assigns it an importance score by considering multiple factors.

**Analogy**: Like a restaurant critic who rates a restaurant by considering food quality, service, atmosphere, and value. Each factor contributes to the overall rating.

**The Four Factors**:

1. **Base Importance**: What type of memory is this? (skills vs random events)
2. **Decay Factor**: How fresh is this memory? (recent vs old)
3. **Access Frequency**: How often is this memory used? (popular vs forgotten)
4. **Relevance Score**: How substantial/strong is the content? (detailed vs brief)

**Key Design Decision**: The function combines multiple scoring algorithms into one comprehensive evaluation, making it easy to understand what contributes to a memory's final importance.

### 3. calculateDecayFactor (The Freshness Calculator)

```go
func (ms *MemoryScorer) calculateDecayFactor(timeSinceAccess time.Duration) float64 {
    hours := timeSinceAccess.Hours()
    
    // Exponential decay: e^(-lambda * t)
    lambda := ms.config.DecayFactor
    decayFactor := math.Exp(-lambda * hours)
    
    // Ensure decay factor is between 0 and 1
    if decayFactor < 0.01 {
        decayFactor = 0.01 // Minimum decay to prevent complete obsolescence
    }
    
    return decayFactor
}
```

**What it does**: Calculates how "fresh" a memory is based on when it was last accessed.

**Analogy**: Like how a newspaper becomes less relevant as days pass. Today's news is very important (decay factor = 1.0), yesterday's news is somewhat important (decay factor = 0.8), last week's news is barely relevant (decay factor = 0.2).

**The Math Explained**:

- Uses "exponential decay" - the same formula used for radioactive decay
- Recent memories fade slowly, old memories fade quickly
- `e^(-lambda * t)` means: as time (t) increases, importance decreases exponentially
- Lambda (λ) controls how fast things fade (higher λ = faster forgetting)

**Why This Approach**: Exponential decay mimics how human memory actually works - we don't forget linearly, we forget rapidly at first, then more slowly.

### 4. calculateBaseImportance (The Content Value Calculator)

```go
func (ms *MemoryScorer) calculateBaseImportance(memory *types.MemoryEntry) float64 {
    baseImportance := float64(memory.Strength)
    
    // Adjust importance based on memory type
    switch memory.Type {
    case types.TypeSemantic:
        baseImportance *= 1.5  // Facts and knowledge
    case types.TypeProcedural:
        baseImportance *= 1.3  // Skills and procedures
    case types.TypeMetacognitive:
        baseImportance *= 1.4  // Learning insights
    case types.TypeEpisodic:
        baseImportance *= 1.0  // Raw experiences
    }
    
    // Factor in content length
    contentLength := float64(len(memory.Content))
    lengthFactor := math.Log(1 + contentLength/1000)
    baseImportance *= (1.0 + lengthFactor*0.1)
    
    // Ensure bounds
    if baseImportance > 1.0 {
        baseImportance = 1.0
    }
    
    return baseImportance
}
```

**What it does**: Determines how inherently valuable a memory is based on its type and content.

**Analogy**: Like how different types of books have different inherent value - a cookbook you use weekly is more valuable than a novel you read once, and a technical manual is more valuable than a magazine.

**The Hierarchy**:

1. **Semantic memories (1.5x)**: General knowledge, facts - "Python is a programming language"
2. **Metacognitive memories (1.4x)**: Learning insights - "I learn better with examples"
3. **Procedural memories (1.3x)**: Skills and procedures - "How to debug code"
4. **Episodic memories (1.0x)**: Raw experiences - "Had lunch at 12pm"

**Content Length Factor**: Longer, more detailed memories get a small bonus because they likely contain more comprehensive information.

### 5. calculateCompositeScore (The Final Calculator)

```go
func (ms *MemoryScorer) calculateCompositeScore(baseImportance, decayFactor float64, accessFreq int, relevanceScore float64) float64 {
    // Normalize access frequency (logarithmic scaling)
    normalizedAccessFreq := math.Log(1 + float64(accessFreq))
    
    // Weight the components according to configuration
    accessComponent := normalizedAccessFreq * ms.config.AccessWeight
    relevanceComponent := relevanceScore * ms.config.RelevanceWeight
    
    // Combine components
    baseScore := baseImportance * (accessComponent + relevanceComponent)
    
    // Apply decay factor
    compositeScore := baseScore * decayFactor
    
    return compositeScore
}
```

**What it does**: Combines all the factors into one final importance score.

**Analogy**: Like calculating a final grade where homework is 30%, tests are 50%, and participation is 20%. Each component contributes according to its configured weight.

**The Formula Breakdown**:

1. **Normalize access frequency**: Use logarithmic scaling so 100 accesses isn't 100x more important than 1 access
2. **Apply weights**: Multiply each factor by its configured importance weight
3. **Combine**: Add weighted factors together
4. **Apply decay**: Multiply by freshness factor (recent memories score higher)

**Key Insight**: The decay factor is applied last, meaning even important memories become less valuable over time if not accessed.

### 6. UpdateMemoryAccess (The Usage Tracker)

```go
func (ms *MemoryScorer) UpdateMemoryAccess(memory *types.MemoryEntry) {
    now := time.Now()
    
    // Update access frequency
    if memory.Score.AccessFrequency == 0 {
        if freq, ok := memory.Metadata["access_count"].(int); ok {
            memory.Score.AccessFrequency = freq + 1
        } else {
            memory.Score.AccessFrequency = 1
        }
    } else {
        memory.Score.AccessFrequency++
    }
    
    // Update access time
    memory.AccessedAt = now
    memory.Score.LastAccessed = now
    
    // Update metadata for backward compatibility
    if memory.Metadata == nil {
        memory.Metadata = make(map[string]any)
    }
    memory.Metadata["access_count"] = memory.Score.AccessFrequency
    memory.Metadata["last_access"] = now.Unix()
    
    // Recalculate score with new access data
    memory.Score = ms.ScoreMemory(memory)
}
```

**What it does**: Updates tracking information every time a memory is accessed and recalculates its importance.

**Analogy**: Like YouTube tracking view counts and updating video recommendations based on popularity. More views = higher ranking in suggestions.

**What Gets Updated**:

1. **Access frequency**: +1 to the "popularity counter"
2. **Last accessed time**: Updates the "freshness timestamp"  
3. **Metadata**: Keeps backward compatibility with old tracking system
4. **Score recalculation**: Immediately updates the importance score

**Feedback Loop**: This creates a positive feedback loop where useful memories become even more accessible, while unused memories naturally fade.

### 7. GetTopScoredMemories (The Best-of Filter)

```go
func (ms *MemoryScorer) GetTopScoredMemories(memories []*types.MemoryEntry, limit int) []*types.MemoryEntry {
    // Score all memories first
    ms.ScoreMemories(memories)
    
    // Sort by composite score (descending)
    sortedMemories := make([]*types.MemoryEntry, len(memories))
    copy(sortedMemories, memories)
    
    // Simple bubble sort (could be optimized for large datasets)
    for i := 0; i < len(sortedMemories)-1; i++ {
        for j := 0; j < len(sortedMemories)-i-1; j++ {
            if sortedMemories[j].Score.CompositeScore < sortedMemories[j+1].Score.CompositeScore {
                sortedMemories[j], sortedMemories[j+1] = sortedMemories[j+1], sortedMemories[j]
            }
        }
    }
    
    // Return top memories up to limit
    if limit > len(sortedMemories) {
        limit = len(sortedMemories)
    }
    
    return sortedMemories[:limit]
}
```

**What it does**: Takes a collection of memories, scores them all, sorts by importance, and returns the top ones.

**Analogy**: Like creating a "greatest hits" playlist from your music library - score every song, sort by rating, return the top 20.

**Process**:

1. **Score everything**: Calculate importance for all memories
2. **Sort by score**: Arrange from most to least important
3. **Return top N**: Give back only the highest-scoring memories up to the limit

**Note on Algorithm**: Uses bubble sort for simplicity. For large datasets, this could be optimized with quicksort or other efficient sorting algorithms.

## Key Design Patterns Demonstrated

### 1. **Separation of Concerns**

Each function has one clear responsibility:

- Scoring logic separated from access tracking
- Different scoring factors calculated independently
- Configuration separated from implementation

### 2. **Configurable Behavior**

The system uses configuration values rather than hardcoded numbers, making it easy to tune without code changes.

### 3. **Backward Compatibility**

New scoring system works alongside existing metadata-based tracking, allowing gradual migration.

### 4. **Feedback Loops**

Access tracking automatically improves future scoring, creating a self-improving system.

## Integration with Memory Consolidation

The scoring system integrates with the existing consolidation engine:

```go
// In consolidation/engine.go
func (e *Engine) selectMemoriesForConsolidation(memories []*types.MemoryEntry) []*types.MemoryEntry {
    scorer := journal.NewMemoryScorer(e.config)
    return scorer.GetTopScoredMemories(memories, e.config.MemoryCountThreshold)
}
```

This replaces the previous simple scoring with the comprehensive system, enabling smarter consolidation decisions.

## Learning Points

1. **Exponential Decay**: Mathematical models can mimic natural processes like forgetting
2. **Logarithmic Scaling**: Prevents extreme values from dominating calculations
3. **Weighted Combinations**: Multiple factors can be balanced according to their relative importance
4. **Feedback Systems**: Usage patterns can inform future importance calculations
5. **Type-Based Logic**: Different categories of data can be handled with appropriate logic
6. **Graceful Degradation**: Systems can work with partial data and migrate gradually

## Future Enhancements

1. **Semantic Similarity**: Use embedding similarity for relevance scoring
2. **Contextual Importance**: Factor in current task relevance
3. **Learning from Outcomes**: Adjust scoring based on consolidation success
4. **Association Weighting**: Factor in memory relationship strength
5. **Adaptive Parameters**: Self-tune configuration based on usage patterns

---

*This documentation demonstrates how complex technical concepts can be made accessible through clear explanations, analogies, and step-by-step breakdowns while preserving the full technical detail.*
