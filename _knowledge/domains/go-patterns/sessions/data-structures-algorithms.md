---
domain: go-patterns
name: data-structures-algorithms
title: Data Structures and Algorithms
duration: 45
status: pending
prerequisites: [go-patterns/pointers-memory]
builds_on: [pointers-memory]
unlocks: [architecture-patterns, performance-optimization, custom-collections]
complexity: foundational
---

# Data Structures and Algorithms

## Concept Overview

Go provides a powerful set of built-in data structures that form the foundation of efficient algorithms. Your persistent-context system uses these extensively for memory storage, processing pipelines, and association tracking. Understanding their internals, performance characteristics, and idiomatic usage patterns is crucial for writing efficient Go code.

**Core Problems This Solves:**

- Choosing the right data structure for each use case
- Understanding performance implications of different operations
- Implementing efficient algorithms using Go's built-in types
- Designing custom data structures when built-ins aren't sufficient

**Why This Matters for Session 14:**
Your consolidation work involves complex data manipulation, memory scoring algorithms, and association graph traversal. Understanding Go's data structures and algorithmic patterns is essential for implementing efficient consolidation features.

## Built-in Data Structures Overview

### Go's Core Data Structures

```go
// Sequential data structures
var array [5]int                    // Fixed-size array
var slice []int                     // Dynamic array
var string string                   // Immutable byte sequence

// Associative data structures  
var hashMap map[string]int          // Hash table
var channel chan int                // Synchronized queue

// Custom structures
type struct { /* fields */ }        // Product type
type interface { /* methods */ }    // Behavior contract
```

### Performance Characteristics Summary

| Operation | Array/Slice | Map | Channel |
|-----------|-------------|-----|---------|
| Access by index | O(1) | N/A | N/A |
| Access by key | O(n) | O(1) avg | N/A |
| Insert/Append | O(1)* | O(1) avg | O(1)* |
| Delete | O(n) | O(1) avg | O(1) |
| Search | O(n) | O(1) avg | N/A |

*Amortized time complexity

## Arrays and Slices Deep Dive

### Array Fundamentals

```go
// Arrays - fixed size, value types
var arr1 [5]int                     // [0 0 0 0 0]
arr2 := [3]string{"a", "b", "c"}    // [a b c]
arr3 := [...]int{1, 2, 3}           // Compiler counts: [3]int

// Arrays are value types - copying creates new array
original := [3]int{1, 2, 3}
copy := original                    // Entire array is copied
copy[0] = 999
fmt.Println(original)               // [1 2 3] - unchanged
fmt.Println(copy)                   // [999 2 3]

// Array comparison (possible because they're value types)
arr1 := [3]int{1, 2, 3}
arr2 := [3]int{1, 2, 3}
fmt.Println(arr1 == arr2)          // true
```

### Slice Internals and Operations

```go
// Slice structure (conceptually)
type slice struct {
    ptr *ElementType  // Pointer to underlying array
    len int          // Current length
    cap int          // Capacity (underlying array size)
}

// Creating slices
var s1 []int                        // nil slice: len=0, cap=0
s2 := make([]int, 5)               // [0 0 0 0 0]: len=5, cap=5
s3 := make([]int, 3, 10)           // [0 0 0]: len=3, cap=10
s4 := []int{1, 2, 3}               // [1 2 3]: len=3, cap=3

// Slice operations
s := []int{1, 2, 3, 4, 5}
fmt.Printf("len=%d, cap=%d\n", len(s), cap(s))  // len=5, cap=5

// Slicing (creates new slice header, same underlying array)
sub := s[1:4]                      // [2 3 4]: shares array with s
sub[0] = 999                       // Modifies shared array
fmt.Println(s)                     // [1 999 3 4 5] - s was affected!

// Append (may reallocate if capacity exceeded)
s = append(s, 6)                   // Returns new slice (may be different array)
s = append(s, 7, 8, 9)            // Append multiple elements
s = append(s, []int{10, 11}...)    // Append another slice
```

### Memory Management with Slices

