package consolidation

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/journal"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// EventType represents different types of consolidation events
type EventType int

const (
	ContextInit EventType = iota
	NewContext
	ThresholdReached
	ConversationEnd
)

// String returns the string representation of the event type
func (e EventType) String() string {
	switch e {
	case ContextInit:
		return "ContextInit"
	case NewContext:
		return "NewContext"
	case ThresholdReached:
		return "ThresholdReached"
	case ConversationEnd:
		return "ConversationEnd"
	default:
		return "Unknown"
	}
}

// ConsolidationEvent represents an event that triggers consolidation
type ConsolidationEvent struct {
	Type         EventType
	Trigger      string
	Memories     []*types.MemoryEntry
	ContextState ContextState
	Timestamp    time.Time
}

// ContextState represents the current state of the context window
type ContextState struct {
	WindowSize    int
	CurrentUsage  int
	EstimatedCost int
	CanProceed    bool
}

// MemoryScore represents the importance score of a memory
type MemoryScore struct {
	AccessCount       int
	LastAccessed      time.Time
	SemanticRelevance float32
	DecayFactor       float32
	TotalScore        float32
}

// ContextMonitor tracks context window usage and safety
type ContextMonitor struct {
	MaxTokens         int
	CurrentTokens     int
	ConsolidationCost int
	SafetyMargin      float64
	mu                sync.RWMutex
}

// NewContextMonitor creates a new context monitor
func NewContextMonitor(maxTokens int, safetyMargin float64) *ContextMonitor {
	return &ContextMonitor{
		MaxTokens:    maxTokens,
		SafetyMargin: safetyMargin,
	}
}

// UpdateUsage updates the current token usage
func (cm *ContextMonitor) UpdateUsage(tokens int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.CurrentTokens = tokens
}

// EstimateConsolidationCost estimates the token cost for consolidation
func (cm *ContextMonitor) EstimateConsolidationCost(memories []*types.MemoryEntry) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	// Simple estimation: content length + overhead
	totalContent := 0
	for _, mem := range memories {
		totalContent += len(mem.Content)
	}
	
	// Add overhead for prompts and processing
	overhead := 1000
	return totalContent + overhead
}

// CanSafelyConsolidate checks if consolidation can proceed safely
func (cm *ContextMonitor) CanSafelyConsolidate(memories []*types.MemoryEntry) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	estimatedCost := cm.EstimateConsolidationCost(memories)
	safeLimit := int(float64(cm.MaxTokens) * cm.SafetyMargin)
	
	return cm.CurrentTokens+estimatedCost < safeLimit
}

// GetContextState returns the current context state
func (cm *ContextMonitor) GetContextState(memories []*types.MemoryEntry) ContextState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	estimatedCost := cm.EstimateConsolidationCost(memories)
	canProceed := cm.CanSafelyConsolidate(memories)
	
	return ContextState{
		WindowSize:    cm.MaxTokens,
		CurrentUsage:  cm.CurrentTokens,
		EstimatedCost: estimatedCost,
		CanProceed:    canProceed,
	}
}

// Use ConsolidationConfig from config package

// Engine handles event-driven memory consolidation
type Engine struct {
	journal journal.Journal
	llmClient   llm.LLM
	config      *config.ConsolidationConfig
	monitor     *ContextMonitor
	eventQueue  chan ConsolidationEvent
	mu          sync.RWMutex
	running     bool
	logger      *slog.Logger
}

// NewEngine creates a new consolidation engine
func NewEngine(journal journal.Journal, llmClient llm.LLM, config *config.ConsolidationConfig) *Engine {
	return &Engine{
		journal: journal,
		llmClient:   llmClient,
		config:      config,
		monitor:     NewContextMonitor(config.MaxTokens, config.SafetyMargin),
		eventQueue:  make(chan ConsolidationEvent, 100),
		logger:      slog.Default(),
	}
}

// Start begins the consolidation engine
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		return nil
	}
	
	e.running = true
	go e.processEvents(ctx)
	
	e.logger.Info("Consolidation engine started")
	return nil
}

// Stop stops the consolidation engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return nil
	}
	
	e.running = false
	close(e.eventQueue)
	
	e.logger.Info("Consolidation engine stopped")
	return nil
}

