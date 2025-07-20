# API to VectorDB Data Flow Documentation

This document traces the complete data flow for each API endpoint from HTTP request to vector database storage and back. Each section includes code references and data transformations at each stage.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [POST /api/memory - Capture Memory](#post-apimemory---capture-memory)
3. [GET /api/memories - Query Memories](#get-apimemories---query-memories)
4. [GET /api/memory/:id - Retrieve Specific Memory](#get-apimemoryid---retrieve-specific-memory)
5. [POST /api/consolidate - Trigger Consolidation](#post-apiconsolidate---trigger-consolidation)
6. [GET /api/stats - Get Statistics](#get-apistats---get-statistics)
7. [Common Patterns](#common-patterns)
8. [Error Handling Flow](#error-handling-flow)

## Architecture Overview

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│ HTTP Client │────▶│ Gin Handlers │────▶│   Journal   │────▶│   Memory     │
│ (MCP/User)  │     │  (server.go) │     │ (journal.go)│     │ Processor    │
└─────────────┘     └──────────────┘     └─────────────┘     └──────┬───────┘
                                                                      │
                                         ┌──────────────┐            │
                                         │     LLM      │◀───────────┘
                                         │  (Ollama)    │
                                         └──────┬───────┘
                                                │
                                         ┌──────▼───────┐
                                         │   VectorDB   │
                                         │   (Qdrant)   │
                                         └──────────────┘
```

## POST /api/memory - Capture Memory

### 1. HTTP Request Entry

**Location**: `src/persistent-context-svc/app/server.go:setupRoutes()`

```go
// Route registration
r.POST("/api/memory", s.handleCaptureMemory)
```

### 2. Request Parsing & Validation

**Location**: `src/persistent-context-svc/app/server.go:handleCaptureMemory()`

```go
// Input structure
type CaptureMemoryRequest struct {
    Content  string                 `json:"content"`
    Metadata map[string]interface{} `json:"metadata"`
}

// Parse request
var req CaptureMemoryRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "Invalid request"})
    return
}
```

**Data at this stage**:

```json
{
    "content": "User asked about Go concurrency patterns",
    "metadata": {
        "session_id": "claude-123",
        "timestamp": "2024-01-20T10:30:00Z"
    }
}
```

### 3. Journal Interface Call

**Location**: `src/persistent-context-svc/app/server.go:handleCaptureMemory()` → `src/pkg/journal/journal.go:CaptureMemory()`

```go
// Create Memory domain object
memory := &models.Memory{
    ID:        uuid.New().String(),
    Content:   req.Content,
    Type:      models.MemoryTypeEpisodic, // Default for captured memories
    Metadata:  req.Metadata,
    Timestamp: time.Now().Unix(),
}

// Journal processes the memory
processedMemory, err := s.journal.CaptureMemory(ctx, memory)
```

### 4. Memory Processing Pipeline

**Location**: `src/pkg/journal/journal.go:CaptureMemory()` → `src/pkg/memory/processor.go:ProcessMemory()`

```go
// Journal delegates to processor
func (j *journal) CaptureMemory(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // Trigger async processing
    j.processor.ProcessMemory(ctx, memory)
    
    // Store immediately (fire-and-forget for embeddings)
    return j.Store(ctx, memory)
}
```

### 5. Async Event Queue

**Location**: `src/pkg/memory/processor.go:ProcessMemory()`

```go
// Memory added to event queue
event := Event{
    Type:      EventNewContext,
    Memory:    memory,
    Timestamp: time.Now(),
}

select {
case p.eventQueue <- event:
    // Event queued successfully
default:
    // Queue full, log and continue
}
```

### 6. Background Processing (Goroutine)

**Location**: `src/pkg/memory/processor.go:processEvents()`

```go
// Running in background goroutine
for event := range p.eventQueue {
    switch event.Type {
    case EventNewContext:
        // Generate embedding
        embedding, err := p.llm.GenerateEmbedding(ctx, event.Memory.Content)
        if err != nil {
            // Log error, continue processing
            continue
        }
        
        // Update memory with embedding
        event.Memory.Embedding = embedding
        
        // Store in vector database
        err = p.vectorDB.Store(ctx, event.Memory)
    }
}
```

### 7. LLM Embedding Generation

**Location**: `src/pkg/llm/ollama.go:GenerateEmbedding()`

```go
// Call Ollama API
resp, err := o.client.Embeddings(ctx, &api.EmbeddingRequest{
    Model:  "phi3:mini", // 3072-dimensional embeddings
    Prompt: content,
})

// Result: []float64 with 3072 dimensions
embedding := resp.Embedding
```

**Data transformation**:

- Input: "User asked about Go concurrency patterns"
- Output: `[]float64{0.123, -0.456, 0.789, ...}` (3072 values)

### 8. Vector Database Storage

**Location**: `src/pkg/vectordb/qdrantdb.go:Store()`

```go
// Convert to Qdrant point
point := &qdrant.PointStruct{
    Id: &qdrant.PointId{
        PointIdOptions: &qdrant.PointId_Uuid{
            Uuid: memory.ID,
        },
    },
    Vectors: &qdrant.Vectors{
        VectorsOptions: &qdrant.Vectors_Vector{
            Vector: &qdrant.Vector{
                Data: memory.Embedding,
            },
        },
    },
    Payload: map[string]*qdrant.Value{
        "content":    {Kind: &qdrant.Value_StringValue{StringValue: memory.Content}},
        "type":       {Kind: &qdrant.Value_StringValue{StringValue: string(memory.Type)}},
        "created_at": {Kind: &qdrant.Value_IntegerValue{IntegerValue: memory.Timestamp}},
        // Additional metadata fields...
    },
}

// Store in collection
_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
    CollectionName: "episodic_memories", // Based on memory type
    Points:         []*qdrant.PointStruct{point},
})
```

### 9. Response to Client

**Location**: `src/persistent-context-svc/app/server.go:handleCaptureMemory()`

```go
// Return success response
c.JSON(http.StatusOK, gin.H{
    "id":        processedMemory.ID,
    "message":   "Memory captured successfully",
    "timestamp": processedMemory.Timestamp,
})
```

**Final response**:

```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Memory captured successfully",
    "timestamp": 1705749000
}
```

## GET /api/memories - Query Memories

### 1. HTTP Request Entry

**Location**: `src/persistent-context-svc/app/server.go:handleQueryMemories()`

```go
// Query parameters
content := c.Query("content")      // Search query
memoryType := c.Query("type")      // Optional: filter by type
limit := c.DefaultQuery("limit", "10")
```

### 2. Generate Query Embedding

**Location**: `src/pkg/journal/journal.go:QueryMemories()`

```go
// Generate embedding for search query
queryEmbedding, err := j.llm.GenerateEmbedding(ctx, query)
if err != nil {
    return nil, fmt.Errorf("failed to generate query embedding: %w", err)
}
```

### 3. Vector Similarity Search

**Location**: `src/pkg/vectordb/qdrantdb.go:Query()`

```go
// Perform similarity search
searchResult, err := q.client.Query(ctx, &qdrant.QueryPoints{
    CollectionName: collectionName,
    Query: &qdrant.QueryInterface{
        Query: &qdrant.QueryInterface_Nearest{
            Nearest: &qdrant.VectorInput{
                VectorInput: &qdrant.VectorInput_Dense{
                    Dense: &qdrant.DenseVector{
                        Data: queryEmbedding,
                    },
                },
            },
        },
    },
    Limit:      &limit,
    WithPayload: &qdrant.WithPayloadSelector{
        SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
    },
})
```

### 4. Transform Results

**Location**: `src/pkg/journal/journal.go:QueryMemories()`

```go
// Convert Qdrant results to domain models
memories := make([]*models.Memory, 0, len(searchResult.Result))
for _, point := range searchResult.Result {
    memory := &models.Memory{
        ID:        point.Id.GetUuid(),
        Content:   point.Payload["content"].GetStringValue(),
        Type:      models.MemoryType(point.Payload["type"].GetStringValue()),
        Score:     point.Score,
        Timestamp: point.Payload["created_at"].GetIntegerValue(),
        // ... additional fields
    }
    memories = append(memories, memory)
}
```

### 5. Response Formation

**Location**: `src/persistent-context-svc/app/server.go:handleQueryMemories()`

```go
// Return formatted response
c.JSON(http.StatusOK, gin.H{
    "memories": memories,
    "count":    len(memories),
    "query":    content,
})
```

## GET /api/memory/:id - Retrieve Specific Memory

### 1. HTTP Request Entry

**Location**: `src/persistent-context-svc/app/server.go:handleGetMemory()`

```go
// Route registration
r.GET("/api/memory/:id", s.handleGetMemory)

// Extract ID from URL path
memoryID := c.Param("id")
if memoryID == "" {
    c.JSON(400, gin.H{"error": "Memory ID required"})
    return
}
```

### 2. Journal Interface Call

**Location**: `src/pkg/journal/journal.go:Retrieve()`

```go
// Direct retrieval by ID
memory, err := s.journal.Retrieve(ctx, memoryID)
if err != nil {
    if errors.Is(err, ErrMemoryNotFound) {
        c.JSON(404, gin.H{"error": "Memory not found"})
        return
    }
    c.JSON(500, gin.H{"error": "Failed to retrieve memory"})
    return
}
```

### 3. Vector Database Lookup

**Location**: `src/pkg/vectordb/qdrantdb.go:Retrieve()`

```go
// Get point by ID from Qdrant
response, err := q.client.Get(ctx, &qdrant.GetPoints{
    CollectionName: collectionName,
    Ids: []*qdrant.PointId{{
        PointIdOptions: &qdrant.PointId_Uuid{
            Uuid: memoryID,
        },
    }},
    WithPayload: &qdrant.WithPayloadSelector{
        SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
    },
    WithVector: &qdrant.WithVectorsSelector{
        SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true},
    },
})
```

### 4. Response Formation

**Location**: `src/persistent-context-svc/app/server.go:handleGetMemory()`

```go
// Return the memory with all fields
c.JSON(http.StatusOK, gin.H{
    "memory": memory,
})
```

**Response structure**:

```json
{
    "memory": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "content": "User asked about Go concurrency patterns",
        "type": "episodic",
        "timestamp": 1705749000,
        "score": 0.95,
        "metadata": {
            "session_id": "claude-123"
        }
    }
}
```

## GET /api/stats - Get Statistics

### 1. HTTP Request Entry

**Location**: `src/persistent-context-svc/app/server.go:handleGetStats()`

```go
// Route registration
r.GET("/api/stats", s.handleGetStats)