```go
// Understanding slice growth
func demonstrateSliceGrowth() {
    var s []int
    
    for i := 0; i < 20; i++ {
        oldCap := cap(s)
        s = append(s, i)
        if cap(s) != oldCap {
            fmt.Printf("Capacity changed from %d to %d\n", oldCap, cap(s))
        }
    }
    // Output shows exponential growth pattern: 0->1->2->4->8->16->32
}

// Efficient slice operations
func efficientSliceUsage() {
    // Pre-allocate if you know the size
    data := make([]int, 0, 1000)    // len=0, cap=1000
    
    for i := 0; i < 1000; i++ {
        data = append(data, i)       // No reallocations needed
    }
    
    // Copy to exact size if needed
    final := make([]int, len(data))
    copy(final, data)               // Now final has cap == len
}
```

### Slice Patterns in Your Codebase

**Location**: `src/pkg/models/models.go`

```go
// Memory struct uses slice for embeddings
type Memory struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    Embedding []float64             `json:"embedding"`  // Large slice (3072 elements)
    // ...
}

// Efficient embedding operations
func ProcessEmbedding(memory *Memory) {
    // Check if embedding exists (nil slice check)
    if memory.Embedding == nil {
        memory.Embedding = make([]float64, 3072)  // Allocate exact size
    }
    
    // Efficient slice operations
    for i := range memory.Embedding {
        memory.Embedding[i] *= 0.5  // Normalize values
    }
}

// Slice aggregation patterns
func CalculateAverageEmbedding(memories []*Memory) []float64 {
    if len(memories) == 0 {
        return nil
    }
    
    // Pre-allocate result slice
    avgEmbedding := make([]float64, len(memories[0].Embedding))
    
    // Accumulate values
    for _, memory := range memories {
        for i, val := range memory.Embedding {
            avgEmbedding[i] += val
        }
    }
    
    // Calculate average
    count := float64(len(memories))
    for i := range avgEmbedding {
        avgEmbedding[i] /= count
    }
    
    return avgEmbedding
}
```

## Maps Deep Dive

### Map Fundamentals

```go
// Map creation
var m1 map[string]int               // nil map (cannot write to)
m2 := make(map[string]int)          // Empty map (can write to)
m3 := map[string]int{               // Map literal
    "alice": 25,
    "bob":   30,
}

// Map operations
m := make(map[string]int)
m["key"] = 42                       // Insert/update
value := m["key"]                   // Read (returns zero value if not found)
value, ok := m["key"]               // Read with existence check
delete(m, "key")                    // Delete
len(m)                              // Number of key-value pairs

// Map iteration (order not guaranteed!)
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

// Keys-only iteration
for key := range m {
    fmt.Printf("Key: %s\n", key)
}
```

### Map Patterns in Your Codebase

**Location**: `src/pkg/models/models.go`

```go
// Memory metadata uses map for flexible key-value storage
type Memory struct {
    // ...
    Metadata map[string]interface{} `json:"metadata"`
}

// Safe map operations
func SetMemoryMetadata(memory *Memory, key string, value interface{}) {
    // Initialize map if nil
    if memory.Metadata == nil {
        memory.Metadata = make(map[string]interface{})
    }
    
    memory.Metadata[key] = value
}

func GetMemoryMetadata(memory *Memory, key string) (interface{}, bool) {
    if memory.Metadata == nil {
        return nil, false
    }
    
    value, exists := memory.Metadata[key]
    return value, exists
}

// Type-safe metadata helpers
func GetStringMetadata(memory *Memory, key string) (string, bool) {
    value, exists := GetMemoryMetadata(memory, key)
    if !exists {
        return "", false
    }
    
    str, ok := value.(string)
    return str, ok
}
```

**Location**: `src/pkg/config/vectordb.go`

```go
// Configuration using maps for flexible collection naming
type VectorDBConfig struct {
    CollectionNames map[string]string `mapstructure:"collection_names"`
}

// Using maps for configuration lookup
func (c *VectorDBConfig) GetCollectionName(memoryType string) string {
    if c.CollectionNames == nil {
        // Default collection names
        defaultNames := map[string]string{
            "episodic":      "episodic_memories",
            "semantic":      "semantic_memories",
            "procedural":    "procedural_memories",
            "metacognitive": "metacognitive_memories",
        }
        c.CollectionNames = defaultNames
    }
    
    if name, exists := c.CollectionNames[memoryType]; exists {
        return name
    }
    
    return memoryType + "_memories"  // Fallback
}
```