// processEvents processes consolidation events from the queue
func (e *Engine) processEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-e.eventQueue:
			if !ok {
				return // Channel closed
			}
			
			if err := e.handleEvent(ctx, event); err != nil {
				e.logger.Error("Failed to handle consolidation event",
					"event_type", event.Type.String(),
					"error", err)
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// handleEvent handles a specific consolidation event
func (e *Engine) handleEvent(ctx context.Context, event ConsolidationEvent) error {
	e.logger.Info("Handling consolidation event",
		"event_type", event.Type.String(),
		"trigger", event.Trigger,
		"memory_count", len(event.Memories))
	
	switch event.Type {
	case ContextInit:
		return e.OnContextInit(ctx, event)
	case NewContext:
		return e.OnNewContext(ctx, event)
	case ThresholdReached:
		return e.OnThresholdReached(ctx, event)
	case ConversationEnd:
		return e.OnConversationEnd(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %v", event.Type)
	}
}

// OnContextInit handles context initialization events
func (e *Engine) OnContextInit(ctx context.Context, event ConsolidationEvent) error {
	e.logger.Info("Processing context initialization consolidation")
	
	// Get recent memories from previous session
	memories, err := e.journal.GetMemories(ctx, uint64(e.config.MemoryCountThreshold))
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}
	
	if len(memories) == 0 {
		e.logger.Info("No memories to consolidate during context init")
		return nil
	}
	
	// Check if we can safely consolidate
	if !e.monitor.CanSafelyConsolidate(memories) {
		e.logger.Warn("Cannot safely consolidate during context init - insufficient context window")
		return nil
	}
	
	// Select memories for consolidation based on importance
	selectedMemories := e.selectMemoriesForConsolidation(memories)
	
	// Perform consolidation
	return e.consolidateMemories(ctx, selectedMemories, "context_init")
}

// OnNewContext handles new context events
func (e *Engine) OnNewContext(ctx context.Context, event ConsolidationEvent) error {
	e.logger.Info("Processing new context consolidation")
	
	// Check for consolidation opportunities
	memories, err := e.journal.GetMemories(ctx, uint64(e.config.MemoryCountThreshold*2))
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}
	
	if len(memories) < e.config.MemoryCountThreshold {
		e.logger.Debug("Not enough memories for new context consolidation")
		return nil
	}
	
	// Check context window safety
	if !e.monitor.CanSafelyConsolidate(memories) {
		e.logger.Warn("Cannot safely consolidate - context window too full")
		return nil
	}
	
	// Select and consolidate memories
	selectedMemories := e.selectMemoriesForConsolidation(memories)
	return e.consolidateMemories(ctx, selectedMemories, "new_context")
}

// OnThresholdReached handles threshold reached events
func (e *Engine) OnThresholdReached(ctx context.Context, event ConsolidationEvent) error {
	e.logger.Info("Processing threshold reached consolidation")
	
	// Get current context state
	contextState := e.monitor.GetContextState(event.Memories)
	
	if !contextState.CanProceed {
		e.logger.Warn("Cannot consolidate - context window safety check failed")
		// Consider scheduling early consolidation or cleanup
		return e.scheduleEarlyConsolidation(ctx)
	}
	
	// Select memories for consolidation
	selectedMemories := e.selectMemoriesForConsolidation(event.Memories)
	
	return e.consolidateMemories(ctx, selectedMemories, "threshold_reached")
}

// OnConversationEnd handles conversation end events
func (e *Engine) OnConversationEnd(ctx context.Context, event ConsolidationEvent) error {
	e.logger.Info("Processing conversation end consolidation")
	
	// Get all recent memories
	memories, err := e.journal.GetMemories(ctx, uint64(e.config.MemoryCountThreshold*3))
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}
	
	if len(memories) == 0 {
		e.logger.Info("No memories to consolidate at conversation end")
		return nil
	}
	
	// Final consolidation with more lenient safety checks
	selectedMemories := e.selectMemoriesForConsolidation(memories)
	
	return e.consolidateMemories(ctx, selectedMemories, "conversation_end")
}

// selectMemoriesForConsolidation selects memories for consolidation based on importance
func (e *Engine) selectMemoriesForConsolidation(memories []*types.MemoryEntry) []*types.MemoryEntry {
	// Score all memories
	scores := make([]struct {
		memory *types.MemoryEntry
		score  MemoryScore
	}, len(memories))
	
	for i, mem := range memories {
		scores[i] = struct {
			memory *types.MemoryEntry
			score  MemoryScore
		}{
			memory: mem,
			score:  e.scoreMemory(mem),
		}
	}
	
	// Sort by total score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score.TotalScore > scores[j].score.TotalScore
	})
	
	// Select top scoring memories up to threshold
	maxMemories := e.config.MemoryCountThreshold
	if len(scores) < maxMemories {
		maxMemories = len(scores)
	}
	
	selected := make([]*types.MemoryEntry, maxMemories)
	for i := 0; i < maxMemories; i++ {
		selected[i] = scores[i].memory
	}
	
	e.logger.Info("Selected memories for consolidation",
		"selected_count", len(selected),
		"total_available", len(memories))
	
	return selected
}

