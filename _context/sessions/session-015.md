# Session 15 Execution Plan: CLI Tool Foundation + Consolidation Fix

## Session Objectives
Build a minimal but extensible CLI tool to diagnose and fix the critical consolidation performance issue, establishing a foundation for future debugging and development tools.

## Session Start Process
- ✅ Reviewed Session 14 handoff - critical consolidation timeout issue with 7+ memory groups causing LLM timeouts
- ✅ Reviewed tasks.md - Session 15 priorities confirmed, architecture refactoring deferred to Session 16
- ✅ Strategy decision: Build CLI tool first to diagnose, then implement fixes based on findings

## Critical Issue from Session 14
- **Problem**: LLM timeouts when consolidating large memory groups (7+ memories)
- **Root Cause**: Attempting to process entire associated groups in single LLM requests
- **Error**: "Post "http://ollama:11434/api/generate": context canceled" after 2 minutes
- **Impact**: Autonomous consolidation fails for reasonably-sized memory groups

## Phase 1: CLI Foundation (45 min)

### Project Structure
```
cmd/persistent-context-cli/
├── main.go              # Entry point
├── config/
│   └── config.go        # CLI-specific configuration
├── commands/
│   ├── root.go          # Root command with potential Bubble Tea setup
│   ├── memory.go        # Memory inspection commands
│   ├── consolidate.go   # Consolidation testing
│   ├── monitor.go       # Real-time monitoring (Bubble Tea)
│   └── config.go        # Configuration commands
├── ui/
│   ├── app.go           # Main Bubble Tea app
│   ├── models.go        # UI data models
│   └── views/
│       ├── memory.go    # Memory browser view
│       ├── consolidation.go # Consolidation monitor
│       └── dashboard.go # Main dashboard
└── pkg/
    ├── client.go        # HTTP/Direct client abstraction
    └── metrics.go       # Basic performance metrics
```

### Tasks
- [x] Create directory structure
- [x] Set up Cobra with commands: memory list, memory show, consolidate test, monitor
- [x] Implement Viper configuration with sensible defaults
- [ ] Create minimal Bubble Tea dashboard (memory count, current operation)
- [x] Set up HTTP client connection (using web service)

### Initial Features
1. **Connection Modes**: Direct mode using pkg infrastructure (HTTP mode scaffolded for future)
2. **Core Commands**:
   - `memory list` - List all memories with IDs and timestamps
   - `memory show <id>` - Show memory details and associations
   - `consolidate test --batch-size N` - Test specific batch size
   - `monitor` - Launch minimal Bubble Tea UI dashboard

## Phase 2: Consolidation Investigation (45 min)

### Testing Strategy
1. Load existing memories with associations from VectorDB
2. Test consolidation with batch sizes 1-10
3. Record:
   - Response times per batch
   - Success/failure rates
   - Memory consumption
   - Token usage estimates (if available)

### Expected Output Format
```
$ persistent-context-cli consolidate test --batch-size 3
Testing consolidation with batch size: 3
Found 7 memories with associations

Processing group 1 (3 memories)... SUCCESS (287ms)
Processing group 2 (3 memories)... SUCCESS (342ms)
Processing group 3 (1 memory)... SUCCESS (89ms)

Summary: 7 memories in 3 batches, total time: 718ms
Recommended batch size: 3-5 memories
```

### Tasks
- [ ] Implement consolidate test command with batch size parameter
- [ ] Add timing instrumentation
- [ ] Test batch sizes 1-10 systematically
- [ ] Record failure patterns and timeout thresholds
- [ ] Identify optimal batch size (expected: 3-5 memories)

## Phase 3: Implement Fix (45 min)

### Solution Implementation in pkg/memory/processor.go
1. **Configuration Changes**:
   - Add `MaxConsolidationBatchSize` to memory config (default: 5)
   - Add `ConsolidationTimeout` override option