## Channels as Data Structures

### Channel Fundamentals

```go
// Channel creation
var ch1 chan int                    // nil channel (blocks forever)
ch2 := make(chan int)               // Unbuffered channel
ch3 := make(chan int, 10)           // Buffered channel (capacity 10)

// Channel operations
ch := make(chan int, 3)
ch <- 1                             // Send (blocks if buffer full)
ch <- 2
ch <- 3

value := <-ch                       // Receive (blocks if buffer empty)
close(ch)                           // Close channel

// Receive with status
value, ok := <-ch                   // ok=false if channel closed and empty

// Range over channel (until closed)
for value := range ch {
    fmt.Println(value)
}
```

### Channel as Queue in Your System

**Location**: `src/pkg/memory/processor.go`

```go
type Event struct {
    Type      EventType     `json:"type"`
    Memory    *Memory       `json:"memory,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
}

type processor struct {
    eventQueue chan Event   // Buffered channel as queue
    // ...
}

// Producer: Add events to queue
func (p *processor) ProcessMemory(ctx context.Context, memory *Memory) {
    event := Event{
        Type:      EventNewContext,
        Memory:    memory,
        Timestamp: time.Now(),
    }
    
    // Non-blocking send (queue pattern)
    select {
    case p.eventQueue <- event:
        p.logger.Debug("Memory queued for processing")
    default:
        p.logger.Warn("Event queue full, dropping event")
    }
}

// Consumer: Process events from queue
func (p *processor) processEvents(ctx context.Context) {
    for {
        select {
        case event := <-p.eventQueue:
            p.handleEvent(ctx, event)   // Process one event at a time
        case <-ctx.Done():
            return                      // Graceful shutdown
        }
    }
}
```

## Algorithm Patterns in Go

### Searching Algorithms

```go
// Linear search for slice
func LinearSearch[T comparable](slice []T, target T) int {
    for i, v := range slice {
        if v == target {
            return i
        }
    }
    return -1
}

