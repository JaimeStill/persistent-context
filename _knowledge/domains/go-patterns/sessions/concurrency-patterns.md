---
domain: go-patterns
name: concurrency-patterns
title: Concurrency with Channels and Goroutines
duration: 30
status: pending
prerequisites: [go-patterns/interfaces-composition]
builds_on: [interfaces-composition]
unlocks: [resource-management, architecture-patterns, async-processing]
complexity: intermediate
---

# Concurrency with Channels and Goroutines

## Concept Overview

Go's concurrency model is built around **goroutines** (lightweight threads) and **channels** (communication primitives). Your persistent-context system uses these extensively for background processing, event handling, and async operations. Understanding these patterns is crucial for Session 14 work.

**Core Problems This Solves:**

- Processing memories asynchronously without blocking API responses
- Coordinating multiple operations safely without shared memory
- Implementing producer-consumer patterns for event processing
- Graceful shutdown and resource management

**Why This Matters for Session 14:**
Your consolidation work will involve event-driven processing, background goroutines, and coordination between multiple components. These patterns are fundamental to the system's architecture.

## Visualization: Concurrency Models

**Traditional Threading (Shared Memory)**:
```
Thread 1 ──┐
            ├─► Shared Memory ──► Race Conditions, Locks
Thread 2 ──┘
```

**Go Concurrency (Message Passing)**:
```
Goroutine 1 ──► Channel ──► Goroutine 2
              (Message)
"Don't communicate by sharing memory; share memory by communicating"
```

**Your System's Event Processing**:
```
HTTP Handler ──► Event Channel ──► Background Processor
(Producer)       (Buffer)          (Consumer)
     │                                 │
     └─ Returns immediately           └─ Processes async
```

## Prerequisites Check

Before starting, ensure you understand:

