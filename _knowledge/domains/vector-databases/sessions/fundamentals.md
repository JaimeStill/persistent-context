---
domain: vector-databases
name: fundamentals
title: Vector Database Fundamentals
duration: 45
status: pending
prerequisites: []
builds_on: []
unlocks: [embeddings-and-vectors, similarity-search, memory-types]
complexity: foundational
---

# Vector Database Fundamentals

## Concept Overview

A vector database is a specialized storage system designed for high-dimensional numerical arrays (vectors) that represent data in a way that captures semantic meaning. Unlike traditional databases that store and query exact values, vector databases enable similarity-based search - finding data points that are "similar" rather than identical.

**Core Problems It Solves:**

- Finding semantically similar content without exact keyword matches
- Enabling AI systems to store and retrieve contextual memories
- Supporting machine learning applications that work with embeddings
- Scaling similarity search across millions of high-dimensional vectors

**Why This Matters for Memory Systems:**
In an AI memory system, each piece of information (conversation snippet, learned fact, experience) becomes a vector that captures its meaning. The vector database becomes the "long-term memory storage" where related memories can be found through similarity rather than exact recall.

## Visualization

Think of a vector database like a **specialized library for meaning-based organization**:

**Traditional Database (Library by Category):**

- Books organized by exact categories: Fiction → Mystery → Author Last Name
- You must know the exact category and location to find what you want
- Finding "books similar to Agatha Christie" requires knowing she's in Mystery → British Authors

**Vector Database (Library by Essence):**

- Each book mapped to coordinates in "meaning space" based on themes, writing style, complexity
- Books with similar themes, moods, or styles cluster together naturally
- Ask for "books like Agatha Christie" and discover mystery novels regardless of author, era, or exact genre classification

Each memory in your AI system becomes a point in this "meaning space" - memories about similar topics, contexts, or concepts naturally cluster together.

## Real-World Context

Vector databases power memory-like systems you encounter daily:

- **ChatGPT/Claude**: Finding relevant training examples for responses
- **Google Search**: Understanding search intent beyond exact keywords  
- **Spotify**: "Discover Weekly" recommendations based on listening patterns
- **E-commerce**: "Customers who bought this also liked" suggestions
- **Photo Apps**: Searching "pictures of my dog" without manual tagging

## Prerequisites Check

Before starting, you should be comfortable with:

- [x] Basic understanding that computers store information in databases
- [x] Concept that AI systems can convert text/images into numbers
- [ ] Linear algebra or high-dimensional mathematics (helpful but not required)

## Practical Exercise: Explore Your Vector Database

Let's work with your actual system to understand how vector databases operate in practice.

### Setup

1. **Start Your System**:

   ```bash
   cd /home/jaime/personal/persistent-context
   docker compose up -d
   ```

2. **Build MCP Binary**:

   ```bash
   cd src && go build -o bin/persistent-context-mcp ./cmd/persistent-context-mcp/
   ```

3. **Verify System Health**:

   ```bash
   docker compose logs persistent-context-web | grep -i "server listening"
   ```

### Exercise 1: Create and Explore Vector Embeddings

**Step 1**: Capture memories with different content types to see vector clustering:

```
Use Claude Code MCP tools:
1. capture_memory: "Learning about Go concurrency patterns and channels"
2. capture_memory: "Implementing goroutines for parallel processing" 
3. capture_memory: "Understanding vector databases and embeddings"
4. capture_memory: "Cooking pasta with marinara sauce"
5. capture_memory: "Vector similarity search algorithms"
```

**What's Happening Behind the Scenes**:

- Each memory text is sent to Ollama (phi3:mini model)
- Ollama generates a 3072-dimensional vector representing the semantic meaning
- Similar topics (Go programming, vector databases) will have similar vectors
- Unrelated topics (cooking) will have very different vectors

**Step 2**: Query for related memories to see vector similarity in action:

```
Use search_memories with content: "Go programming"
```

**Expected Results**: Should find memories 1 and 2 (about Go) with high similarity scores (>0.8), but not the cooking memory.

**Step 3**: Test semantic understanding:

```
Use search_memories with content: "parallel programming"
```

**Expected Results**: Should find Go concurrency memories even though they don't contain "parallel programming" exactly.

### Exercise 2: Understand Vector Dimensions and Storage

**Step 1**: Check your system's vector configuration by examining logs:

```bash
docker compose logs qdrant | grep -i dimension
```

**What to Look For**: You should see collection creation with 3072 dimensions (matching phi3:mini model).

**Step 2**: Explore actual vector data through the API:

```bash
# Get stats to see memory count
curl localhost:8543/api/stats

# Get a specific memory to see vector structure (use an ID from search results)
curl localhost:8543/api/memory/{MEMORY_ID}
```

**What You'll See**: The actual 3072-dimensional vector - a long array of float64 values between -1 and 1.

### Exercise 3: Test Similarity Thresholds

**Step 1**: Create memories with varying levels of similarity:

```
1. capture_memory: "Go channels for communication between goroutines"
2. capture_memory: "Goroutine communication using channels"  
3. capture_memory: "Python asyncio for concurrent programming"
4. capture_memory: "JavaScript promises and async/await"
```

**Step 2**: Search and observe similarity scores:

