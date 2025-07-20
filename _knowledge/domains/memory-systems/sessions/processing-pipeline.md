---
domain: memory-systems
name: processing-pipeline
title: Memory Processing Pipeline
duration: 45
status: pending
prerequisites: [vector-databases/fundamentals]
builds_on: [vector-database-fundamentals]
unlocks: [memory-consolidation, event-driven-processing, journal-interface]
complexity: intermediate
---

# Memory Processing Pipeline

## Concept Overview

The memory processing pipeline is the heart of your persistent context system - it's what transforms raw conversation context into structured, searchable memories that can be retrieved later. Think of it as a sophisticated **assembly line for thoughts**.

**Core Problems It Solves:**

- Converting unstructured conversation text into searchable vector representations
- Managing the flow from immediate storage to background processing
- Ensuring memories are captured even if background processing fails
- Coordinating between different system components (LLM, VectorDB, Journal)

**Why This Matters for Session 14:**
Understanding this pipeline is crucial because it's where episodic memories are created, processed, and prepared for consolidation. The consolidation features you'll be working on build directly on this foundation.

## Visualization: Assembly Line Analogy

Think of the memory processing pipeline like a **modern manufacturing assembly line**:

**Traditional Assembly Line (Car Manufacturing):**

1. **Raw Materials** → Chassis frame
2. **Assembly Station 1** → Add engine
3. **Assembly Station 2** → Add body panels  
4. **Quality Control** → Test and inspect
5. **Finished Product** → Ready car

**Memory Processing Pipeline:**

1. **Raw Context** → "User asked about Go concurrency"
2. **Parsing Station** → Convert to Memory struct
3. **Embedding Station** → Generate vector representation
4. **Storage Station** → Save to VectorDB
5. **Event Station** → Queue for consolidation
6. **Finished Memory** → Searchable, retrievable memory

Just like a car assembly line, each station can work independently, and if one station has problems, the others keep running.

## Prerequisites Check

Before starting, ensure you understand:

- [x] Vector databases store meaning as numbers (Session 001)
- [x] HTTP APIs receive and respond to requests
- [x] Go interfaces define behavior contracts
- [ ] How Go channels work for async communication (we'll cover this)

## Step-by-Step Data Flow

### Step 1: Context Capture (API Entry Point)

**Location**: `src/persistent-context-svc/app/server.go:handleCaptureMemory()`

When you call the MCP tool `capture_memory` from Claude Code, here's what happens:

```go
// 1. HTTP request arrives
POST /api/memory
{
    "content": "User asked about Go concurrency patterns",
    "metadata": {"session_id": "claude-session-123"}
}

// 2. Gin handler parses the request
func (s *Server) handleCaptureMemory(c *gin.Context) {
    var req CaptureMemoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // 3. Create Memory domain object
    memory := &models.Memory{
        ID:        uuid.New().String(),
        Content:   req.Content,          // The actual text
        Type:      models.MemoryTypeEpisodic, // Always episodic for new captures
        Metadata:  req.Metadata,
        Timestamp: time.Now().Unix(),    // When it was captured
    }
```

**Key Insight**: At this point, the memory exists but has no vector embedding yet. It's like a car chassis without an engine.

### Step 2: Journal Interface (Coordination Layer)

**Location**: `src/pkg/journal/journal.go:CaptureMemory()`

The Journal acts like a **production manager** - it coordinates the process but doesn't do the heavy lifting:

```go
func (j *journal) CaptureMemory(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // 1. Trigger async processing (fire-and-forget)
    j.processor.ProcessMemory(ctx, memory)
    
    // 2. Store immediately (without embedding for now)
    return j.Store(ctx, memory)
}
```

**Key Design Pattern**: This demonstrates Go's approach to service coordination - the Journal orchestrates operations while delegating specialized work to other components.

**Why This Design?**: The API responds immediately (good user experience) while background processing happens separately (resilience).

### Step 3: Memory Processor (Event Queue)

**Location**: `src/pkg/memory/processor.go:ProcessMemory()`

This is where the **Go concurrency magic** happens. The processor uses channels (Go's async communication):

```go
func (p *processor) ProcessMemory(ctx context.Context, memory *models.Memory) {
    // Create processing event
    event := Event{
        Type:      EventNewContext,
        Memory:    memory,
        Timestamp: time.Now(),
    }
    
    // Try to add to queue (non-blocking)
    select {
    case p.eventQueue <- event:
        // Event queued successfully
        p.logger.Debug("Memory queued for processing", "id", memory.ID)
    default:
        // Queue is full - log but don't fail the request
        p.logger.Warn("Event queue full, dropping event", "id", memory.ID)
    }
}
```

**Go Concurrency Pattern**: This `select` statement demonstrates Go's approach to non-blocking operations:

- **Channel Send**: Attempt to send to the event queue
- **Default Case**: If queue is full, handle gracefully without blocking
- **Overflow Protection**: System remains responsive even under high load

### Step 4: Background Processing (Goroutine)

**Location**: `src/pkg/memory/processor.go:processEvents()`

This runs in a separate goroutine for background processing:

```go
func (p *processor) processEvents(ctx context.Context) {
    for {
        select {
        case event := <-p.eventQueue:
            // Process one event at a time
            p.handleEvent(ctx, event)
            
        case <-ctx.Done():
            // Graceful shutdown
            return
        }
    }
}

func (p *processor) handleEvent(ctx context.Context, event Event) {
    switch event.Type {
    case EventNewContext:
        // 1. Generate embedding
        embedding, err := p.llm.GenerateEmbedding(ctx, event.Memory.Content)
        if err != nil {
            p.logger.Error("Failed to generate embedding", "error", err)
            return // Skip this memory, don't crash
        }
        
        // 2. Update memory with embedding
        event.Memory.Embedding = embedding
        
        // 3. Store in vector database
        err = p.vectorDB.Store(ctx, event.Memory)
        if err != nil {
            p.logger.Error("Failed to store memory", "error", err)
        }
    }
}
```

**Go Pattern**: This demonstrates Go's approach to long-running background processes:
- **Infinite Loop**: `for` loop keeps the goroutine alive
- **Channel Receive**: Blocks until events arrive
- **Context Cancellation**: Graceful shutdown when context is cancelled
- **Sequential Processing**: Events handled one at a time for safety

### Step 5: LLM Integration (Embedding Generation)

**Location**: `src/pkg/llm/ollama.go:GenerateEmbedding()`

This converts text into the vector representation:

```go
func (o *ollama) GenerateEmbedding(ctx context.Context, content string) ([]float64, error) {
    // Call Ollama API
    resp, err := o.client.Embeddings(ctx, &api.EmbeddingRequest{
        Model:  "phi3:mini",  // 3072-dimensional model
        Prompt: content,
    })
    if err != nil {
        return nil, fmt.Errorf("ollama embedding request failed: %w", err)
    }
    
    return resp.Embedding, nil
}
```

**Data Transformation**:

- Input: `"User asked about Go concurrency patterns"`
- Output: `[]float64{0.123, -0.456, 0.789, ...}` (3072 numbers)

This is like taking a fingerprint of the meaning.

### Step 6: Vector Database Storage

**Location**: `src/pkg/vectordb/qdrantdb.go:Store()`

Finally, the memory with its embedding gets stored in Qdrant:

```go
func (q *qdrant) Store(ctx context.Context, memory *models.Memory) error {
    // Convert to Qdrant format
    point := &qdrant.PointStruct{
        Id: &qdrant.PointId{
            PointIdOptions: &qdrant.PointId_Uuid{Uuid: memory.ID},
        },
        Vectors: &qdrant.Vectors{
            VectorsOptions: &qdrant.Vectors_Vector{
                Vector: &qdrant.Vector{Data: memory.Embedding},
            },
        },
        Payload: map[string]*qdrant.Value{
            "content":    {Kind: &qdrant.Value_StringValue{StringValue: memory.Content}},
            "type":       {Kind: &qdrant.Value_StringValue{StringValue: string(memory.Type)}},
            "created_at": {Kind: &qdrant.Value_IntegerValue{IntegerValue: memory.Timestamp}},
        },
    }
    
    // Store in appropriate collection
    collectionName := q.getCollectionName(memory.Type) // "episodic_memories"
    _, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
        CollectionName: collectionName,
        Points:         []*qdrant.PointStruct{point},
    })
    
    return err
}
```

## Practical Exercise: Trace a Memory

Let's trace an actual memory through the system:

### Setup

1. Ensure your Docker stack is running: `docker compose up -d`
2. Build the MCP binary: `cd src && go build -o bin/persistent-context-mcp ./cmd/persistent-context-mcp/`

### Exercise 1: Capture and Trace

**Step 1**: Capture a memory using Claude Code's MCP tool:

```
Use the capture_memory tool with content: "Learning about memory processing pipeline"
```

**Step 2**: Check the logs to see the processing:

```bash
docker compose logs persistent-context-web
```

**Expected Log Flow**:

```
INFO  Memory capture request received
DEBUG Memory queued for processing id=abc-123
DEBUG Generating embedding for content=Learning about...
DEBUG Storing memory in vector database
INFO  Memory processing complete
```

### Exercise 2: Verify Storage

**Step 1**: Use the stats tool to see the memory count:

```
Use the get_stats tool
```

**Step 2**: Search for your memory:

```
Use the search_memories tool with content: "memory processing"
```

**Expected Result**: Your memory should appear in the search results with a high similarity score.

## Common Issues and Solutions

### Issue 1: "Memory captured but not searchable"

**Symptoms**: Memory appears in stats but not in search results
**Cause**: Embedding generation failed
**Debug**: Check logs for "Failed to generate embedding"
**Solution**: Ensure Ollama is running and accessible

### Issue 2: "Event queue full" warnings

**Symptoms**: Warning logs about dropped events
**Cause**: Background processing is slower than memory capture rate
**Solution**: This is by design - the system prioritizes API responsiveness

### Issue 3: "Vector dimension mismatch"

**Symptoms**: Error storing in Qdrant
**Cause**: Embedding model change without updating collection
**Solution**: Recreate collections or use consistent model

## Go Patterns Deep Dive

### Channel Pattern (Event Queue)

```go
// Channel creation (buffered for performance)
eventQueue := make(chan Event, 1000)

// Non-blocking send with overflow protection
select {
case eventQueue <- event:
    // Success
default:
    // Queue full - handle gracefully
}

// Blocking receive in background goroutine
for event := range eventQueue {
    // Process events one by one
}
```

**Key Go Concepts**:
- **Buffered Channels**: Pre-allocate space for performance
- **Non-blocking Operations**: Use `select` with `default` for graceful handling
- **Range over Channel**: Simple pattern for consuming all events
- **Channel Closure**: Closing channels signals completion to receivers

### Interface Composition Pattern

```go
// Small, focused interfaces
type LLM interface {
    GenerateEmbedding(ctx context.Context, content string) ([]float64, error)
}

type VectorDB interface {
    Store(ctx context.Context, memory *models.Memory) error
}

// Composed in larger service
type processor struct {
    llm      LLM
    vectorDB VectorDB
    // Dependencies injected via struct fields
}
```

**Key Design Benefits**:

- **Testability**: Easy to mock interfaces for unit testing
- **Flexibility**: Swap implementations without changing dependent code
- **Clear Dependencies**: Required services are explicit in struct definition
- **Decoupling**: High-level policies don't depend on low-level implementation details

## Comprehension Checkpoint

Answer these questions to validate your understanding:

1. **Trace the Flow**: Describe what happens to a memory between the HTTP request and final storage. Include which goroutine handles each step.

2. **Identify the Pattern**: Why does the system use "fire-and-forget" for embedding generation instead of blocking the API response?

3. **Debug This Scenario**: A user reports that memories are being captured but are not showing up in search results. What are the 3 most likely causes and how would you debug each?

4. **Design Question**: If you needed to add a new processing step (e.g., content filtering), where in the pipeline would you add it and why?

## Connection to Session 14

This pipeline directly enables the consolidation work you'll be doing:

- **Event Queue**: Already handles consolidation events
- **Memory Selection**: Processor can query stored memories
- **LLM Integration**: Same interface used for consolidation prompts
- **Storage Layer**: Where consolidated semantic memories are stored

Understanding this pipeline means you can confidently modify consolidation triggers, memory selection algorithms, and scoring systems.

## Notes

<!-- Add your observations as you work through this:
- Which parts felt clear vs confusing?
- What questions came up about the async processing design?
- Which Go patterns were most intuitive vs surprising?
- Which debugging techniques were most helpful?
-->