// Binary search (requires sorted slice)
func BinarySearch[T comparable](slice []T, target T, compare func(T, T) int) int {
    left, right := 0, len(slice)-1
    
    for left <= right {
        mid := (left + right) / 2
        cmp := compare(slice[mid], target)
        
        if cmp == 0 {
            return mid
        } else if cmp < 0 {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    
    return -1
}

// Using Go's sort package for searching
import "sort"

func SearchMemoriesByTimestamp(memories []*Memory, timestamp int64) int {
    // Assumes memories are sorted by timestamp
    return sort.Search(len(memories), func(i int) bool {
        return memories[i].Timestamp >= timestamp
    })
}
```

### Sorting Algorithms

```go
import "sort"

// Sorting memories by timestamp
func SortMemoriesByTimestamp(memories []*Memory) {
    sort.Slice(memories, func(i, j int) bool {
        return memories[i].Timestamp < memories[j].Timestamp
    })
}

// Sorting by custom criteria (association strength)
func SortAssociationsByStrength(associations []*Association) {
    sort.Slice(associations, func(i, j int) bool {
        return associations[i].Strength > associations[j].Strength  // Descending
    })
}

// Stable sort (maintains relative order of equal elements)
func StableSortMemories(memories []*Memory, compareFunc func(*Memory, *Memory) bool) {
    sort.SliceStable(memories, func(i, j int) bool {
        return compareFunc(memories[i], memories[j])
    })
}

// Custom sort for complex criteria
type MemorySorter struct {
    memories []*Memory
    by       func(m1, m2 *Memory) bool
}

func (ms *MemorySorter) Len() int           { return len(ms.memories) }
func (ms *MemorySorter) Swap(i, j int)      { ms.memories[i], ms.memories[j] = ms.memories[j], ms.memories[i] }
func (ms *MemorySorter) Less(i, j int) bool { return ms.by(ms.memories[i], ms.memories[j]) }

func SortMemoriesBy(memories []*Memory, by func(m1, m2 *Memory) bool) {
    sorter := &MemorySorter{
        memories: memories,
        by:       by,
    }
    sort.Sort(sorter)
}
```

### Graph Algorithms for Associations

**Location**: Association tracking in your system

```go
// Graph representation using adjacency list
type AssociationGraph struct {
    edges map[string][]*Association  // memoryID -> list of associations
}

func NewAssociationGraph() *AssociationGraph {
    return &AssociationGraph{
        edges: make(map[string][]*Association),
    }
}

func (g *AssociationGraph) AddAssociation(assoc *Association) {
    g.edges[assoc.SourceID] = append(g.edges[assoc.SourceID], assoc)
    
    // For undirected graph, add reverse edge
    reverse := &Association{
        ID:       assoc.ID + "_reverse",
        SourceID: assoc.TargetID,
        TargetID: assoc.SourceID,
        Type:     assoc.Type,
        Strength: assoc.Strength,
    }
    g.edges[assoc.TargetID] = append(g.edges[assoc.TargetID], reverse)
}

// Breadth-First Search for related memories
func (g *AssociationGraph) FindRelatedMemories(startID string, maxDepth int) []string {
    if maxDepth <= 0 {
        return nil
    }
    
    visited := make(map[string]bool)
    queue := []struct {
        id    string
        depth int
    }{{startID, 0}}
    
    var result []string
    
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        
        if visited[current.id] || current.depth >= maxDepth {
            continue
        }
        
        visited[current.id] = true
        if current.id != startID {  // Don't include start node
            result = append(result, current.id)
        }
        
        // Add neighbors to queue
        for _, assoc := range g.edges[current.id] {
            if !visited[assoc.TargetID] {
                queue = append(queue, struct {
                    id    string
                    depth int
                }{assoc.TargetID, current.depth + 1})
            }
        }
    }
    
    return result
}

// Depth-First Search with association strength threshold
func (g *AssociationGraph) FindStronglyConnected(startID string, minStrength float64) []string {
    visited := make(map[string]bool)
    var result []string
    
    var dfs func(string)
    dfs = func(nodeID string) {
        if visited[nodeID] {
            return
        }
        
        visited[nodeID] = true
        if nodeID != startID {
            result = append(result, nodeID)
        }
        
        // Visit neighbors with strong connections
        for _, assoc := range g.edges[nodeID] {
            if assoc.Strength >= minStrength && !visited[assoc.TargetID] {
                dfs(assoc.TargetID)
            }
        }
    }
    
    dfs(startID)
    return result
}
```

## Custom Data Structures

### Priority Queue for Memory Scoring

```go
import "container/heap"

// MemoryScore for priority queue
type MemoryScore struct {
    Memory *Memory
    Score  float64
    Index  int  // For heap updates
}

// Priority queue implementation
type MemoryScoreHeap []*MemoryScore

func (h MemoryScoreHeap) Len() int           { return len(h) }
func (h MemoryScoreHeap) Less(i, j int) bool { return h[i].Score > h[j].Score } // Max heap
func (h MemoryScoreHeap) Swap(i, j int) {
    h[i], h[j] = h[j], h[i]
    h[i].Index = i
    h[j].Index = j
}

func (h *MemoryScoreHeap) Push(x interface{}) {
    n := len(*h)
    item := x.(*MemoryScore)
    item.Index = n
    *h = append(*h, item)
}

func (h *MemoryScoreHeap) Pop() interface{} {
    old := *h
    n := len(old)
    item := old[n-1]
    old[n-1] = nil  // Avoid memory leak
    item.Index = -1 // For safety
    *h = old[0 : n-1]
    return item
}

// Usage for top-K memory selection
func SelectTopMemories(memories []*Memory, scoreFunc func(*Memory) float64, k int) []*Memory {
    pq := &MemoryScoreHeap{}
    heap.Init(pq)
    
    // Add all memories to priority queue
    for _, memory := range memories {
        score := scoreFunc(memory)
        heap.Push(pq, &MemoryScore{
            Memory: memory,
            Score:  score,
        })
    }
    
    // Extract top K memories
    result := make([]*Memory, 0, k)
    for i := 0; i < k && pq.Len() > 0; i++ {
        item := heap.Pop(pq).(*MemoryScore)
        result = append(result, item.Memory)
    }
    
    return result
}
```

### LRU Cache for Memory Caching

```go
import "container/list"

type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    list     *list.List
}