// No request parameters needed
```

### 2. Journal Interface Call

**Location**: `src/pkg/journal/journal.go:GetStats()`

```go
// Get statistics from journal
stats, err := s.journal.GetStats(ctx)
if err != nil {
    c.JSON(500, gin.H{"error": "Failed to retrieve statistics"})
    return
}
```

### 3. Vector Database Collection Info

**Location**: `src/pkg/vectordb/qdrantdb.go:GetStats()`

```go
// Get collection info for each memory type
collections := []string{"episodic_memories", "semantic_memories", "procedural_memories"}
stats := &models.JournalStats{}

for _, collection := range collections {
    info, err := q.client.CollectionInfo(ctx, &qdrant.GetCollectionInfoRequest{
        CollectionName: collection,
    })
    if err != nil {
        continue // Skip failed collections
    }
    
    switch collection {
    case "episodic_memories":
        stats.EpisodicCount = info.Result.PointsCount
    case "semantic_memories":
        stats.SemanticCount = info.Result.PointsCount
    case "procedural_memories":
        stats.ProceduralCount = info.Result.PointsCount
    }
}

// Calculate totals
stats.TotalMemories = stats.EpisodicCount + stats.SemanticCount + stats.ProceduralCount
```

### 4. Memory Processor Stats

**Location**: `src/pkg/memory/processor.go:GetStats()`

```go
// Get processing statistics
p.mu.RLock()
defer p.mu.RUnlock()

