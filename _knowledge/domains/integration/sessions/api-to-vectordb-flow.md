---
domain: integration
name: api-to-vectordb-flow
title: API to VectorDB Data Flow
duration: 30
status: pending
prerequisites: [vector-databases/fundamentals, memory-systems/processing-pipeline]
builds_on: [vector-database-fundamentals, memory-processing-pipeline]
unlocks: [error-handling-patterns, api-design-patterns, system-debugging]
complexity: intermediate
---

# API to VectorDB Data Flow

## Concept Overview

Understanding the complete data flow from HTTP API endpoints to vector database storage is crucial for debugging, optimization, and feature development. This session traces a memory through every system component, transformation, and integration point.

**Core Problems This Solves:**

- Understanding where data transformations happen and why
- Debugging issues by knowing which component to examine
- Optimizing performance by identifying bottlenecks
- Designing new features by understanding existing patterns

**Why This Matters for Session 14:**
You'll need to understand this flow to implement consolidation triggers, debug memory processing issues, and optimize the system's performance.

## Visualization: Complete System Flow

```
Claude Code MCP Tool
         ↓
   HTTP API Call (POST /api/memory)
         ↓
   Gin HTTP Handler (handleCaptureMemory)
         ↓
   JSON → Domain Model Transformation
         ↓
   Journal Interface (CaptureMemory)
         ↓
   Memory Processor (ProcessMemory)
         ↓
   Event Queue (Go Channel)
         ↓
   Background Goroutine (processEvents)
         ↓
   LLM Service (GenerateEmbedding)
         ↓
   VectorDB Interface (Store)
         ↓
   Qdrant Implementation (gRPC)
         ↓
   Persistent Storage
```

## Prerequisites Check

Before starting, ensure you understand:

