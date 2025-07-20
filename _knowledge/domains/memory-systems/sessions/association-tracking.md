---
domain: memory-systems
name: association-tracking
title: Association Tracking System
duration: 30
status: pending
prerequisites: [memory-systems/processing-pipeline, memory-systems/consolidation]
builds_on: [memory-processing-pipeline, memory-consolidation]
unlocks: [memory-relationships, graph-concepts, association-scoring]
complexity: advanced
---

# Association Tracking System

## Concept Overview

Association tracking is what transforms your memory system from a simple storage mechanism into an **intelligent knowledge graph**. Instead of isolated memories, you get a web of interconnected knowledge where each memory can lead to related memories, creating a more natural and powerful retrieval system.

**Core Problems It Solves:**

- Finding related memories without exact keyword matches
- Understanding context and relationships between different conversations
- Enabling serendipitous discovery of relevant knowledge
- Supporting more sophisticated consolidation by understanding memory relationships

**Why This Matters for Session 14:**
Association tracking is crucial for advanced consolidation features. The system uses association strength to determine which memories should be consolidated together and how to weight memory importance.

## Visualization: Knowledge Graph Analogy

Think of association tracking like **building a map of knowledge**:

**Traditional Search (No Associations):**

```
Query: "Go channels" → Find memories containing "Go channels"
Results: Only exact matches, no context
```

**Association-Enhanced Search:**

```
Query: "Go channels" → Find memories containing "Go channels"
├─ Associated: "goroutine patterns" (semantic similarity)
├─ Associated: "race condition debugging" (temporal proximity)  
├─ Associated: "worker pool implementation" (causal relationship)
└─ Associated: "concurrency best practices" (contextual relevance)
```

**Real-World Analogy**: Like how your brain works - thinking about "Go channels" might trigger memories of debugging concurrency issues, even if those memories don't mention channels directly.

## Association Types

**Location**: `src/pkg/models/models.go`

Your system tracks four types of associations:

```go
type AssociationType string

const (
    // Memories created close together in time
    AssociationTypeTemporal   AssociationType = "temporal"
    
    // Memories with similar semantic content
    AssociationTypeSemantic   AssociationType = "semantic"
    
    // Memories with cause-effect relationships
    AssociationTypeCausal     AssociationType = "causal"
    
    // Memories from similar contexts/sessions
    AssociationTypeContextual AssociationType = "contextual"
)

type Association struct {
    ID         string          `json:"id"`
    SourceID   string          `json:"source_id"`    // Memory A
    TargetID   string          `json:"target_id"`    // Memory B  
    Type       AssociationType `json:"type"`         // How they're related
    Strength   float64         `json:"strength"`     // How strong the relationship is (0.0-1.0)
    CreatedAt  int64          `json:"created_at"`   // When association was discovered
    Metadata   map[string]interface{} `json:"metadata"` // Additional context
}
```

**Relationship Modeling**: This represents a many-to-many relationship between memories:

- **Source/Target IDs**: The two memories being connected
- **Association Type**: How they're related (temporal, semantic, causal, contextual)
- **Strength**: Quantifies the relationship intensity (0.0-1.0)
- **Metadata**: Additional context about the relationship

## Association Creation Process

**Location**: `src/pkg/journal/associations.go`

### Temporal Associations

Created when memories are captured close together in time:

```go
func (j *journal) createTemporalAssociations(ctx context.Context, newMemory *models.Memory) error {
    // Find memories created within the last hour
    timeWindow := 1 * time.Hour
    recentMemories, err := j.getMemoriesInTimeWindow(ctx, newMemory.Timestamp, timeWindow)
    if err != nil {
        return err
    }
    
    for _, recentMemory := range recentMemories {
        if recentMemory.ID == newMemory.ID {
            continue // Don't associate with self
        }
        
        // Calculate temporal strength (closer in time = stronger)
        timeDiff := math.Abs(float64(newMemory.Timestamp - recentMemory.Timestamp))
        maxDiff := timeWindow.Seconds()
        strength := 1.0 - (timeDiff / maxDiff) // 1.0 for same time, 0.0 for edge of window
        
        association := &models.Association{
            ID:        uuid.New().String(),
            SourceID:  newMemory.ID,
            TargetID:  recentMemory.ID,
            Type:      models.AssociationTypeTemporal,
            Strength:  strength,
            CreatedAt: time.Now().Unix(),
        }
        
        err = j.storeAssociation(ctx, association)
        if err != nil {
            j.logger.Error("Failed to store temporal association", "error", err)
        }
    }
    
    return nil
}
```