processorStats := &models.ProcessorStats{
    QueueLength:        len(p.eventQueue),
    EventsProcessed:    p.eventsProcessed,
    ConsolidationCount: p.consolidationCount,
    LastConsolidation:  p.lastConsolidation,
}
```

### 5. Response Formation

**Location**: `src/persistent-context-svc/app/server.go:handleGetStats()`

```go
// Combine all statistics
response := gin.H{
    "journal": stats,
    "processor": processorStats,
    "collections": gin.H{
        "episodic":   stats.EpisodicCount,
        "semantic":   stats.SemanticCount,
        "procedural": stats.ProceduralCount,
        "total":      stats.TotalMemories,
    },
}

c.JSON(http.StatusOK, response)
```

**Response structure**:

```json
{
    "journal": {
        "total_memories": 1250,
        "episodic_count": 800,
        "semantic_count": 400,
        "procedural_count": 50
    },
    "processor": {
        "queue_length": 3,
        "events_processed": 5420,
        "consolidation_count": 85,
        "last_consolidation": 1705745600
    },
    "collections": {
        "episodic": 800,
        "semantic": 400,
        "procedural": 50,
        "total": 1250
    }
}
```

## POST /api/consolidate - Trigger Consolidation

### 1. Request Processing

**Location**: `src/persistent-context-svc/app/server.go:handleTriggerConsolidation()`

```go
type ConsolidateRequest struct {
    Force bool `json:"force"`
}
```

### 2. Event Creation

**Location**: `src/pkg/memory/processor.go:TriggerConsolidation()`

```go
// Create consolidation event
event := Event{
    Type:      EventConversationEnd, // Manual trigger
    Timestamp: time.Now(),
}