- [x] Vector databases store embeddings for similarity search
- [x] Go interfaces provide abstraction boundaries
- [x] HTTP APIs handle request/response cycles
- [ ] How Go channels enable async processing (we'll cover this)

## Step-by-Step Data Flow Trace

### Step 1: MCP Tool Invocation

**Starting Point**: When you use `capture_memory` in Claude Code

```bash
# MCP tool translates to HTTP call
curl -X POST http://localhost:8543/api/memory \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Learning about API data flow",
    "metadata": {"session_id": "claude-session-123"}
  }'
```

**Key Insight**: MCP tools are just convenient wrappers around HTTP API calls.

### Step 2: HTTP Handler Reception

**Location**: `src/persistent-context-svc/app/server.go:handleCaptureMemory()`

```go
func (s *Server) handleCaptureMemory(c *gin.Context) {
    // 1. Parse incoming JSON request
    var req CaptureMemoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }
    
    // 2. Validate required fields
    if req.Content == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
        return
    }
    
    // 3. Create domain model from HTTP request
    memory := &models.Memory{
        ID:        uuid.New().String(),
        Content:   req.Content,
        Type:      models.MemoryTypeEpisodic,
        Metadata:  req.Metadata,
        Timestamp: time.Now().Unix(),
        // Note: Embedding is nil at this point
    }
    
    // 4. Call business logic layer
    result, err := s.journal.CaptureMemory(c.Request.Context(), memory)
    if err != nil {
        s.logger.Error("Failed to capture memory", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to capture memory"})
        return
    }
    
    // 5. Return success response
    c.JSON(http.StatusCreated, gin.H{
        "id":         result.ID,
        "message":    "Memory captured successfully",
        "timestamp":  result.Timestamp,
    })
}
```

**Data Transformation**: JSON → Go struct → Domain Model

**Error Handling Pattern**: Validate early, return specific error messages

### Step 3: Journal Interface Layer

**Location**: `src/pkg/journal/journal.go:CaptureMemory()`

```go
func (j *journal) CaptureMemory(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // 1. Immediate storage (without embedding)
    storedMemory, err := j.Store(ctx, memory)
    if err != nil {
        return nil, fmt.Errorf("failed to store memory: %w", err)
    }
    
    // 2. Trigger async processing (fire-and-forget)
    j.processor.ProcessMemory(ctx, storedMemory)
    
    // 3. Return immediately (don't wait for embedding)
    return storedMemory, nil
}
```

**Key Design Decision**: Store first, process embeddings asynchronously

**Why This Pattern**: 
- Fast API response (good UX)
- Resilient to embedding failures
- Memories are never lost due to processing issues

### Step 4: Memory Processor Queueing

**Location**: `src/pkg/memory/processor.go:ProcessMemory()`

```go
func (p *processor) ProcessMemory(ctx context.Context, memory *models.Memory) {
    // 1. Create processing event
    event := Event{
        Type:      EventNewContext,
        Memory:    memory,
        Timestamp: time.Now(),
    }
    
    // 2. Non-blocking queue insertion
    select {
    case p.eventQueue <- event:
        p.logger.Debug("Memory queued for processing", "id", memory.ID)
    default:
        // Queue is full - graceful degradation
        p.logger.Warn("Event queue full, dropping event", "id", memory.ID)
        // Note: Memory is still stored, just won't get embedding
    }
}
```

**Go Pattern**: `select` with `default` for non-blocking operations

**Graceful Degradation**: System continues working even under high load

### Step 5: Background Processing

**Location**: `src/pkg/memory/processor.go:processEvents()`

```go
func (p *processor) processEvents(ctx context.Context) {
    // Long-running goroutine
    for {
        select {
        case event := <-p.eventQueue:
            // Process one event at a time
            p.handleEvent(ctx, event)
            
        case <-ctx.Done():
            // Graceful shutdown
            p.logger.Info("Processing stopped due to context cancellation")
            return
        }
    }
}

func (p *processor) handleEvent(ctx context.Context, event Event) {
    switch event.Type {
    case EventNewContext:
        p.processNewMemory(ctx, event.Memory)
    case EventConsolidationTrigger:
        p.processConsolidation(ctx)
    }
}
```

**Go Concurrency Pattern**: Goroutine with channel-based event loop

**Context Usage**: Proper cancellation and timeout handling

### Step 6: LLM Embedding Generation

**Location**: `src/pkg/llm/ollama.go:GenerateEmbedding()`

```go
func (o *ollama) GenerateEmbedding(ctx context.Context, content string) ([]float64, error) {
    // 1. Call external Ollama service
    resp, err := o.client.Embeddings(ctx, &api.EmbeddingRequest{
        Model:  "phi3:mini",  // 3072-dimensional model
        Prompt: content,
    })
    if err != nil {
        return nil, fmt.Errorf("ollama embedding request failed: %w", err)
    }
    
    // 2. Validate response
    if len(resp.Embedding) == 0 {
        return nil, fmt.Errorf("empty embedding returned")
    }
    
    // 3. Return vector representation
    return resp.Embedding, nil
}
```

**External Integration**: HTTP call to Ollama service running in Docker

**Data Transformation**: String → float64 array (3072 dimensions)

**Error Handling**: Wrap external errors with context

### Step 7: VectorDB Storage

**Location**: `src/pkg/vectordb/qdrantdb.go:Store()`

```go
func (q *qdrant) Store(ctx context.Context, memory *models.Memory) error {
    // 1. Convert domain model to Qdrant format
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
            "metadata":   {Kind: &qdrant.Value_StringValue{StringValue: q.serializeMetadata(memory.Metadata)}},
        },
    }
    
    // 2. Determine collection based on memory type
    collectionName := q.getCollectionName(memory.Type) // "episodic_memories"
    
    // 3. Store via gRPC
    _, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
        CollectionName: collectionName,
        Points:         []*qdrant.PointStruct{point},
    })
    
    if err != nil {
        return fmt.Errorf("failed to store in qdrant: %w", err)
    }
    
    return nil
}
```

**Data Transformation**: Domain Model → Qdrant Point Structure

**Collection Strategy**: Different memory types stored in separate collections

**gRPC Integration**: Type-safe protocol buffer communication

## Error Flow Analysis

Understanding how errors propagate backwards through the system:

### Error Origination Points

1. **Qdrant Storage Error**:
   ```go
   // VectorDB layer
   _, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{...})
   if err != nil {
       return fmt.Errorf("failed to store in qdrant: %w", err)
   }
   ```

2. **LLM Service Error**:
   ```go
   // LLM layer
   resp, err := o.client.Embeddings(ctx, &api.EmbeddingRequest{...})
   if err != nil {
       return nil, fmt.Errorf("ollama embedding request failed: %w", err)
   }
   ```

3. **Journal Interface Error**:
   ```go
   // Journal layer
   storedMemory, err := j.Store(ctx, memory)
   if err != nil {
       return nil, fmt.Errorf("failed to store memory: %w", err)
   }
   ```

### Error Propagation Pattern

```
Qdrant Error → VectorDB Interface → Memory Processor → Journal → HTTP Handler → JSON Response
```

**Error Wrapping**: Each layer adds context using `fmt.Errorf("context: %w", err)`

**Error Boundaries**: HTTP handler converts all errors to appropriate HTTP status codes

## Practical Exercise: Trace a Real Request

### Setup

Ensure your system is running:

```bash
docker compose up -d
cd src && go build -o bin/persistent-context-mcp ./cmd/persistent-context-mcp/
```

### Exercise 1: Complete Flow Tracing

**Step 1**: Enable debug logging:

```bash
# Check current log level
docker compose logs persistent-context-web | head -5

# If needed, restart with debug level
```

**Step 2**: Capture a memory with MCP tool:

```
Use capture_memory with content: "Tracing data flow through the system"
```

**Step 3**: Follow the logs in real-time:

```bash
docker compose logs -f persistent-context-web
```

**Expected Log Sequence**:

```
INFO  HTTP request received POST /api/memory
DEBUG Memory created with ID abc-123
DEBUG Memory queued for processing
DEBUG Generating embedding for content=Tracing data flow...
DEBUG Embedding generated successfully (3072 dimensions)
DEBUG Storing memory in vector database
INFO  Memory processing complete
```

### Exercise 2: Error Injection Testing

**Step 1**: Stop Ollama to simulate LLM failure:

```bash
docker compose stop ollama
```

**Step 2**: Capture a memory:

```
Use capture_memory with content: "Testing error handling"
```

**Step 3**: Observe error handling:

```bash
docker compose logs persistent-context-web | tail -10
```

**Expected Behavior**: 
- Memory still gets stored (immediate storage)
- Background processing logs error
- API returns success (graceful degradation)

**Step 4**: Restart Ollama:

```bash
docker compose start ollama
```

### Exercise 3: Performance Analysis

**Step 1**: Capture multiple memories rapidly:

```
Quickly use capture_memory 5 times with different content
```

**Step 2**: Analyze response times:

```bash
# Check API response times in logs
docker compose logs persistent-context-web | grep "request completed"

# Check queue behavior
docker compose logs persistent-context-web | grep "queue"
```

**Observations**:
- API responses should be fast (< 50ms)
- Background processing takes longer
- Queue prevents blocking under load

## Data Transformation Deep Dive

### Transformation 1: HTTP JSON → Go Struct

**Input**:
```json
{
  "content": "Learning about data flow",
  "metadata": {"session_id": "claude-123"}
}
```

**Output**:
```go
CaptureMemoryRequest{
    Content:  "Learning about data flow",
    Metadata: map[string]interface{}{"session_id": "claude-123"},
}
```

### Transformation 2: Request → Domain Model

**Input**: CaptureMemoryRequest

**Output**:
```go
models.Memory{
    ID:        "abc-123-def-456",
    Content:   "Learning about data flow",
    Type:      models.MemoryTypeEpisodic,
    Metadata:  map[string]interface{}{"session_id": "claude-123"},
    Timestamp: 1642765432,
    Embedding: nil, // Added later
}
```

### Transformation 3: Text → Vector

**Input**: "Learning about data flow"

**Processing**: Ollama phi3:mini model

**Output**: `[]float64{0.123, -0.456, 0.789, ...}` (3072 values)

### Transformation 4: Domain Model → Qdrant Point

**Input**: models.Memory with embedding

**Output**:
```go
qdrant.PointStruct{
    Id: &qdrant.PointId{Uuid: "abc-123-def-456"},
    Vectors: &qdrant.Vectors{Vector: &qdrant.Vector{Data: embedding}},
    Payload: map[string]*qdrant.Value{
        "content":    {StringValue: "Learning about data flow"},
        "type":       {StringValue: "episodic"},
        "created_at": {IntegerValue: 1642765432},
    },
}
```

## Performance Considerations

### Bottleneck Analysis

1. **LLM Embedding Generation**: ~100-500ms per memory
   - **Solution**: Async processing prevents blocking API
   - **Monitoring**: Track embedding queue depth

2. **Vector Database Storage**: ~10-50ms per memory
   - **Solution**: Batch operations for multiple memories
   - **Monitoring**: Track Qdrant response times

3. **HTTP JSON Parsing**: ~1-5ms per request
   - **Solution**: Already optimized with Gin
   - **Monitoring**: Track request parsing times

### Optimization Strategies

1. **Batch Processing**: Group multiple memories for single LLM call
2. **Connection Pooling**: Reuse HTTP connections to Ollama
3. **Memory Pooling**: Reuse Go structs to reduce allocations
4. **Async Responses**: Return before embedding completion (already implemented)

## Debugging Guide

### Common Issues and Diagnostic Steps

**Issue 1**: "Memory not appearing in search results"

**Diagnostic Flow**:
1. Check if memory was stored: `GET /api/stats`
2. Check embedding generation: Search logs for embedding errors
3. Check vector storage: Verify Qdrant connection
4. Check search parameters: Verify similarity thresholds

**Issue 2**: "API timeouts under load"

**Diagnostic Flow**:
1. Check queue depth: Look for "queue full" warnings
2. Check LLM service: Verify Ollama is responsive
3. Check database: Verify Qdrant performance
4. Check resource limits: Memory, CPU, network

**Issue 3**: "Data corruption or loss"

**Diagnostic Flow**:
1. Check immediate storage: Memory should exist even without embedding
2. Check error propagation: Look for wrapped error messages
3. Check transaction boundaries: Verify rollback behavior
4. Check data validation: Verify input sanitization

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Flow Sequence**: Describe the path a memory takes from MCP tool to vector storage, including which components handle what transformations.

2. **Error Handling**: Explain why the system stores memories immediately but generates embeddings asynchronously. What happens if embedding generation fails?

3. **Performance Design**: Why does the system use a queue for memory processing instead of handling everything synchronously in the HTTP handler?

4. **Debug Scenario**: A user reports memories are being captured but not appearing in search results. Walk through your debugging approach using the data flow knowledge.

## Connection to Session 14

This data flow understanding directly supports Session 14 work:

- **Consolidation Triggers**: Understanding event queue patterns
- **Memory Selection**: Knowing how memories are stored and retrieved
- **Performance Optimization**: Identifying bottlenecks in the processing pipeline
- **Error Handling**: Debugging consolidation and processing issues

## Notes

<!-- Add your observations as you work through this:
- Which transformations felt most complex vs straightforward?
- How does the async pattern compare to synchronous processing you've worked with?
- What questions came up about error propagation patterns?
- Which performance bottlenecks seemed most concerning?
-->