type cacheEntry struct {
    key    string
    memory *Memory
}

func NewLRUCache(capacity int) *LRUCache {
    return &LRUCache{
        capacity: capacity,
        cache:    make(map[string]*list.Element),
        list:     list.New(),
    }
}

func (c *LRUCache) Get(key string) (*Memory, bool) {
    if elem, exists := c.cache[key]; exists {
        // Move to front (most recently used)
        c.list.MoveToFront(elem)
        return elem.Value.(*cacheEntry).memory, true
    }
    return nil, false
}

func (c *LRUCache) Put(key string, memory *Memory) {
    if elem, exists := c.cache[key]; exists {
        // Update existing entry
        c.list.MoveToFront(elem)
        elem.Value.(*cacheEntry).memory = memory
        return
    }
    
    // Add new entry
    entry := &cacheEntry{key: key, memory: memory}
    elem := c.list.PushFront(entry)
    c.cache[key] = elem
    
    // Evict least recently used if over capacity
    if c.list.Len() > c.capacity {
        oldest := c.list.Back()
        if oldest != nil {
            c.list.Remove(oldest)
            delete(c.cache, oldest.Value.(*cacheEntry).key)
        }
    }
}
```

## Performance Optimization Patterns

### Benchmarking Data Structure Operations

```go
import "testing"

func BenchmarkSliceAppend(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var slice []int
        for j := 0; j < 1000; j++ {
            slice = append(slice, j)
        }
    }
}

func BenchmarkSlicePrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        slice := make([]int, 0, 1000)  // Pre-allocate capacity
        for j := 0; j < 1000; j++ {
            slice = append(slice, j)
        }
    }
}

func BenchmarkMapAccess(b *testing.B) {
    m := make(map[string]int, 1000)
    for i := 0; i < 1000; i++ {
        m[fmt.Sprintf("key%d", i)] = i
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = m["key500"]
    }
}
```

### Memory-Efficient Patterns

```go
// Pool pattern for reusing expensive objects
import "sync"

var memoryPool = sync.Pool{
    New: func() interface{} {
        return &Memory{
            Embedding: make([]float64, 3072),
            Metadata:  make(map[string]interface{}),
        }
    },
}

func GetPooledMemory() *Memory {
    return memoryPool.Get().(*Memory)
}

func ReturnPooledMemory(memory *Memory) {
    // Reset memory state
    memory.ID = ""
    memory.Content = ""
    memory.Type = ""
    memory.Timestamp = 0
    
    // Clear slice without deallocating
    for i := range memory.Embedding {
        memory.Embedding[i] = 0
    }
    
    // Clear map
    for k := range memory.Metadata {
        delete(memory.Metadata, k)
    }
    
    memoryPool.Put(memory)
}

// String builder for efficient string concatenation
import "strings"