// scoreMemory calculates the importance score for a memory
func (e *Engine) scoreMemory(mem *types.MemoryEntry) MemoryScore {
	now := time.Now()
	
	// Calculate access frequency (from metadata if available)
	accessCount := 1
	if count, ok := mem.Metadata["access_count"].(int); ok {
		accessCount = count
	}
	
	// Calculate time decay
	timeSinceAccess := now.Sub(mem.AccessedAt)
	decayFactor := float32(1.0 / (1.0 + timeSinceAccess.Hours()*e.config.DecayFactor))
	
	// Calculate semantic relevance (simplified - could use embedding similarity)
	semanticRelevance := mem.Strength // Use existing strength as proxy
	
	// Calculate total score
	accessScore := float32(accessCount) * float32(e.config.AccessWeight)
	relevanceScore := semanticRelevance * float32(e.config.RelevanceWeight)
	totalScore := (accessScore + relevanceScore) * decayFactor
	
	return MemoryScore{
		AccessCount:       accessCount,
		LastAccessed:      mem.AccessedAt,
		SemanticRelevance: semanticRelevance,
		DecayFactor:       decayFactor,
		TotalScore:        totalScore,
	}
}

// consolidateMemories performs the actual consolidation of memories
func (e *Engine) consolidateMemories(ctx context.Context, memories []*types.MemoryEntry, trigger string) error {
	if len(memories) == 0 {
		return nil
	}
	
	e.logger.Info("Starting memory consolidation",
		"memory_count", len(memories),
		"trigger", trigger)
	
	// Perform consolidation using the memory store
	if err := e.journal.ConsolidateMemories(ctx, memories); err != nil {
		return fmt.Errorf("failed to consolidate memories: %w", err)
	}
	
	// Update access tracking for consolidated memories
	for _, mem := range memories {
		if err := e.updateMemoryAccess(ctx, mem); err != nil {
			e.logger.Warn("Failed to update memory access tracking",
				"memory_id", mem.ID,
				"error", err)
		}
	}
	
	e.logger.Info("Memory consolidation completed successfully",
		"consolidated_count", len(memories),
		"trigger", trigger)
	
	return nil
}

// updateMemoryAccess updates access tracking for a memory
func (e *Engine) updateMemoryAccess(ctx context.Context, mem *types.MemoryEntry) error {
	// Update access count and timestamp
	if mem.Metadata == nil {
		mem.Metadata = make(map[string]any)
	}
	
	accessCount := 1
	if count, ok := mem.Metadata["access_count"].(int); ok {
		accessCount = count + 1
	}
	
	mem.Metadata["access_count"] = accessCount
	mem.Metadata["last_consolidation"] = time.Now().Unix()
	mem.AccessedAt = time.Now()
	
	return nil
}

// scheduleEarlyConsolidation schedules early consolidation when context window is full
func (e *Engine) scheduleEarlyConsolidation(ctx context.Context) error {
	e.logger.Info("Scheduling early consolidation due to context window constraints")
	
	// Get a smaller set of memories for emergency consolidation
	memories, err := e.journal.GetMemories(ctx, uint64(e.config.MemoryCountThreshold/2))
	if err != nil {
		return fmt.Errorf("failed to get memories for early consolidation: %w", err)
	}
	
	if len(memories) == 0 {
		return nil
	}
	
	// Select only the most important memories
	selectedMemories := e.selectMemoriesForConsolidation(memories)
	if len(selectedMemories) > e.config.MemoryCountThreshold/3 {
		selectedMemories = selectedMemories[:e.config.MemoryCountThreshold/3]
	}
	
	return e.consolidateMemories(ctx, selectedMemories, "early_consolidation")
}

// TriggerEvent triggers a consolidation event
func (e *Engine) TriggerEvent(eventType EventType, trigger string, memories []*types.MemoryEntry) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.running {
		return fmt.Errorf("consolidation engine not running")
	}
	
	event := ConsolidationEvent{
		Type:         eventType,
		Trigger:      trigger,
		Memories:     memories,
		ContextState: e.monitor.GetContextState(memories),
		Timestamp:    time.Now(),
	}
	
	select {
	case e.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// UpdateContextUsage updates the context monitor with current usage
func (e *Engine) UpdateContextUsage(tokens int) {
	e.monitor.UpdateUsage(tokens)
}

// GetStats returns consolidation engine statistics
func (e *Engine) GetStats() map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return map[string]any{
		"running":            e.running,
		"queue_length":       len(e.eventQueue),
		"current_tokens":     e.monitor.CurrentTokens,
		"max_tokens":         e.monitor.MaxTokens,
		"safety_margin":      e.monitor.SafetyMargin,
		"memory_threshold":   e.config.MemoryCountThreshold,
		"context_threshold":  e.config.ContextUsageThreshold,
	}
}