```
search_memories with content: "Go concurrency"
```

**Expected Similarity Pattern**:

- Memories 1-2: Very high similarity (>0.9) - nearly identical concepts
- Memory 3: Moderate similarity (~0.6-0.7) - related concept, different language
- Memory 4: Lower similarity (~0.3-0.5) - related programming concept, different paradigm

### Exercise 4: Debug Vector Issues

**Step 1**: Create a memory without waiting for embedding:

```
capture_memory: "Test memory for vector debugging"
```

**Step 2**: Immediately search for it:

```
search_memories with content: "vector debugging" 
```

**Expected Result**: Memory might not appear in search results initially because embedding generation happens asynchronously.

**Step 3**: Wait and search again:

```bash
# Check logs for embedding completion
docker compose logs persistent-context-web | tail -20

# Search again after embedding is complete
```

**Learning Point**: This demonstrates the asynchronous nature of your vector processing pipeline.

## Understanding Your System's Vector Flow

Based on your codebase, here's what happens during the exercises:

### Text → Vector Transformation

**Location**: `src/pkg/llm/ollama.go:GenerateEmbedding()`

```go
// Your system calls this for each memory
resp, err := o.client.Embeddings(ctx, &api.EmbeddingRequest{
    Model:  "phi3:mini",  // 3072-dimensional vectors
    Prompt: content,
})
```

**Real Data Transformation**:

- Input: `"Learning about Go concurrency patterns"`
- Output: `[]float64{0.123, -0.456, 0.789, ...}` (3072 values)

### Vector → Storage Process

**Location**: `src/pkg/vectordb/qdrantdb.go:Store()`

```go
// Your system stores vectors like this
point := &qdrant.PointStruct{
    Vectors: &qdrant.Vectors{
        VectorsOptions: &qdrant.Vectors_Vector{
            Vector: &qdrant.Vector{Data: memory.Embedding},
        },
    },
    Payload: map[string]*qdrant.Value{
        "content": {Kind: &qdrant.Value_StringValue{StringValue: memory.Content}},
        // ... metadata
    },
}
```

### Similarity Search Process

**Location**: `src/pkg/vectordb/qdrantdb.go:Query()`

```go
// When you search, this happens
searchResult, err := q.client.Query(ctx, &qdrant.QueryPoints{
    Query: &qdrant.QueryInterface{
        Query: &qdrant.QueryInterface_Nearest{
            Nearest: &qdrant.VectorInput{
                VectorInput: &qdrant.VectorInput_Dense{
                    Dense: &qdrant.DenseVector{Data: queryEmbedding},
                },
            },
        },
    },
    Limit: &limit,
})
```

## Vector Database Operations in Go

**Vector Similarity Search** demonstrates Go's interface-based design:

```go
// Clean interface for vector operations
type VectorDB interface {
    Query(ctx context.Context, vector []float64, limit int, threshold float64) ([]*Memory, error)
    Store(ctx context.Context, memory *Memory) error
}

// Usage with explicit error handling
func findSimilarMemories(db VectorDB, queryVector []float64) ([]*Memory, error) {
    memories, err := db.Query(ctx, queryVector, 10, 0.7)
    if err != nil {
        return nil, fmt.Errorf("similarity search failed: %w", err)
    }
    return memories, nil
}
```

**Asynchronous Processing** with Go patterns:

```go
// Go's approach to async processing
func (s *service) CaptureMemory(ctx context.Context, content string) (*Memory, error) {
    memory := &Memory{Content: content}
    
    // Store immediately
    if err := s.store(ctx, memory); err != nil {
        return nil, err
    }
    
    // Generate embedding asynchronously
    go func() {
        if embedding, err := s.llm.GenerateEmbedding(ctx, content); err == nil {
            memory.Embedding = embedding
            s.update(ctx, memory) // Handle error appropriately
        }
    }()
    
    return memory, nil
}
```

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Explain to a colleague**: How is a vector database different from a SQL database in terms of how it finds information?

2. **Identify the use case**: When would you choose vector similarity search over exact keyword search for a memory system?

3. **Debug this scenario**: Your vector database returns seemingly random results when searching for similar memories. What are the most likely causes?

## Common Pitfalls

- **Pitfall 1**: Expecting vectors to be meaningful to humans
  - *Why it happens*: Vectors are mathematical representations, not human-readable
  - *How to avoid*: Focus on the similarity relationships, not individual vector values

- **Pitfall 2**: Using random or poorly-generated vectors for testing
  - *Why it happens*: Similarity only works with properly trained embeddings
  - *How to avoid*: Use real embedding models (like those in your Ollama setup)

- **Pitfall 3**: Ignoring the importance of vector dimensions
  - *Why it happens*: Dimension count affects both accuracy and performance
  - *How to avoid*: Understand your embedding model's output dimensions and optimize accordingly

## Going Deeper

If you want to explore further:

- Investigate how your Ollama embedding model generates vectors from text
- Experiment with different similarity metrics (cosine vs euclidean vs dot product)
- Explore how vector databases handle millions of vectors efficiently (this leads to indexing strategies)

## Notes

<!-- Add your observations as you work through this:
- Which concepts felt clear vs confusing?
- How does this connect to your Go implementation?
- What questions came up about your specific memory system architecture?
-->