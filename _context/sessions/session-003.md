# Session 3 Execution Plan: Event-Driven Memory Consolidation

## Overview

Session 3 focuses on implementing event-driven memory consolidation with context window awareness. Based on user feedback, we're moving away from time-based cycles to event-based consolidation that makes sense for LLM context processing.

## Key Design Changes

- **No Time-Based Cycles**: Consolidation driven by LLM lifecycle events, not time
- **Context Window Awareness**: Safety checks to prevent consolidation when context is near capacity
- **Event-Driven Architecture**: Consolidation triggered by meaningful events in the LLM's experience

## Session Progress

### 0. Documentation Setup (5 minutes) - COMPLETED

- [x] Archive Session 2 execution plan to `_context/sessions/session-002.md`
- [x] Create new execution plan for Session 3
- [ ] Update at end of session with results

### 1. Wire Up Components (20 minutes) - PENDING

- [ ] Update main.go to properly initialize all components
- [ ] Fix Qdrant API compatibility issues (Search vs Query methods)
- [ ] Connect MCP server to memory store
- [ ] Test basic end-to-end memory capture flow
- [ ] Verify health checks work across all services

### 2. Implement Event-Driven Consolidation (25 minutes) - PENDING

- [ ] Create `internal/consolidation/` package structure
- [ ] Implement `engine.go` with context-aware event handlers:
  - `OnContextInit()` - Consolidate previous session memories
  - `OnNewContext()` - Check for consolidation opportunities
  - `OnThresholdReached()` - Trigger immediate consolidation with context checks
  - `OnConversationEnd()` - Final consolidation and cleanup
- [ ] Add context window monitoring:

  ```go
  type ContextMonitor struct {
      MaxTokens          int
      CurrentTokens      int
      ConsolidationCost  int
      SafetyMargin       float64
  }
  ```

- [ ] Implement configurable thresholds:
  - Memory count threshold
  - Total embedding size threshold
  - Context window usage threshold (e.g., 70%)
  - Minimum remaining tokens for consolidation
- [ ] Create background consolidation worker with event queue

### 3. Add Memory Lifecycle Management (10 minutes) - PENDING

- [ ] Implement importance scoring algorithm:
  - Access frequency tracking
  - Semantic relevance scoring
  - Recency and decay factors
- [ ] Create memory pruning for low-importance memories
- [ ] Add pre-consolidation memory selection to minimize token usage
- [ ] Implement metadata tracking for consolidation history

### 4. Testing & Documentation (5 minutes) - PENDING

- [ ] Test event-driven consolidation triggers
- [ ] Test context window safety mechanisms
- [ ] Document consolidation behavior
- [ ] Update execution plan with results
- [ ] Create handoff notes for next session

## Implementation Details

### Event System Architecture

```go
// Consolidation events with context awareness
type ConsolidationEvent struct {
    Type             EventType
    Trigger          string
    Memories         []Memory
    ContextState     ContextState
    Timestamp        time.Time
}

type EventType int
const (
    ContextInit EventType = iota
    NewContext
    ThresholdReached
    ConversationEnd
)

type ContextState struct {
    WindowSize       int
    CurrentUsage     int
    EstimatedCost    int
    CanProceed       bool
}
```

### Context-Aware Consolidation

```go
func (c *ConsolidationEngine) OnThresholdReached(memories []Memory) error {
    // Check if we have enough context window remaining
    remainingTokens := c.monitor.MaxTokens - c.monitor.CurrentTokens
    requiredTokens := c.estimateConsolidationTokens(memories)
    
    if remainingTokens < requiredTokens * 1.5 { // 50% safety buffer
        // Defer consolidation or trigger early termination
        return c.scheduleEarlyConsolidation()
    }
    
    return c.consolidate(memories)
}

// Pre-consolidation safety check
func (c *ConsolidationEngine) canSafelyConsolidate() bool {
    state := c.getContextState()
    return state.CurrentUsage + state.EstimatedCost < 
           int(float64(state.WindowSize) * c.config.SafetyMargin)
}
```

### Memory Importance Scoring

```go
type MemoryScore struct {
    AccessCount      int
    LastAccessed     time.Time
    SemanticRelevance float32
    DecayFactor      float32
    TotalScore       float32
}

func (c *ConsolidationEngine) scoreMemory(m *Memory) MemoryScore {
    // Implementation will consider multiple factors
}
```

## Success Criteria

- All components integrated and working together
- Event-driven consolidation responding to LLM lifecycle events
- Context window safety preventing consolidation overflow
- Memory lifecycle managed based on importance scoring
- System ready for continuous operation
- Clear documentation for handoff

## Notes

- Focus on events that make sense for LLM context processing
- Ensure consolidation never causes context window overflow
- Design for extensibility to add more event types later
- Keep implementation simple but robust

## Session 3 Results

### Major Accomplishments

1. **Service Registration Architecture**: Implemented complete service registry and lifecycle management system in `app/` package:
   - Base service interface with initialization, startup, shutdown, health checks
   - Service registry with dependency resolution and startup ordering
   - Lifecycle manager for graceful shutdown handling

2. **Interface Abstraction Pattern**: Created proper abstractions for all infrastructure:
   - `VectorDB` interface (internal/vectordb/vectordb.go) with Qdrant implementation
   - `LLM` interface (internal/llm/llm.go) with Ollama implementation
   - Service wrappers for vectordb, llm, memory, http, mcp

3. **Package Reorganization**:

   - Merged `internal/storage` into `internal/memory` for better cohesion
   - Created `internal/types` package to resolve import cycles
   - Moved all memory types to shared location

4. **Configuration Improvements**:
   - Updated memory config to use `uint64` for BatchSize/MaxMemorySize to match VectorDB requirements
   - Fixed HTTP server to accept HTTPConfig instead of full Config

5. **API Compatibility Fixes**:
   - Updated Qdrant client to use modern Go client API (Query vs Search)
   - Fixed method signatures and response handling

### Current Status

**Completed:**

- Service registry and lifecycle infrastructure
- All service wrappers with proper interfaces
- Package reorganization
- Basic API compatibility fixes

**In Progress:**

- Import cycle resolution (types package created, vectordb updated)
- Memory store needs to import types instead of defining locally

**Pending for Session 4:**

- Complete import cycle fixes (update memory/store.go and qdrantdb.go imports)
- Build pipeline infrastructure and memory pipeline
- Create app orchestrator and update main.go
- Implement consolidation engine with event handlers
- Add context window safety and lifecycle management

### Architecture Achievements

The codebase now has a solid service-oriented architecture with:

- Clean separation between core business logic (`internal/`) and orchestration (`app/`)
- Interface-based design allowing easy swapping of implementations
- Proper dependency injection and lifecycle management
- Foundation ready for pipeline middleware and consolidation engine

### Next Session Priority

1. **Fix remaining import issues** - Complete the types package migration
2. **Build app orchestrator** - Wire up all services in main.go
3. **Implement memory pipeline** - Create middleware infrastructure
4. **Add consolidation engine** - Event-driven consolidation system

The foundation is now solid for implementing the event-driven consolidation system discussed with the user.