**Key Insight**: Memories created close together in time are likely related to the same conversation or topic.

### Semantic Associations

Created based on content similarity using vector embeddings:

```go
func (j *journal) createSemanticAssociations(ctx context.Context, newMemory *models.Memory) error {
    if len(newMemory.Embedding) == 0 {
        return nil // Can't create semantic associations without embedding
    }
    
    // Find similar memories using vector similarity
    similarMemories, err := j.vectorDB.Query(ctx, newMemory.Embedding, 10, 0.7) // Top 10, min 70% similarity
    if err != nil {
        return err
    }
    
    for _, similar := range similarMemories {
        if similar.ID == newMemory.ID {
            continue
        }
        
        association := &models.Association{
            ID:        uuid.New().String(),
            SourceID:  newMemory.ID,
            TargetID:  similar.ID,
            Type:      models.AssociationTypeSemantic,
            Strength:  similar.Score, // Use similarity score as strength
            CreatedAt: time.Now().Unix(),
            Metadata: map[string]interface{}{
                "similarity_score": similar.Score,
            },
        }
        
        err = j.storeAssociation(ctx, association)
        if err != nil {
            j.logger.Error("Failed to store semantic association", "error", err)
        }
    }
    
    return nil
}
```

**Key Insight**: Vector similarity automatically finds memories about related topics, even with different wording.

### Causal Associations

Created when memories have cause-effect relationships:

```go
func (j *journal) createCausalAssociations(ctx context.Context, newMemory *models.Memory) error {
    // Look for causal patterns in content
    causalPatterns := []string{
        "because of", "due to", "caused by", "resulted in", 
        "led to", "fixed by", "solved by", "after implementing",
    }
    
    content := strings.ToLower(newMemory.Content)
    hasCausalLanguage := false
    for _, pattern := range causalPatterns {
        if strings.Contains(content, pattern) {
            hasCausalLanguage = true
            break
        }
    }
    
    if !hasCausalLanguage {
        return nil // No causal indicators
    }
    
    // Find recent memories that might be causes
    recentMemories, err := j.getMemoriesInTimeWindow(ctx, newMemory.Timestamp, 2*time.Hour)
    if err != nil {
        return err
    }
    
    for _, recentMemory := range recentMemories {
        // Use simple heuristics to detect causal relationships
        strength := j.calculateCausalStrength(newMemory, recentMemory)
        if strength > 0.5 { // Only create strong causal associations
            association := &models.Association{
                ID:        uuid.New().String(),
                SourceID:  recentMemory.ID, // Cause
                TargetID:  newMemory.ID,    // Effect
                Type:      models.AssociationTypeCausal,
                Strength:  strength,
                CreatedAt: time.Now().Unix(),
            }
            
            err = j.storeAssociation(ctx, association)
            if err != nil {
                j.logger.Error("Failed to store causal association", "error", err)
            }
        }
    }
    
    return nil
}

func (j *journal) calculateCausalStrength(effect, cause *models.Memory) float64 {
    // Simple heuristic: look for shared keywords between cause and effect
    effectWords := strings.Fields(strings.ToLower(effect.Content))
    causeWords := strings.Fields(strings.ToLower(cause.Content))
    
    sharedWords := 0
    effectWordSet := make(map[string]bool)
    for _, word := range effectWords {
        effectWordSet[word] = true
    }
    
    for _, word := range causeWords {
        if effectWordSet[word] && len(word) > 3 { // Ignore short common words
            sharedWords++
        }
    }
    
    // Normalize by average number of words
    avgWords := float64(len(effectWords)+len(causeWords)) / 2.0
    return float64(sharedWords) / avgWords
}
```

**Key Insight**: Causal associations help understand problem-solution pairs and learning progressions.

### Contextual Associations

Created when memories share similar contexts (same session, user, topic):