func BuildMemoryDescription(memories []*Memory) string {
    var builder strings.Builder
    builder.Grow(len(memories) * 100)  // Pre-allocate approximate size
    
    for i, memory := range memories {
        if i > 0 {
            builder.WriteString("\n")
        }
        builder.WriteString(fmt.Sprintf("[%s] %s", memory.ID, memory.Content))
    }
    
    return builder.String()
}
```

## Practical Exercise: Algorithm Implementation

### Setup

Let's implement some algorithms using your actual data structures:

### Exercise 1: Memory Similarity Algorithm

```go
// Implement cosine similarity for memory embeddings
func CosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) {
        return 0.0
    }
    
    var dotProduct, normA, normB float64
    
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    
    if normA == 0 || normB == 0 {
        return 0.0
    }
    
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Find most similar memories
func FindSimilarMemories(target *Memory, candidates []*Memory, threshold float64) []*Memory {
    var similar []*Memory
    
    for _, candidate := range candidates {
        if candidate.ID == target.ID {
            continue  // Skip self
        }
        
        similarity := CosineSimilarity(target.Embedding, candidate.Embedding)
        if similarity >= threshold {
            similar = append(similar, candidate)
        }
    }
    
    // Sort by similarity (descending)
    sort.Slice(similar, func(i, j int) bool {
        simI := CosineSimilarity(target.Embedding, similar[i].Embedding)
        simJ := CosineSimilarity(target.Embedding, similar[j].Embedding)
        return simI > simJ
    })
    
    return similar
}
```

### Exercise 2: Memory Clustering Algorithm

```go
// Simple K-means clustering for memories
func ClusterMemories(memories []*Memory, k int, maxIterations int) [][]*Memory {
    if len(memories) <= k {
        // Each memory is its own cluster
        clusters := make([][]*Memory, len(memories))
        for i, memory := range memories {
            clusters[i] = []*Memory{memory}
        }
        return clusters
    }
    
    embeddingSize := len(memories[0].Embedding)
    
    // Initialize centroids randomly
    centroids := make([][]float64, k)
    for i := range centroids {
        centroids[i] = make([]float64, embeddingSize)
        copy(centroids[i], memories[i%len(memories)].Embedding)
    }
    
    clusters := make([][]*Memory, k)
    
    for iteration := 0; iteration < maxIterations; iteration++ {
        // Clear clusters
        for i := range clusters {
            clusters[i] = clusters[i][:0]
        }
        
        // Assign memories to closest centroid
        for _, memory := range memories {
            bestCluster := 0
            bestDistance := math.Inf(1)
            
            for i, centroid := range centroids {
                distance := euclideanDistance(memory.Embedding, centroid)
                if distance < bestDistance {
                    bestDistance = distance
                    bestCluster = i
                }
            }
            
            clusters[bestCluster] = append(clusters[bestCluster], memory)
        }
        
        // Update centroids
        converged := true
        for i, cluster := range clusters {
            if len(cluster) == 0 {
                continue
            }
            
            newCentroid := calculateCentroid(cluster)
            if !slicesEqual(centroids[i], newCentroid) {
                converged = false
                centroids[i] = newCentroid
            }
        }
        
        if converged {
            break
        }
    }
    
    return clusters
}

func euclideanDistance(a, b []float64) float64 {
    var sum float64
    for i := range a {
        diff := a[i] - b[i]
        sum += diff * diff
    }
    return math.Sqrt(sum)
}

func calculateCentroid(memories []*Memory) []float64 {
    if len(memories) == 0 {
        return nil
    }
    
    embeddingSize := len(memories[0].Embedding)
    centroid := make([]float64, embeddingSize)
    
    for _, memory := range memories {
        for i, val := range memory.Embedding {
            centroid[i] += val
        }
    }
    
    count := float64(len(memories))
    for i := range centroid {
        centroid[i] /= count
    }
    
    return centroid
}
```

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Slice vs Array**: When would you choose a slice over an array? How does this affect memory usage in your memory processing system?

2. **Map Performance**: Why are maps O(1) average case but not O(1) worst case? When might this matter in your association tracking?

3. **Channel Buffering**: How does the buffer size of your event queue affect system performance and memory usage?

4. **Algorithm Choice**: For finding the top-K most similar memories, would you use sorting or a priority queue? Why?

## Connection to Session 14

Data structure knowledge directly supports Session 14 work:

- **Memory Consolidation**: Efficiently grouping and processing large sets of memories
- **Association Analysis**: Graph algorithms for finding related memories
- **Performance Optimization**: Choosing appropriate data structures for intensive operations
- **Scoring Algorithms**: Priority queues and sorting for memory ranking

Understanding these patterns enables you to implement efficient, scalable consolidation features.

## Notes

<!-- Add your observations as you work through this:
- Which data structures felt most natural for memory operations vs surprising?
- How do Go's built-in performance characteristics compare to other languages?
- What questions came up about algorithm complexity in your use cases?
- Which custom data structure patterns were most useful?
-->