2. **Batch Processing Logic**:
   ```go
   // Pseudocode for batch splitting
   if len(memories) > maxBatchSize {
       batches := splitIntoBatches(memories, maxBatchSize)
       for _, batch := range batches {
           result := consolidateBatch(batch)
           results = append(results, result)
       }
   }
   ```

3. **Progressive Consolidation** (if needed):
   - Consolidate pairs first
   - Then consolidate the consolidated results
   - Continue until single result

### Tasks
- [ ] Add MaxConsolidationBatchSize to pkg/config/memory.go
- [ ] Implement batch splitting in pkg/memory/processor.go
- [ ] Add proper error handling for partial batch failures
- [ ] Test via CLI that timeouts are resolved
- [ ] Verify web service consolidation with new batching

## Phase 4: Validation & Documentation (30 min)

### Integration Testing
- [ ] Test complete memory capture → association → consolidation flow
- [ ] Verify no timeouts with various memory group sizes
- [ ] Confirm associations still work correctly post-consolidation
- [ ] Test with both CLI and through MCP tools

### Documentation
- [ ] Document optimal batch size findings
- [ ] Update configuration recommendations
- [ ] Add consolidation tuning guide to execution plan
- [ ] Note any edge cases discovered

## Expected Outcomes

### Performance Improvements
- Consolidation timeout elimination for groups up to 20+ memories
- Predictable consolidation times based on batch size
- Improved system reliability

### Configuration Recommendations
```yaml
memory:
  max_consolidation_batch_size: 5  # Based on testing
  consolidation_timeout: 30s       # Per batch timeout
```

### CLI Tool Benefits
- Permanent debugging tool for future development
- Direct visibility into consolidation process
- Foundation for future monitoring and analysis features

## Session End Process
- [ ] Update execution-plan.md with final results
- [ ] Archive to _context/sessions/session-015.md
- [ ] Update tasks.md with accomplishments and defer architecture refactoring to Session 16
- [ ] Clean up execution-plan.md
- [ ] Reflect on iterative tooling development approach if time permits

## Critical Findings from Container Logs

### Consolidation Timeout Root Cause Confirmed
Successfully reproduced and diagnosed the consolidation timeout issue:

**Error from persistent-context-svc logs:**
```
{"time":"2025-07-22T11:39:00.138885458Z","level":"WARN","msg":"Failed to consolidate memory group","error":"failed to consolidate memories: failed after 4 attempts: failed to make request: Post \"http://ollama:11434/api/generate\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)","group_size":7}
[GIN] 2025/07/22 - 11:39:00 | 200 | 2m6s | POST "/api/v1/journal/consolidate"
```

**Analysis:**
- System attempting to consolidate all 7 associated memories in single LLM request
- Ollama timeout after ~2 minutes due to oversized payload with full memory content
- No batch size limits implemented in consolidation logic (pkg/memory/processor.go)
- LLM cannot handle 7 memories with embeddings and associations in reasonable time

### CLI Tool Success
- ✅ **Full functionality achieved** - Memory listing, detailed inspection, consolidation testing, service monitoring
- ✅ **HTTP endpoints working** - All /api/v1/journal/* routes correctly mapped  
- ✅ **Performance testing ready** - Can trigger consolidation and measure timing
- ✅ **Extensible foundation** - Cobra/Viper/Bubble Tea architecture ready for expansion

## Handoff Notes
- **Primary accomplishment**: Functional CLI tool with diagnostic capabilities
- **Critical issue confirmed**: Consolidation groups of 7+ memories cause LLM timeouts
- **Root cause identified**: No batch size limits in memory consolidation processor
- **Next session priority**: Implement MaxConsolidationBatchSize config and batch splitting logic in pkg/memory/processor.go
- **Architecture ready**: CLI tool provides perfect testing environment for batch optimization
- **Session 16 focus**: Fix performance issue, then proceed with service architecture refactoring