// Queue for processing
p.eventQueue <- event
```

### 3. Consolidation Process

**Location**: `src/pkg/memory/processor.go:performConsolidation()`

```go
// 1. Select memories for consolidation
memories := p.selectMemoriesForConsolidation(ctx)

// 2. Check context window constraints
if !p.checkContextWindow(memories) {
    return // Skip if would exceed limits
}

// 3. Generate consolidated memory via LLM
prompt := p.buildConsolidationPrompt(memories)
consolidatedContent, err := p.llm.Consolidate(ctx, prompt)

// 4. Create semantic memory
semanticMemory := &models.Memory{
    ID:      uuid.New().String(),
    Type:    models.MemoryTypeSemantic,
    Content: consolidatedContent,
    SourceMemories: extractIDs(memories),
}

// 5. Store new semantic memory
p.journal.Store(ctx, semanticMemory)
```

## Common Patterns

### Fire-and-Forget Pattern

Used for non-critical operations that shouldn't block the API response:

```go
// In journal.CaptureMemory
go func() {
    // Generate embedding and update in background
    embedding, _ := j.llm.GenerateEmbedding(ctx, memory.Content)
    memory.Embedding = embedding
    j.vectorDB.Update(ctx, memory)
}()

// Return immediately
return memory, nil
```

### Error Propagation

Errors flow up through the layers:

```
Qdrant Error → VectorDB Interface → Journal → HTTP Handler → JSON Response
```

Example:

```go
// In vectorDB
if err != nil {
    return fmt.Errorf("qdrant query failed: %w", err)
}

// In journal
if err != nil {
    return nil, fmt.Errorf("vector search failed: %w", err)
}

// In HTTP handler
if err != nil {
    c.JSON(500, gin.H{"error": err.Error()})
    return
}
```

### Async Processing Queue

Events are processed sequentially in background:

```go
// Single goroutine processes all events
go func() {
    for event := range p.eventQueue {
        // Process one at a time
        p.handleEvent(ctx, event)
    }
}()
```

## Error Handling Flow

### Storage Layer Errors

1. **Qdrant Connection Error**
   - Caught in `qdrantdb.go`
   - Wrapped with context: "qdrant connection failed"
   - Propagated to journal

2. **Embedding Generation Error**
   - Caught in `ollama.go`
   - Logged but doesn't fail the request
   - Memory stored without embedding

3. **Validation Errors**
   - Caught in HTTP handlers
   - Return 400 Bad Request immediately
   - No propagation needed

### Graceful Degradation

- If embedding fails: Store memory without vector (no similarity search)
- If consolidation fails: Log error, continue normal operations
- If queue full: Drop event, log warning

## Key Insights

1. **Asynchronous Processing**: Embeddings are generated after API response
2. **Two-Phase Storage**: Immediate storage + background vector update
3. **Event-Driven**: All major operations trigger events for processing
4. **Graceful Failures**: System continues operating even if components fail
5. **Interface Abstraction**: Each layer only knows about interfaces, not implementations