- [x] Go interfaces and struct composition
- [x] Basic function and method syntax
- [ ] How goroutines relate to OS threads (we'll cover this)
- [ ] Channel operations and buffering (we'll cover this)

## Goroutine Fundamentals

### What Are Goroutines?

Goroutines are lightweight, cooperatively scheduled threads managed by the Go runtime:

```go
// Regular function call (synchronous)
processMemory(memory)

// Goroutine (asynchronous)
go processMemory(memory)

// Multiple goroutines
for i := 0; i < 5; i++ {
    go func(id int) {
        fmt.Printf("Goroutine %d running\n", id)
    }(i)
}
```

**Key Characteristics**:
- **Lightweight**: ~2KB initial stack (vs ~1MB for OS threads)
- **Multiplexed**: Many goroutines run on few OS threads
- **Managed**: Go runtime handles scheduling, not the OS

### Goroutines in Your System

**Location**: `src/pkg/memory/processor.go:Start()`

```go
func (p *processor) Start(ctx context.Context) error {
    // Start background processing goroutine
    go p.processEvents(ctx)
    
    p.logger.Info("Memory processor started")
    return nil
}

func (p *processor) processEvents(ctx context.Context) {
    // Long-running goroutine for event processing
    for {
        select {
        case event := <-p.eventQueue:
            // Process one event
            p.handleEvent(ctx, event)
            
        case <-ctx.Done():
            // Graceful shutdown
            p.logger.Info("Event processing stopped")
            return
        }
    }
}
```

**Pattern**: Long-running background goroutine with graceful shutdown

## Channel Fundamentals

### Channel Types and Operations

```go
// Channel creation
unbuffered := make(chan string)     // Blocks until receiver ready
buffered := make(chan string, 100)  // Buffer of 100 messages

// Sending (blocks if no receiver/buffer full)
ch <- "message"

// Receiving (blocks until message available)
message := <-ch

// Non-blocking operations with select
select {
case ch <- "message":
    // Sent successfully
default:
    // Channel full, do something else
}

select {
case message := <-ch:
    // Received message
default:
    // No message available
}

// Closing channels
close(ch)

// Range over channel (until closed)
for message := range ch {
    // Process each message
}
```

### Channel Patterns in Your System

**Location**: `src/pkg/memory/processor.go`

#### Pattern 1: Event Queue (Producer-Consumer)

```go
type processor struct {
    // Buffered channel for event queue
    eventQueue chan Event
    // Other fields...
}

// Producer: Add events to queue (non-blocking)
func (p *processor) ProcessMemory(ctx context.Context, memory *models.Memory) {
    event := Event{
        Type:      EventNewContext,
        Memory:    memory,
        Timestamp: time.Now(),
    }
    
    // Non-blocking send with overflow protection
    select {
    case p.eventQueue <- event:
        p.logger.Debug("Memory queued for processing", "id", memory.ID)
    default:
        // Queue is full - graceful degradation
        p.logger.Warn("Event queue full, dropping event", "id", memory.ID)
    }
}

// Consumer: Process events from queue
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
```

**Key Benefits**:
- **Decoupling**: Producers don't wait for consumers
- **Buffering**: Handles burst traffic without blocking
- **Overflow Protection**: System degrades gracefully under load

#### Pattern 2: Worker Pool

While not currently in your system, this is a common Go pattern:

```go
// Worker pool for parallel processing
func (p *processor) startWorkerPool(ctx context.Context, numWorkers int) {
    jobs := make(chan *models.Memory, 100)
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        go func(workerID int) {
            for {
                select {
                case memory := <-jobs:
                    // Process memory
                    p.processMemory(ctx, memory)
                    
                case <-ctx.Done():
                    return
                }
            }
        }(i)
    }
    
    // Send jobs to workers
    go func() {
        for memory := range p.inputChannel {
            select {
            case jobs <- memory:
                // Job sent to worker
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

#### Pattern 3: Fan-out, Fan-in

```go
// Fan-out: One input, multiple processors
func (p *processor) fanOut(input <-chan *models.Memory) (<-chan *models.Memory, <-chan *models.Memory) {
    out1 := make(chan *models.Memory)
    out2 := make(chan *models.Memory)
    
    go func() {
        defer close(out1)
        defer close(out2)
        
        for memory := range input {
            // Send to both channels
            select {
            case out1 <- memory:
            case <-time.After(time.Second):
                // Timeout protection
            }
            
            select {
            case out2 <- memory:
            case <-time.After(time.Second):
                // Timeout protection
            }
        }
    }()
    
    return out1, out2
}

// Fan-in: Multiple inputs, one output
func (p *processor) fanIn(input1, input2 <-chan *models.Memory) <-chan *models.Memory {
    output := make(chan *models.Memory)
    
    go func() {
        defer close(output)
        
        for {
            select {
            case memory := <-input1:
                if memory == nil {
                    input1 = nil // Channel closed
                } else {
                    output <- memory
                }
                
            case memory := <-input2:
                if memory == nil {
                    input2 = nil // Channel closed
                } else {
                    output <- memory
                }
                
            default:
                if input1 == nil && input2 == nil {
                    return // Both channels closed
                }
            }
        }
    }()
    
    return output
}
```

## Context and Cancellation

### Context Usage Pattern

**Location**: Throughout your codebase

```go
// Context propagation through call chain
func (s *Server) handleCaptureMemory(c *gin.Context) {
    // HTTP request context
    ctx := c.Request.Context()
    
    // Pass context through all calls
    result, err := s.journal.CaptureMemory(ctx, memory)
    // ...
}

func (j *journal) CaptureMemory(ctx context.Context, memory *models.Memory) (*models.Memory, error) {
    // Pass context to storage layer
    storedMemory, err := j.Store(ctx, memory)
    
    // Pass context to processor
    j.processor.ProcessMemory(ctx, storedMemory)
    
    return storedMemory, nil
}

// Background processing respects context cancellation
func (p *processor) processEvents(ctx context.Context) {
    for {
        select {
        case event := <-p.eventQueue:
            // Pass context to event handler
            p.handleEvent(ctx, event)
            
        case <-ctx.Done():
            // Graceful shutdown when context cancelled
            p.logger.Info("Processing stopped due to context cancellation")
            return
        }
    }
}
```

### Context Types and Usage

```go
import "context"

// Background context (never cancelled)
ctx := context.Background()

// Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel() // Always call cancel to free resources

// Context with deadline
deadline := time.Now().Add(time.Minute)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

// Context with cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Context with values (use sparingly)
ctx = context.WithValue(ctx, "user_id", "12345")
```

## Synchronization Patterns

### WaitGroup for Coordination

```go
import "sync"

// Wait for multiple goroutines to complete
func (p *processor) processMemoriesConcurrently(memories []*models.Memory) error {
    var wg sync.WaitGroup
    errorChan := make(chan error, len(memories))
    
    for _, memory := range memories {
        wg.Add(1)
        
        go func(m *models.Memory) {
            defer wg.Done()
            
            if err := p.processMemory(ctx, m); err != nil {
                errorChan <- err
            }
        }(memory)
    }
    
    // Wait for all goroutines to complete
    wg.Wait()
    close(errorChan)
    
    // Check for errors
    for err := range errorChan {
        if err != nil {
            return err // Return first error
        }
    }
    
    return nil
}
```

### Mutex for Shared State

```go
import "sync"

type processor struct {
    mutex    sync.RWMutex
    stats    ProcessingStats
    // Other fields...
}

type ProcessingStats struct {
    TotalProcessed int64
    ErrorCount     int64
    LastProcessed  time.Time
}

// Thread-safe read
func (p *processor) GetStats() ProcessingStats {
    p.mutex.RLock()         // Read lock
    defer p.mutex.RUnlock() // Always unlock
    
    return p.stats // Safe to return copy
}

// Thread-safe write
func (p *processor) incrementProcessed() {
    p.mutex.Lock()         // Write lock
    defer p.mutex.Unlock()
    
    p.stats.TotalProcessed++
    p.stats.LastProcessed = time.Now()
}
```

## Real-World Examples from Your System

### Example 1: Association Creation Parallelization

**Location**: `src/pkg/memory/processor.go` (enhanced version)

```go
func (p *processor) createAllAssociations(ctx context.Context, memory *models.Memory) error {
    var wg sync.WaitGroup
    errorChan := make(chan error, 4) // Buffer for 4 possible errors
    
    // Create different association types in parallel
    associationTasks := []func() error{
        func() error { return p.journal.CreateTemporalAssociations(ctx, memory) },
        func() error { return p.journal.CreateSemanticAssociations(ctx, memory) },
        func() error { return p.journal.CreateCausalAssociations(ctx, memory) },
        func() error { return p.journal.CreateContextualAssociations(ctx, memory) },
    }
    
    for _, task := range associationTasks {
        wg.Add(1)
        
        go func(taskFunc func() error) {
            defer wg.Done()
            
            if err := taskFunc(); err != nil {
                select {
                case errorChan <- err:
                    // Error recorded
                case <-ctx.Done():
                    // Context cancelled
                }
            }
        }(task)
    }
    
    // Wait for all association creation to complete
    wg.Wait()
    close(errorChan)
    
    // Log errors but don't fail the entire operation
    for err := range errorChan {
        p.logger.Error("Association creation failed", "error", err)
    }
    
    return nil
}
```

### Example 2: Graceful Shutdown

**Location**: `src/persistent-context-svc/main.go` (enhanced)

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start services
    processor := memory.NewProcessor(config)
    err := processor.Start(ctx)
    if err != nil {
        log.Fatal("Failed to start processor:", err)
    }
    
    server := app.NewServer(config)
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Start HTTP server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("HTTP server error: %v", err)
        }
    }()
    
    // Wait for shutdown signal
    <-sigChan
    log.Println("Shutdown signal received")
    
    // Cancel context (stops all background processing)
    cancel()
    
    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Shutdown complete")
}
```

## Practical Exercise: Channel Patterns

### Setup

Let's explore the concurrency patterns in your actual system:

```bash
# Look for goroutine usage
grep -r "go func\|go p\|go s" src/pkg/ --include="*.go"

# Look for channel usage
grep -r "make(chan\|<-\|chan " src/pkg/ --include="*.go"
```

### Exercise 1: Event Queue Analysis

**Step 1**: Examine the event queue:

```bash
# Look at the event queue implementation
grep -A 20 -B 5 "eventQueue" src/pkg/memory/processor.go
```

**Questions to Answer**:
1. What is the buffer size of the event queue?
2. What happens when the queue is full?
3. How does the system ensure events are processed in order?

**Step 2**: Test queue behavior:

```
# Quickly capture multiple memories to test queue behavior
capture_memory: "Test memory 1"
capture_memory: "Test memory 2"
capture_memory: "Test memory 3"
capture_memory: "Test memory 4"
capture_memory: "Test memory 5"
```

**Step 3**: Check processing logs:

```bash
docker compose logs persistent-context-web | grep -i "queue\|processing"
```

### Exercise 2: Context Cancellation

**Step 1**: Start a long-running operation:

```
# Trigger consolidation (long-running operation)
trigger_consolidation
```

**Step 2**: Test graceful shutdown:

```bash
# Send shutdown signal to test graceful shutdown
docker compose stop persistent-context-web
```

**Step 3**: Check shutdown logs:

```bash
docker compose logs persistent-context-web | tail -20
```

**Expected Behavior**: Should see graceful shutdown messages, not abrupt termination.

## Performance Considerations

### Channel Buffer Sizing

```go
// Too small: Frequent blocking
eventQueue := make(chan Event, 10)

// Too large: High memory usage
eventQueue := make(chan Event, 10000)

// Right size: Based on throughput analysis
eventQueue := make(chan Event, 100) // Your system uses this
```

**Rule of Thumb**: Buffer size should handle burst traffic without blocking producers.

### Goroutine Lifecycle

```go
// Good: Bounded goroutines
func (p *processor) Start(ctx context.Context) {
    // Start fixed number of goroutines
    go p.processEvents(ctx)           // 1 processor
    go p.consolidationMonitor(ctx)    // 1 monitor
    go p.metricsReporter(ctx)         // 1 reporter
}

// Problematic: Unbounded goroutines
func (p *processor) ProcessMemory(memory *models.Memory) {
    // DON'T DO THIS - creates goroutine per memory
    go func() {
        p.processMemorySync(memory)
    }()
}
```

### Resource Management

```go
// Always clean up resources
func (p *processor) processWithTimeout(ctx context.Context, memory *models.Memory) error {
    // Create timeout context
    timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel() // Always cancel to free resources
    
    // Use timeout context for operations
    return p.llm.GenerateEmbedding(timeoutCtx, memory.Content)
}
```

## Common Pitfalls and Solutions

### Pitfall 1: Goroutine Leaks

**Problem**:
```go
// Goroutine never exits
go func() {
    for {
        // Process forever, no exit condition
        processData()
    }
}()
```

**Solution**:
```go
// Proper exit condition
go func() {
    for {
        select {
        case data := <-dataChan:
            processData(data)
        case <-ctx.Done():
            return // Exit when context cancelled
        }
    }
}()
```

### Pitfall 2: Channel Deadlocks

**Problem**:
```go
// Deadlock: no receiver
ch := make(chan string)
ch <- "message" // Blocks forever
```

**Solution**:
```go
// Use buffered channel or ensure receiver
ch := make(chan string, 1)
ch <- "message" // Won't block

// Or use select with default
select {
case ch <- "message":
    // Sent successfully
default:
    // Handle full channel
}
```

### Pitfall 3: Race Conditions

**Problem**:
```go
// Race condition on shared variable
var counter int
go func() { counter++ }()
go func() { counter++ }()
```

**Solution**:
```go
// Use channels for communication
counterChan := make(chan int)
go func() { counterChan <- 1 }()
go func() { counterChan <- 1 }()

// Or use sync.Mutex for shared state
var mutex sync.Mutex
var counter int

func increment() {
    mutex.Lock()
    defer mutex.Unlock()
    counter++
}
```

## Comprehension Checkpoint

Answer these questions to validate understanding:

1. **Channel vs Goroutine**: Explain the relationship between channels and goroutines. Why are they used together?

2. **Buffered vs Unbuffered**: When would you choose a buffered channel over an unbuffered channel in the memory processing system?

3. **Context Cancellation**: How does context cancellation provide graceful shutdown in long-running goroutines?

4. **Producer-Consumer**: Explain how the event queue pattern in your system prevents API requests from blocking on slow processing.

## Connection to Session 14

Concurrency patterns directly support your Session 14 work:

- **Consolidation Events**: Adding new event types to the processing queue
- **Parallel Processing**: Using goroutines for concurrent memory analysis
- **Resource Management**: Proper context handling for long-running consolidation
- **System Integration**: Understanding how async processing fits into the overall architecture

Understanding these patterns enables you to implement efficient, responsive consolidation features.

## Notes

<!-- Add your observations as you work through this:
- Which concurrency patterns felt most intuitive vs complex?
- How does Go's channel-based approach compare to async patterns you've used?
- What questions came up about goroutine lifecycle management?
- Which examples from your system were most helpful for understanding?
-->