```go
func (j *journal) createContextualAssociations(ctx context.Context, newMemory *models.Memory) error {
    // Extract session ID from metadata
    sessionID, exists := newMemory.Metadata["session_id"]
    if !exists {
        return nil // No session context
    }
    
    // Find other memories from the same session
    sessionMemories, err := j.getMemoriesByMetadata(ctx, "session_id", sessionID)
    if err != nil {
        return err
    }
    
    for _, sessionMemory := range sessionMemories {
        if sessionMemory.ID == newMemory.ID {
            continue
        }
        
        // Contextual strength based on session position
        strength := j.calculateContextualStrength(newMemory, sessionMemory)
        
        association := &models.Association{
            ID:        uuid.New().String(),
            SourceID:  newMemory.ID,
            TargetID:  sessionMemory.ID,
            Type:      models.AssociationTypeContextual,
            Strength:  strength,
            CreatedAt: time.Now().Unix(),
            Metadata: map[string]interface{}{
                "session_id": sessionID,
            },
        }
        
        err = j.storeAssociation(ctx, association)
        if err != nil {
            j.logger.Error("Failed to store contextual association", "error", err)
        }
    }
    
    return nil
}
```

## Association Storage and Retrieval

**Location**: `src/pkg/journal/associations.go`

### Bidirectional Indexing

Associations are stored bidirectionally for efficient lookups:

```go
func (j *journal) storeAssociation(ctx context.Context, association *models.Association) error {
    // Store the association itself
    err := j.vectorDB.StoreAssociation(ctx, association)
    if err != nil {
        return err
    }
    
    // Create reverse association for bidirectional lookup
    reverseAssociation := &models.Association{
        ID:        uuid.New().String(),
        SourceID:  association.TargetID,  // Swap source and target
        TargetID:  association.SourceID,
        Type:      association.Type,
        Strength:  association.Strength,
        CreatedAt: association.CreatedAt,
        Metadata:  association.Metadata,
    }
    
    return j.vectorDB.StoreAssociation(ctx, reverseAssociation)
}

func (j *journal) GetAssociations(ctx context.Context, memoryID string) ([]*models.Association, error) {
    // Get all associations where this memory is the source
    associations, err := j.vectorDB.GetAssociationsBySource(ctx, memoryID)
    if err != nil {
        return nil, err
    }
    
    // Filter by minimum strength threshold
    filtered := make([]*models.Association, 0)
    for _, assoc := range associations {
        if assoc.Strength >= j.config.MinAssociationStrength { // e.g., 0.3
            filtered = append(filtered, assoc)
        }
    }
    
    return filtered, nil
}
```

**Bidirectional Indexing**: This ensures fast lookups in both directions:

- **Forward Lookup**: Find all memories that this memory points to
- **Reverse Lookup**: Find all memories that point to this memory
- **Performance**: O(1) lookups instead of scanning all associations
- **Flexibility**: Supports graph traversal in any direction

## Integration with Memory Processing

**Location**: `src/pkg/memory/processor.go`

Associations are created as part of the memory processing pipeline:

```go
func (p *processor) processMemoryWithAssociations(ctx context.Context, memory *models.Memory) error {
    // 1. Store the memory first
    err := p.journal.Store(ctx, memory)
    if err != nil {
        return err
    }
    
    // 2. Create associations (run in parallel)
    var wg sync.WaitGroup
    wg.Add(4)
    
    // Temporal associations
    go func() {
        defer wg.Done()
        if err := p.journal.CreateTemporalAssociations(ctx, memory); err != nil {
            p.logger.Error("Failed to create temporal associations", "error", err)
        }
    }()
    
    // Semantic associations (requires embedding)
    go func() {
        defer wg.Done()
        if len(memory.Embedding) > 0 {
            if err := p.journal.CreateSemanticAssociations(ctx, memory); err != nil {
                p.logger.Error("Failed to create semantic associations", "error", err)
            }
        }
    }()
    
    // Causal associations
    go func() {
        defer wg.Done()
        if err := p.journal.CreateCausalAssociations(ctx, memory); err != nil {
            p.logger.Error("Failed to create causal associations", "error", err)
        }
    }()
    
    // Contextual associations
    go func() {
        defer wg.Done()
        if err := p.journal.CreateContextualAssociations(ctx, memory); err != nil {
            p.logger.Error("Failed to create contextual associations", "error", err)
        }
    }()
    
    wg.Wait() // Wait for all association creation to complete
    return nil
}
```

**Key Pattern**: Association creation runs in parallel using goroutines, speeding up the process.

## Practical Exercise: Explore Associations

Let's create some memories and observe the associations:

### Setup

