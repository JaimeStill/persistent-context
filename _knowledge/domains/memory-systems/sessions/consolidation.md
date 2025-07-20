---
domain: memory-systems
name: consolidation
title: Memory Consolidation Deep Dive
duration: 45
status: pending
prerequisites: [memory-systems/processing-pipeline]
builds_on: [memory-processing-pipeline, event-driven-processing]
unlocks: [memory-scoring, consolidation-triggers, semantic-memory]
complexity: advanced
---

# Memory Consolidation Deep Dive

## Concept Overview

Memory consolidation is where the **magic of learning** happens in your system. It's the process that transforms a collection of individual experiences (episodic memories) into general knowledge and patterns (semantic memories). Think of it as the difference between **remembering individual math problems** vs **understanding mathematical principles**.

**Core Problems It Solves:**

- Prevents memory overflow by summarizing related experiences
- Extracts patterns and insights from individual events
- Creates more efficient knowledge representation
- Enables higher-level reasoning from accumulated experience

**Why This Is Critical for Session 14:**
This is likely the most complex system you'll be working on. Understanding how consolidation works - when it triggers, how it selects memories, and how it creates new knowledge - is essential for implementing the remaining backend features.

## Visualization: Human Memory Analogy

**Human Memory Consolidation:**

```
Day 1: "Learned about Go channels in tutorial"
Day 2: "Used channels to solve concurrency problem"  
Day 3: "Debugged deadlock in channel communication"
Day 4: "Taught colleague about channel patterns"

↓ Sleep/Time ↓ (Consolidation Process)

Semantic Memory: "Channels are Go's primary mechanism for goroutine communication. Common patterns include fan-out, fan-in, and worker pools. Key pitfall is deadlocks from blocking operations."
```

**Your System's Consolidation:**

```
Memory 1: "User asked about Go concurrency patterns"
Memory 2: "User implemented worker pool with channels"
Memory 3: "User debugged race condition in Go code"
Memory 4: "User asked about goroutine best practices"

↓ Consolidation Algorithm ↓

Semantic Memory: "User is learning Go concurrency, focusing on practical patterns. Has experience with channels and worker pools but needed help with debugging. Shows progression from basic questions to implementation challenges."
```

## Prerequisites Check

Before starting, ensure you understand:

- [x] Memory processing pipeline (Session 002)
- [x] Event-driven architecture basics
- [x] Go interfaces and dependency injection
- [ ] How LLM prompting works for text transformation (we'll cover this)

## Consolidation System Architecture

### Event-Driven Triggers

**Location**: `src/pkg/memory/processor.go`

Consolidation doesn't happen on a schedule - it's **event-driven** based on system state:

```go
// Consolidation trigger events
type EventType string

const (
    EventContextInit      EventType = "context_init"      // New conversation starts
    EventNewContext      EventType = "new_context"       // New memory added  
    EventThresholdReached EventType = "threshold_reached" // Too many memories
    EventConversationEnd EventType = "conversation_end"  // Session ending
)
```

**Event-Driven Design**: This represents different system states that warrant consolidation:

- **EventContextInit**: When a new conversation session begins
- **EventNewContext**: When new memories are captured (reactive consolidation)
- **EventThresholdReached**: When memory volume exceeds capacity limits
- **EventConversationEnd**: When a session concludes (proactive consolidation)

### Context Window Management

**Location**: `src/pkg/memory/processor.go:checkContextWindow()`

One of the most sophisticated parts - preventing LLM context overflow:

```go
type ContextWindowConfig struct {
    MaxTokens     uint32 `mapstructure:"max_tokens"`
    SafetyMargin  uint32 `mapstructure:"safety_margin"`
    TokensPerChar uint32 `mapstructure:"tokens_per_char"`
}

func (p *processor) checkContextWindow(memories []*models.Memory) bool {
    totalTokens := uint32(0)
    
    // Estimate tokens for each memory
    for _, memory := range memories {
        // Rough estimation: 1 token per 4 characters
        memoryTokens := uint32(len(memory.Content)) / p.config.TokensPerChar
        totalTokens += memoryTokens
    }
    
    // Add safety margin for consolidation prompt overhead
    maxAllowed := p.config.MaxTokens - p.config.SafetyMargin
    
    return totalTokens <= maxAllowed
}
```

**Why This Matters**: LLMs have token limits (e.g., 4K, 8K, 32K). If you send too much context, the request fails. This system ensures consolidation works within constraints.

## Memory Selection Algorithm

**Location**: `src/pkg/memory/processor.go:selectMemoriesForConsolidation()`

This is where the system decides **which memories to consolidate together**:

```go
func (p *processor) selectMemoriesForConsolidation(ctx context.Context) []*models.Memory {
    // 1. Get recent episodic memories (last 24 hours)
    recentMemories, err := p.journal.GetRecent(ctx, models.MemoryTypeEpisodic, 24*time.Hour)
    if err != nil {
        return nil
    }
    
    // 2. Score memories for importance
    scoredMemories := p.scoreMemories(recentMemories)
    
    // 3. Select top N memories within context window
    selected := []*models.Memory{}
    for _, memory := range scoredMemories {
        // Check if adding this memory exceeds context window
        candidate := append(selected, memory)
        if !p.checkContextWindow(candidate) {
            break // Stop adding memories
        }
        selected = candidate
    }
    
    return selected
}
```

**Key Insight**: This is a **greedy algorithm** - it picks the highest-scoring memories that fit within the context window.

## Memory Scoring System

**Location**: `src/pkg/journal/scoring.go`

This is probably the most complex algorithm in your system:

```go
type ScoringConfig struct {
    TimeDecayFactor    float64 `mapstructure:"time_decay_factor"`    // 0.9
    FrequencyWeight    float64 `mapstructure:"frequency_weight"`     // 0.3
    AssociationWeight  float64 `mapstructure:"association_weight"`   // 0.4
    RelevanceWeight    float64 `mapstructure:"relevance_weight"`     // 0.3
}

func (j *journal) scoreMemories(memories []*models.Memory) []*ScoredMemory {
    scored := make([]*ScoredMemory, 0, len(memories))
    
    for _, memory := range memories {
        score := j.calculateCompositeScore(memory)
        scored = append(scored, &ScoredMemory{
            Memory: memory,
            Score:  score,
        })
    }
    
    // Sort by score (highest first)
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })
    
    return scored
}

func (j *journal) calculateCompositeScore(memory *models.Memory) float64 {
    // 1. Time decay (newer memories score higher)
    timeScore := j.calculateTimeScore(memory.Timestamp)
    
    // 2. Access frequency (more accessed = more important)
    frequencyScore := j.calculateFrequencyScore(memory.ID)
    
    // 3. Association strength (connected memories score higher)
    associationScore := j.calculateAssociationScore(memory.ID)
    
    // 4. Relevance to recent context
    relevanceScore := j.calculateRelevanceScore(memory)
    
    // Weighted combination
    composite := (timeScore * j.config.FrequencyWeight) +
                 (frequencyScore * j.config.FrequencyWeight) +
                 (associationScore * j.config.AssociationWeight) +
                 (relevanceScore * j.config.RelevanceWeight)
    
    return composite
}
```

### Time Decay Calculation

```go
func (j *journal) calculateTimeScore(timestamp int64) float64 {
    now := time.Now().Unix()
    ageHours := float64(now-timestamp) / 3600.0
    
    // Exponential decay: score = e^(-decay_factor * age)
    timeScore := math.Exp(-j.config.TimeDecayFactor * ageHours / 24.0)
    
    return timeScore
}
```

**What This Does**: Newer memories get higher scores. A memory from 1 hour ago might score 0.95, while one from 24 hours ago might score 0.1.

### Association Scoring

```go
func (j *journal) calculateAssociationScore(memoryID string) float64 {
    // Get all associations for this memory
    associations, err := j.GetAssociations(context.Background(), memoryID)
    if err != nil {
        return 0.0
    }
    
    // Score based on number and strength of associations
    score := 0.0
    for _, assoc := range associations {
        // Stronger associations contribute more to score
        score += assoc.Strength
    }
    
    // Logarithmic scaling to prevent runaway scores
    return math.Log1p(score) / 10.0
}
```

**Key Insight**: Memories with more connections are considered more important for consolidation.

## LLM-Powered Consolidation

**Location**: `src/pkg/memory/processor.go:performConsolidation()`

This is where the actual consolidation happens using the LLM:

```go
func (p *processor) performConsolidation(ctx context.Context, memories []*models.Memory) error {
    // 1. Build consolidation prompt
    prompt := p.buildConsolidationPrompt(memories)
    
    // 2. Call LLM for consolidation
    consolidatedContent, err := p.llm.Consolidate(ctx, prompt)
    if err != nil {
        return fmt.Errorf("LLM consolidation failed: %w", err)
    }
    
    // 3. Create semantic memory
    semanticMemory := &models.Memory{
        ID:             uuid.New().String(),
        Type:           models.MemoryTypeSemantic,
        Content:        consolidatedContent,
        Timestamp:      time.Now().Unix(),
        SourceMemories: p.extractMemoryIDs(memories),
    }
    
    // 4. Store consolidated memory
    _, err = p.journal.Store(ctx, semanticMemory)
    if err != nil {
        return fmt.Errorf("failed to store semantic memory: %w", err)
    }
    
    // 5. Create associations between source and semantic memories
    err = p.createConsolidationAssociations(ctx, memories, semanticMemory)
    
    return err
}
```

### Consolidation Prompt Engineering

```go
func (p *processor) buildConsolidationPrompt(memories []*models.Memory) string {
    var builder strings.Builder
    
    builder.WriteString("You are consolidating episodic memories into semantic knowledge.\n\n")
    builder.WriteString("Episodic Memories to Consolidate:\n")
    
    for i, memory := range memories {
        builder.WriteString(fmt.Sprintf("%d. [%s] %s\n", 
            i+1, 
            time.Unix(memory.Timestamp, 0).Format("2006-01-02 15:04"),
            memory.Content))
    }
    
    builder.WriteString("\nConsolidate these memories into semantic knowledge by:\n")
    builder.WriteString("1. Identifying common themes and patterns\n")
    builder.WriteString("2. Extracting key insights and learnings\n")
    builder.WriteString("3. Noting progression or evolution in understanding\n")
    builder.WriteString("4. Creating concise, actionable knowledge\n\n")
    builder.WriteString("Consolidated Memory:")
    
    return builder.String()
}
```

**Example Prompt Output**:

```
You are consolidating episodic memories into semantic knowledge.

Episodic Memories to Consolidate:
1. [2024-01-20 10:30] User asked about Go concurrency patterns
2. [2024-01-20 11:15] User implemented worker pool with channels  
3. [2024-01-20 14:20] User debugged race condition in Go code
4. [2024-01-20 16:45] User asked about goroutine best practices

Consolidate these memories into semantic knowledge by:
1. Identifying common themes and patterns
2. Extracting key insights and learnings
3. Noting progression or evolution in understanding
4. Creating concise, actionable knowledge

Consolidated Memory:
```

## Practical Exercise: Trigger Consolidation

Let's trigger consolidation and observe the process:

### Setup

Ensure your system is running and has some episodic memories.

### Exercise 1: Manual Consolidation

**Step 1**: Check current memory stats:

```
Use get_stats tool - note the semantic memory count
```

**Step 2**: Trigger consolidation:

```
Use trigger_consolidation tool with force: true
```

**Step 3**: Check logs for consolidation process:

```bash
docker compose logs persistent-context-web | grep -i consolidation
```

**Expected Log Flow**:

```
INFO  Consolidation triggered manually
DEBUG Selecting memories for consolidation
DEBUG Found 8 episodic memories for consolidation  
DEBUG Building consolidation prompt (1,200 tokens)
DEBUG Calling LLM for consolidation
INFO  Created semantic memory id=abc-123
DEBUG Creating consolidation associations
INFO  Consolidation complete
```

**Step 4**: Verify new semantic memory:

```
Use get_stats tool - semantic count should increase
```

### Exercise 2: Threshold-Based Consolidation

**Step 1**: Capture multiple memories rapidly:

```
- capture_memory: "First test memory"
- capture_memory: "Second test memory"  
- capture_memory: "Third test memory"
- (continue until threshold reached)
```

**Step 2**: Watch for automatic consolidation:

```bash
docker compose logs -f persistent-context-web
```

**Expected**: When enough memories accumulate, you should see automatic consolidation triggered.

## Configuration Deep Dive

**Location**: `src/pkg/config/memory.go`

Understanding the configuration helps you tune consolidation behavior:

```go
type MemoryConfig struct {
    // Consolidation triggers
    ConsolidationThreshold uint32 `mapstructure:"consolidation_threshold"` // 50 memories
    
    // Context window management  
    MaxTokens     uint32 `mapstructure:"max_tokens"`      // 3000 tokens
    SafetyMargin  uint32 `mapstructure:"safety_margin"`   // 500 tokens
    TokensPerChar uint32 `mapstructure:"tokens_per_char"` // 4 chars per token
    
    // Scoring weights
    TimeDecayFactor   float64 `mapstructure:"time_decay_factor"`   // 0.1
    FrequencyWeight   float64 `mapstructure:"frequency_weight"`    // 0.3
    AssociationWeight float64 `mapstructure:"association_weight"`  // 0.4
    RelevanceWeight   float64 `mapstructure:"relevance_weight"`    // 0.3
}
```

**Tuning Guide**:

- **ConsolidationThreshold**: Lower = more frequent consolidation, higher = fewer but larger consolidations
- **TimeDecayFactor**: Higher = memories age faster, lower = memories stay relevant longer
- **FrequencyWeight**: Higher = prioritizes often-accessed memories
- **AssociationWeight**: Higher = prioritizes well-connected memories

## Common Issues and Debugging

### Issue 1: "No memories selected for consolidation"

**Symptoms**: Consolidation triggered but no semantic memory created
**Causes**:

- All episodic memories filtered out by time window
- Context window too small for any memories
- No episodic memories exist

**Debug Steps**:

```bash
# Check episodic memory count
curl localhost:8543/api/stats

# Check recent memories
curl "localhost:8543/api/memories?type=episodic&limit=10"

# Check consolidation config
docker compose logs persistent-context-web | grep -i "consolidation config"
```

### Issue 2: "Context window exceeded"

**Symptoms**: Consolidation fails with token limit errors
**Solutions**:

- Increase `MaxTokens` in configuration
- Increase `SafetyMargin`
- Decrease `ConsolidationThreshold`

### Issue 3: "Low-quality consolidated memories"

**Symptoms**: Semantic memories are too generic or miss important details
**Solutions**:

- Adjust scoring weights to prioritize different factors
- Modify consolidation prompt for better instructions
- Check LLM model performance

## Advanced Concepts

### Cascading Consolidation

The system can consolidate semantic memories into higher-level knowledge:

```go
// Future enhancement: Meta-consolidation
if semanticCount > metaThreshold {
    // Consolidate semantic memories into procedural knowledge
    proceduralMemory := consolidateSemanticMemories(semanticMemories)
}
```

### Adaptive Scoring

Scoring weights could adapt based on user behavior:

```go
// Future enhancement: Learning scoring weights
if userFrequentlyAccessesRecentMemories {
    config.TimeDecayFactor *= 0.9 // Keep recent memories longer
}
```

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Event Triggers**: Explain the difference between `EventThresholdReached` and `EventConversationEnd` in terms of when they fire and what they optimize for.

2. **Scoring Algorithm**: A memory has timeScore=0.8, frequencyScore=0.6, associationScore=0.9, relevanceScore=0.7. Using default weights, what's the composite score?

3. **Debug Scenario**: Consolidation is triggered but produces very generic semantic memories. What are 3 potential causes and how would you fix each?

4. **Design Question**: If you wanted to add "user importance tagging" (letting users mark memories as important), where in the scoring algorithm would you add it?

## Connection to Session 14 Work

This consolidation system is the foundation for your Session 14 tasks:

- **Consolidation Triggers**: You may need to implement new event types
- **Memory Scoring**: You might enhance the algorithm with new factors  
- **Association Tracking**: Critical for understanding memory relationships
- **LLM Integration**: Same patterns used for enhanced consolidation

Understanding this system means you can confidently modify any aspect of memory consolidation.

## Notes

<!-- Add your observations as you work through this:
- Which parts of the scoring algorithm felt intuitive vs complex?
- How does the LLM consolidation compare to how you think about summarizing information?
- What questions came up about the context window management?
- Which configuration parameters seem most important to tune?
-->