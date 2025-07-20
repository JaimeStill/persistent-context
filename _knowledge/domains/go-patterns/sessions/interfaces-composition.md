---
domain: go-patterns
name: interfaces-composition
title: Interface Design and Composition
duration: 30
status: pending
prerequisites: []
builds_on: []
unlocks: [dependency-injection, interface-satisfaction, architecture-patterns]
complexity: foundational
---

# Interface Design and Composition

## Concept Overview

Go's interface system is fundamentally different from most languages - it's based on **implicit satisfaction** and **composition over inheritance**. Understanding this system is crucial because your entire persistent-context architecture is built around small, focused interfaces that compose into larger behaviors.

**Core Problems This Solves:**

- Creating testable, modular code through dependency injection
- Building complex systems from simple, focused components
- Enabling flexible implementations without tight coupling
- Supporting incremental development and easy refactoring

**Why This Matters for Session 14:**
Your consolidation work will involve implementing new interfaces and composing existing ones. Understanding Go's interface philosophy will help you design clean, maintainable code.

## Visualization: Interface Philosophy

**Traditional Inheritance (Other Languages)**:
```
Animal (base class)
  ├─ Mammal extends Animal
      ├─ Dog extends Mammal
      └─ Cat extends Mammal
  └─ Bird extends Animal
      └─ Eagle extends Bird
```

**Go Composition Approach**:
```
Interfaces (behaviors):
- Walker interface { Walk() }
- Swimmer interface { Swim() }
- Flyer interface { Fly() }

Concrete Types:
- Dog: implements Walker
- Duck: implements Walker, Swimmer, Flyer
- Fish: implements Swimmer
```

**Key Insight**: In Go, you define what something *can do* (interfaces) rather than what it *is* (inheritance).

## Prerequisites Check

Before starting, you should understand:

- [x] Basic Go syntax (structs, methods, functions)
- [x] How to define and call methods on structs
- [ ] Interface declarations and implicit satisfaction (we'll cover this)

## Go Interface Fundamentals

### Interface Declaration

**Location**: Throughout your codebase, but great examples in `src/pkg/journal/journal.go`

```go
// Small, focused interface
type Journal interface {
    Store(ctx context.Context, memory *models.Memory) (*models.Memory, error)
    Retrieve(ctx context.Context, id string) (*models.Memory, error)
    Query(ctx context.Context, query string, limit int) ([]*models.Memory, error)
}

// Even smaller interfaces for specific operations
type MemoryStore interface {
    Store(ctx context.Context, memory *models.Memory) (*models.Memory, error)
}

type MemoryRetriever interface {
    Retrieve(ctx context.Context, id string) (*models.Memory, error)
}
```

**Go Philosophy**: "The bigger the interface, the weaker the abstraction"

### Implicit Interface Satisfaction

Unlike other languages, Go types don't explicitly declare they implement an interface:

```go
// Define an interface
type VectorDB interface {
    Store(ctx context.Context, memory *models.Memory) error
    Query(ctx context.Context, vector []float64, limit int) ([]*models.Memory, error)
}

// Any type with these methods automatically satisfies the interface
type qdrant struct {
    client qdrant.QdrantClient
    logger *slog.Logger
}

// This method makes qdrant satisfy VectorDB interface
func (q *qdrant) Store(ctx context.Context, memory *models.Memory) error {
    // Implementation details...
    return nil
}

func (q *qdrant) Query(ctx context.Context, vector []float64, limit int) ([]*models.Memory, error) {
    // Implementation details...
    return nil, nil
}

// No explicit "implements" declaration needed!
```

**Key Benefit**: You can make existing types satisfy new interfaces without modifying their code.

## Interface Composition Patterns

### Pattern 1: Small Interface Combination

**Location**: `src/pkg/memory/processor.go`

Your memory processor composes multiple small interfaces:

```go
type processor struct {
    journal  Journal      // For storing and retrieving memories
    llm      LLM          // For generating embeddings
    vectorDB VectorDB     // For vector operations
    logger   *slog.Logger // For logging
}

// Each interface is focused on one responsibility
type LLM interface {
    GenerateEmbedding(ctx context.Context, content string) ([]float64, error)
    Consolidate(ctx context.Context, prompt string) (string, error)
}

type VectorDB interface {
    Store(ctx context.Context, memory *models.Memory) error
    Query(ctx context.Context, vector []float64, limit int, threshold float64) ([]*models.Memory, error)
}
```

**Benefits**:
- Easy to test (mock individual interfaces)
- Easy to swap implementations
- Clear separation of concerns
- Single responsibility principle

### Pattern 2: Interface Embedding

Go allows embedding interfaces within other interfaces:

```go
// Basic operations
type Reader interface {
    Read(ctx context.Context, id string) (*models.Memory, error)
}

type Writer interface {
    Write(ctx context.Context, memory *models.Memory) error
}

// Composed interface using embedding
type ReadWriter interface {
    Reader  // Embeds all methods from Reader
    Writer  // Embeds all methods from Writer
}

// Usage example
func ProcessMemory(rw ReadWriter, id string) error {
    // Can use both Read and Write methods
    memory, err := rw.Read(ctx, id)
    if err != nil {
        return err
    }
    
    // Modify memory...
    
    return rw.Write(ctx, memory)
}
```

### Pattern 3: Optional Interface Extension

Check if a type supports additional functionality:

```go
// Basic interface
type VectorDB interface {
    Store(ctx context.Context, memory *models.Memory) error
    Query(ctx context.Context, vector []float64, limit int) ([]*models.Memory, error)
}

// Extended interface for advanced features
type AdvancedVectorDB interface {
    VectorDB // Embed the basic interface
    BatchStore(ctx context.Context, memories []*models.Memory) error
    CreateIndex(ctx context.Context, config IndexConfig) error
}

// Usage with type assertion
func OptimizedStore(db VectorDB, memories []*models.Memory) error {
    // Check if the implementation supports batch operations
    if advancedDB, ok := db.(AdvancedVectorDB); ok {
        // Use optimized batch operation
        return advancedDB.BatchStore(ctx, memories)
    }
    
    // Fall back to individual stores
    for _, memory := range memories {
        if err := db.Store(ctx, memory); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Real Project Examples

### Example 1: Journal Interface Design

**Location**: `src/pkg/journal/journal.go`

```go
// The interface defines what a journal can do
type Journal interface {
    // Core memory operations
    Store(ctx context.Context, memory *models.Memory) (*models.Memory, error)
    Retrieve(ctx context.Context, id string) (*models.Memory, error)
    Query(ctx context.Context, query string, limit int) ([]*models.Memory, error)
    
    // Advanced operations
    CaptureMemory(ctx context.Context, memory *models.Memory) (*models.Memory, error)
    GetRecent(ctx context.Context, memoryType models.MemoryType, window time.Duration) ([]*models.Memory, error)
    
    // Association operations
    CreateAssociation(ctx context.Context, association *models.Association) error
    GetAssociations(ctx context.Context, memoryID string) ([]*models.Association, error)
}

// The concrete implementation
type journal struct {
    vectorDB  VectorDB     // Composed interface
    processor Processor    // Composed interface
    logger    *slog.Logger // Concrete type
    config    *config.JournalConfig
}

// Methods that satisfy the Journal interface
func (j *journal) Store(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // Delegate to composed VectorDB interface
    err := j.vectorDB.Store(ctx, memory)
    if err != nil {
        return nil, fmt.Errorf("failed to store memory: %w", err)
    }
    
    return memory, nil
}
```

**Key Patterns**:
- **Interface composition**: journal embeds other interfaces
- **Delegation**: journal methods delegate to composed interfaces
- **Error wrapping**: adds context while preserving error chain

### Example 2: Configurable Interface

**Location**: `src/pkg/config/config.go`

```go
// Interface for configurable components
type Configurable interface {
    Configure(config interface{}) error
    Validate() error
}

// LLM service that supports configuration
type ollama struct {
    client api.Client
    config *LLMConfig
}

// Satisfies Configurable interface
func (o *ollama) Configure(config interface{}) error {
    llmConfig, ok := config.(*LLMConfig)
    if !ok {
        return fmt.Errorf("invalid config type for LLM service")
    }
    
    o.config = llmConfig
    return nil
}

func (o *ollama) Validate() error {
    if o.config == nil {
        return fmt.Errorf("LLM service not configured")
    }
    
    if o.config.ModelName == "" {
        return fmt.Errorf("model name is required")
    }
    
    return nil
}
```

**Benefits**:
- Standardized configuration across components
- Type-safe configuration validation
- Easy to add new configurable components

## Practical Exercise: Interface Design

Let's practice interface design with your actual codebase:

### Setup

Explore your existing interfaces:

```bash
# Find all interface definitions
grep -r "type.*interface" src/pkg/ --include="*.go"
```

### Exercise 1: Analyze Existing Interfaces

**Step 1**: Examine the VectorDB interface:

```bash
# Look at the interface definition
cat src/pkg/vectordb/vectordb.go | grep -A 10 "type.*VectorDB"
```

**Step 2**: Find implementations:

```bash
# Find types that implement VectorDB
grep -r "func.*\*.*) Store" src/pkg/vectordb/ --include="*.go"
```

**Expected Discovery**: You should find the qdrant implementation satisfies VectorDB.

### Exercise 2: Interface Composition Analysis

**Step 1**: Examine the processor struct:

```go
// In src/pkg/memory/processor.go
type processor struct {
    journal    Journal      // What methods does this provide?
    llm        LLM         // What methods does this provide?
    vectorDB   VectorDB    // What methods does this provide?
    eventQueue chan Event  // Go primitive
    logger     *slog.Logger // Concrete type
    config     *config.MemoryConfig
}
```

**Questions to Answer**:
1. How many different interfaces does processor depend on?
2. Which dependencies are interfaces vs concrete types?
3. What would you need to mock to test this struct?

### Exercise 3: Create a New Interface

Design a simple interface for memory scoring:

```go
// Design challenge: Create an interface for memory scoring
type MemoryScorer interface {
    // What methods would you include?
    // Consider: scoring individual memories, scoring for consolidation, etc.
}

// Then design a simple implementation
type basicScorer struct {
    // What fields would you need?
}

// Implement the interface methods
func (b *basicScorer) ScoreMemory(memory *models.Memory) (float64, error) {
    // Basic scoring implementation
    return 0.0, nil
}
```

## Interface Testing Patterns

### Mock Interface for Testing

```go
// Test double that satisfies Journal interface
type mockJournal struct {
    storeFunc    func(ctx context.Context, memory *models.Memory) (*models.Memory, error)
    retrieveFunc func(ctx context.Context, id string) (*models.Memory, error)
    
    // Track calls for verification
    storeCalls    []StoreCall
    retrieveCalls []RetrieveCall
}

type StoreCall struct {
    Memory *models.Memory
    Result *models.Memory
    Error  error
}

func (m *mockJournal) Store(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // Record the call
    call := StoreCall{Memory: memory}
    
    if m.storeFunc != nil {
        call.Result, call.Error = m.storeFunc(ctx, memory)
    }
    
    m.storeCalls = append(m.storeCalls, call)
    return call.Result, call.Error
}

// Usage in tests
func TestProcessorStore(t *testing.T) {
    mock := &mockJournal{
        storeFunc: func(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
            return memory, nil // Success case
        },
    }
    
    processor := NewProcessor(mock, nil, nil) // Inject mock
    
    // Test the processor
    result, err := processor.StoreMemory(ctx, testMemory)
    
    // Verify behavior
    assert.NoError(t, err)
    assert.Len(t, mock.storeCalls, 1)
    assert.Equal(t, testMemory, mock.storeCalls[0].Memory)
}
```

## Go Interface Best Practices

### 1. Accept Interfaces, Return Structs

```go
// Good: Function accepts interface
func ProcessMemories(journal Journal, memories []*models.Memory) error {
    for _, memory := range memories {
        _, err := journal.Store(ctx, memory)
        if err != nil {
            return err
        }
    }
    return nil
}

// Good: Function returns concrete type
func NewInMemoryJournal() *inMemoryJournal {
    return &inMemoryJournal{
        memories: make(map[string]*models.Memory),
    }
}
```

### 2. Keep Interfaces Small

```go
// Good: Single responsibility
type MemoryStore interface {
    Store(ctx context.Context, memory *models.Memory) error
}

// Good: Single responsibility
type MemoryRetriever interface {
    Retrieve(ctx context.Context, id string) (*models.Memory, error)
}

// Compose when needed
type MemoryRepository interface {
    MemoryStore
    MemoryRetriever
}
```

### 3. Define Interfaces Where You Use Them

```go
// Define interface in the package that uses it
package memory

// This package needs a logger
type Logger interface {
    Debug(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}

type Processor struct {
    logger Logger // Use the interface, not *slog.Logger
}
```

## Common Pitfalls and Solutions

### Pitfall 1: Interfaces Too Large

**Problem**:
```go
// Too large - violates single responsibility
type MegaService interface {
    Store(memory *models.Memory) error
    Query(query string) ([]*models.Memory, error)
    GenerateEmbedding(content string) ([]float64, error)
    Consolidate(memories []*models.Memory) (*models.Memory, error)
    SendEmail(to, subject, body string) error
    LogMetrics(metric string, value float64) error
}
```

**Solution**:
```go
// Separate by responsibility
type MemoryStore interface {
    Store(memory *models.Memory) error
    Query(query string) ([]*models.Memory, error)
}

type EmbeddingService interface {
    GenerateEmbedding(content string) ([]float64, error)
}

type ConsolidationService interface {
    Consolidate(memories []*models.Memory) (*models.Memory, error)
}
```

### Pitfall 2: Premature Interface Creation

**Problem**: Creating interfaces before you need them

**Solution**: Start with concrete types, extract interfaces when you have multiple implementations or need testing

### Pitfall 3: Interface Pollution

**Problem**: Every type has an interface

**Solution**: Only create interfaces when you need abstraction (multiple implementations, testing, modularity)

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Interface Satisfaction**: Explain how a Go type satisfies an interface. Why is this different from inheritance?

2. **Composition Benefits**: Why does the processor struct compose multiple interfaces instead of implementing everything itself?

3. **Interface Size**: What makes a good interface in Go? Why are smaller interfaces preferred?

4. **Testing Value**: How do interfaces make testing easier? Provide a specific example from the processor.

## Connection to Session 14

Interface design directly supports your Session 14 work:

- **New Consolidation Interfaces**: You may need to create interfaces for consolidation algorithms
- **Service Composition**: Consolidation services will compose existing interfaces
- **Testing**: Interfaces enable isolated testing of consolidation logic
- **Extensibility**: Interfaces allow adding new consolidation strategies

Understanding composition means you can design clean, testable consolidation features.

## Notes

<!-- Add your observations as you work through this:
- Which interface patterns felt most intuitive vs surprising?
- How does Go's implicit satisfaction compare to explicit interface implementation?
- What questions came up about when to create new interfaces?
- Which examples from your codebase were most helpful?
-->