Ensure your system is running with some existing memories.

### Exercise 1: Create Related Memories

**Step 1**: Create a sequence of related memories:

```
1. capture_memory: "Started learning about Go channels for concurrency"
2. capture_memory: "Implemented a simple channel example with goroutines"  
3. capture_memory: "Ran into deadlock issue with unbuffered channels"
4. capture_memory: "Fixed deadlock by using buffered channels"
5. capture_memory: "Now understand the difference between buffered and unbuffered channels"
```

**Step 2**: Search for one of the memories:

```
search_memories with content: "channel deadlock"
```

**Expected Result**: Should find the deadlock memory plus associated memories about channels.

### Exercise 2: Explore Association API

**Step 1**: Find a memory ID from your search results

**Step 2**: Use curl to explore associations (if API exists):

```bash
curl "localhost:8543/api/memory/{memory-id}/associations"
```

**Expected Response**: List of associations with types and strengths.

## Association-Enhanced Consolidation

**Location**: `src/pkg/memory/processor.go:selectMemoriesForConsolidation()`

Associations influence which memories get consolidated together:

```go
func (p *processor) selectRelatedMemoriesForConsolidation(ctx context.Context, seedMemory *models.Memory) []*models.Memory {
    selected := []*models.Memory{seedMemory}
    toExplore := []*models.Memory{seedMemory}
    explored := make(map[string]bool)
    
    for len(toExplore) > 0 && len(selected) < p.config.MaxConsolidationSize {
        current := toExplore[0]
        toExplore = toExplore[1:]
        
        if explored[current.ID] {
            continue
        }
        explored[current.ID] = true
        
        // Get associations for current memory
        associations, err := p.journal.GetAssociations(ctx, current.ID)
        if err != nil {
            continue
        }
        
        for _, assoc := range associations {
            // Only include strong associations
            if assoc.Strength < 0.7 {
                continue
            }
            
            // Get the target memory
            targetMemory, err := p.journal.Retrieve(ctx, assoc.TargetID)
            if err != nil {
                continue
            }
            
            // Add to consolidation set if not already included
            if !explored[targetMemory.ID] {
                selected = append(selected, targetMemory)
                toExplore = append(toExplore, targetMemory)
            }
        }
    }
    
    return selected
}
```

**Key Insight**: This creates consolidation clusters based on association strength, leading to more coherent semantic memories.

## Configuration

**Location**: `src/pkg/config/journal.go`

Association behavior is configurable:

```go
type AssociationConfig struct {
    // Minimum strength threshold for storing associations
    MinAssociationStrength float64 `mapstructure:"min_association_strength"` // 0.3
    
    // Time windows for different association types
    TemporalWindow    time.Duration `mapstructure:"temporal_window"`     // 1 hour
    CausalWindow      time.Duration `mapstructure:"causal_window"`       // 2 hours
    ContextualWindow  time.Duration `mapstructure:"contextual_window"`   // 24 hours
    
    // Semantic association parameters
    SemanticSimilarityThreshold float64 `mapstructure:"semantic_similarity_threshold"` // 0.7
    MaxSemanticAssociations     int     `mapstructure:"max_semantic_associations"`     // 10
    
    // Performance limits
    MaxAssociationsPerMemory int `mapstructure:"max_associations_per_memory"` // 50
}
```

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Association Types**: Explain the difference between semantic and contextual associations. When would each be most useful?

2. **Bidirectional Storage**: Why does the system store both A→B and B→A associations instead of just A→B?

3. **Performance Trade-offs**: What are the computational costs of association creation, and how does the system manage them?

4. **Consolidation Impact**: How do associations change which memories get consolidated together compared to a system without associations?

## Connection to Session 14

Association tracking directly supports your Session 14 work:

- **Enhanced Consolidation**: Associations determine which memories consolidate together
- **Memory Scoring**: Association strength influences memory importance scores
- **Retrieval Quality**: Related memories provide richer context for responses
- **Graph Algorithms**: Foundation for future graph-based memory navigation

Understanding associations means you can enhance consolidation algorithms and debug relationship-based issues.

## Notes

<!-- Add your observations as you work through this:
- Which association types seem most useful for your use cases?
- How effective are the Go concurrent patterns (goroutines, WaitGroup) for parallel association creation?
- What questions came up about association strength calculation?
- How might you enhance the causal association detection?
-->