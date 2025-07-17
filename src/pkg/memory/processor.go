package memory

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/config"
	"github.com/JaimeStill/persistent-context/pkg/journal"
	"github.com/JaimeStill/persistent-context/pkg/llm"
	"github.com/JaimeStill/persistent-context/pkg/models"
)

// EventType represents different types of memory processing events
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

// ProcessingEvent represents an event that triggers memory processing
type ProcessingEvent struct {
	Type         EventType
	Trigger      string
	Memories     []*models.MemoryEntry
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

// EstimateProcessingCost estimates the token cost for memory processing
func (cm *ContextMonitor) EstimateProcessingCost(memories []*models.MemoryEntry) int {
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

// CanSafelyProcess checks if memory processing can proceed safely
func (cm *ContextMonitor) CanSafelyProcess(memories []*models.MemoryEntry) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	estimatedCost := cm.EstimateProcessingCost(memories)
	safeLimit := int(float64(cm.MaxTokens) * cm.SafetyMargin)

	return cm.CurrentTokens+estimatedCost < safeLimit
}

// GetContextState returns the current context state
func (cm *ContextMonitor) GetContextState(memories []*models.MemoryEntry) ContextState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	estimatedCost := cm.EstimateProcessingCost(memories)
	canProceed := cm.CanSafelyProcess(memories)

	return ContextState{
		WindowSize:    cm.MaxTokens,
		CurrentUsage:  cm.CurrentTokens,
		EstimatedCost: estimatedCost,
		CanProceed:    canProceed,
	}
}

// Use MemoryConfig from config package

// Processor handles event-driven memory processing orchestration
type Processor struct {
	journal    journal.Journal
	llmClient  llm.LLM
	config     *config.MemoryConfig
	monitor    *ContextMonitor
	eventQueue chan ProcessingEvent
	mu         sync.RWMutex
	running    bool
	logger     *slog.Logger
}

// NewProcessor creates a new memory processor
func NewProcessor(journal journal.Journal, llmClient llm.LLM, config *config.MemoryConfig) *Processor {
	return &Processor{
		journal:    journal,
		llmClient:  llmClient,
		config:     config,
		monitor:    NewContextMonitor(config.MaxTokens, config.SafetyMargin),
		eventQueue: make(chan ProcessingEvent, 100),
		logger:     slog.Default(),
	}
}

// Start begins the memory processor
func (p *Processor) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	p.running = true
	go p.processEvents(ctx)

	p.logger.Info("Memory processor started")
	return nil
}

// Stop stops the memory processor
func (p *Processor) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.running = false
	close(p.eventQueue)

	p.logger.Info("Memory processor stopped")
	return nil
}

// processEvents processes consolidation events from the queue
func (p *Processor) processEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-p.eventQueue:
			if !ok {
				return // Channel closed
			}

			if err := p.handleEvent(ctx, event); err != nil {
				p.logger.Error("Failed to handle consolidation event",
					"event_type", event.Type.String(),
					"error", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

// handleEvent handles a specific consolidation event
func (p *Processor) handleEvent(ctx context.Context, event ProcessingEvent) error {
	p.logger.Info("Handling consolidation event",
		"event_type", event.Type.String(),
		"trigger", event.Trigger,
		"memory_count", len(event.Memories))

	switch event.Type {
	case ContextInit:
		return p.OnContextInit(ctx, event)
	case NewContext:
		return p.OnNewContext(ctx, event)
	case ThresholdReached:
		return p.OnThresholdReached(ctx, event)
	case ConversationEnd:
		return p.OnConversationEnd(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %v", event.Type)
	}
}

// OnContextInit handles context initialization events
func (p *Processor) OnContextInit(ctx context.Context, event ProcessingEvent) error {
	p.logger.Info("Processing context initialization consolidation")

	// Get recent memories from previous session
	memories, err := p.journal.GetMemories(ctx, p.config.MemoryCountThreshold)
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}

	if len(memories) == 0 {
		p.logger.Info("No memories to consolidate during context init")
		return nil
	}

	// Check if we can safely consolidate
	if !p.monitor.CanSafelyProcess(memories) {
		p.logger.Warn("Cannot safely consolidate during context init - insufficient context window")
		return nil
	}

	// Select memories for consolidation based on importance
	selectedMemories := p.selectMemoriesForConsolidation(memories)

	// Perform consolidation
	return p.processMemories(ctx, selectedMemories, "context_init")
}

// OnNewContext handles new context events
func (p *Processor) OnNewContext(ctx context.Context, event ProcessingEvent) error {
	p.logger.Info("Processing new context consolidation")

	// Check for consolidation opportunities
	memories, err := p.journal.GetMemories(ctx, p.config.MemoryCountThreshold*2)
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}

	if uint32(len(memories)) < p.config.MemoryCountThreshold {
		p.logger.Debug("Not enough memories for new context consolidation")
		return nil
	}

	// Check context window safety
	if !p.monitor.CanSafelyProcess(memories) {
		p.logger.Warn("Cannot safely consolidate - context window too full")
		return nil
	}

	// Select and consolidate memories
	selectedMemories := p.selectMemoriesForConsolidation(memories)
	return p.processMemories(ctx, selectedMemories, "new_context")
}

// OnThresholdReached handles threshold reached events
func (p *Processor) OnThresholdReached(ctx context.Context, event ProcessingEvent) error {
	p.logger.Info("Processing threshold reached consolidation")

	// Get current context state
	contextState := p.monitor.GetContextState(event.Memories)

	if !contextState.CanProceed {
		p.logger.Warn("Cannot consolidate - context window safety check failed")
		// Consider scheduling early consolidation or cleanup
		return p.scheduleEarlyConsolidation(ctx)
	}

	// Select memories for consolidation
	selectedMemories := p.selectMemoriesForConsolidation(event.Memories)

	return p.processMemories(ctx, selectedMemories, "threshold_reached")
}

// OnConversationEnd handles conversation end events
func (p *Processor) OnConversationEnd(ctx context.Context, event ProcessingEvent) error {
	p.logger.Info("Processing conversation end consolidation")

	// Get all recent memories
	memories, err := p.journal.GetMemories(ctx, p.config.MemoryCountThreshold*3)
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}

	if len(memories) == 0 {
		p.logger.Info("No memories to consolidate at conversation end")
		return nil
	}

	// Final consolidation with more lenient safety checks
	selectedMemories := p.selectMemoriesForConsolidation(memories)

	return p.processMemories(ctx, selectedMemories, "conversation_end")
}

// selectMemoriesForConsolidation selects memories for consolidation based on importance
func (p *Processor) selectMemoriesForConsolidation(memories []*models.MemoryEntry) []*models.MemoryEntry {
	// Score all memories
	scores := make([]struct {
		memory *models.MemoryEntry
		score  models.MemoryScore
	}, len(memories))

	for i, mem := range memories {
		scores[i] = struct {
			memory *models.MemoryEntry
			score  models.MemoryScore
		}{
			memory: mem,
			score:  p.scoreMemory(mem),
		}
	}

	// Sort by composite score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score.CompositeScore > scores[j].score.CompositeScore
	})

	// Select top scoring memories up to threshold
	scoreCount := uint32(len(scores))
	maxMemories := p.config.MemoryCountThreshold
	if scoreCount < maxMemories {
		maxMemories = scoreCount
	}

	selected := make([]*models.MemoryEntry, maxMemories)
	for i := 0; i < int(maxMemories); i++ {
		selected[i] = scores[i].memory
	}

	p.logger.Info("Selected memories for consolidation",
		"selected_count", len(selected),
		"total_available", len(memories))

	return selected
}

// scoreMemory calculates the importance score for a memory
func (p *Processor) scoreMemory(mem *models.MemoryEntry) models.MemoryScore {
	now := time.Now()

	// Calculate access frequency (from metadata if available)
	accessCount := 1
	if count, ok := mem.Metadata["access_count"].(int); ok {
		accessCount = count
	}

	// Calculate time decay
	timeSinceAccess := now.Sub(mem.AccessedAt)
	decayFactor := 1.0 / (1.0 + timeSinceAccess.Hours()*p.config.DecayFactor)

	// Calculate semantic relevance (simplified - could use embedding similarity)
	relevanceScore := float64(mem.Strength) // Use existing strength as proxy

	// Boost score based on association count (more connections = more important)
	associationBoost := 1.0
	if len(mem.AssociationIDs) > 0 {
		// Logarithmic scaling to prevent excessive boost
		associationBoost = 1.0 + math.Log(1+float64(len(mem.AssociationIDs)))*0.2
	}

	// Calculate composite score
	accessScore := float64(accessCount) * p.config.AccessWeight
	weightedRelevanceScore := relevanceScore * p.config.RelevanceWeight
	compositeScore := (accessScore + weightedRelevanceScore) * decayFactor * associationBoost

	return models.MemoryScore{
		BaseImportance:  float64(mem.Strength), // Original importance from memory strength
		AccessFrequency: accessCount,
		LastAccessed:    mem.AccessedAt,
		RelevanceScore:  relevanceScore,
		DecayFactor:     decayFactor,
		CompositeScore:  compositeScore,
	}
}

// processMemories performs the actual consolidation of memories via the journal
func (p *Processor) processMemories(ctx context.Context, memories []*models.MemoryEntry, trigger string) error {
	if len(memories) == 0 {
		return nil
	}

	p.logger.Info("Starting memory consolidation",
		"memory_count", len(memories),
		"trigger", trigger)

	// Perform consolidation using the memory store
	if err := p.journal.ConsolidateMemories(ctx, memories); err != nil {
		return fmt.Errorf("failed to consolidate memories: %w", err)
	}

	// Update access tracking for consolidated memories
	for _, mem := range memories {
		if err := p.updateMemoryAccess(ctx, mem); err != nil {
			p.logger.Warn("Failed to update memory access tracking",
				"memory_id", mem.ID,
				"error", err)
		}
	}

	p.logger.Info("Memory consolidation completed successfully",
		"consolidated_count", len(memories),
		"trigger", trigger)

	return nil
}

// updateMemoryAccess updates access tracking for a memory
func (p *Processor) updateMemoryAccess(ctx context.Context, mem *models.MemoryEntry) error {
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
func (p *Processor) scheduleEarlyConsolidation(ctx context.Context) error {
	p.logger.Info("Scheduling early consolidation due to context window constraints")

	// Get a smaller set of memories for emergency consolidation
	memories, err := p.journal.GetMemories(ctx, p.config.MemoryCountThreshold/2)
	if err != nil {
		return fmt.Errorf("failed to get memories for early consolidation: %w", err)
	}

	if len(memories) == 0 {
		return nil
	}

	// Select only the most important memories
	selectedMemories := p.selectMemoriesForConsolidation(memories)
	maxSelected := p.config.MemoryCountThreshold / 3
	if uint32(len(selectedMemories)) > maxSelected {
		selectedMemories = selectedMemories[:maxSelected]
	}

	return p.processMemories(ctx, selectedMemories, "early_consolidation")
}

// TriggerEvent triggers a consolidation event
func (p *Processor) TriggerEvent(eventType EventType, trigger string, memories []*models.MemoryEntry) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return fmt.Errorf("memory processor not running")
	}

	event := ProcessingEvent{
		Type:         eventType,
		Trigger:      trigger,
		Memories:     memories,
		ContextState: p.monitor.GetContextState(memories),
		Timestamp:    time.Now(),
	}

	select {
	case p.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// UpdateContextUsage updates the context monitor with current usage
func (p *Processor) UpdateContextUsage(tokens int) {
	p.monitor.UpdateUsage(tokens)
}

// GetStats returns memory processor statistics
func (p *Processor) GetStats() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]any{
		"running":           p.running,
		"queue_length":      len(p.eventQueue),
		"current_tokens":    p.monitor.CurrentTokens,
		"max_tokens":        p.monitor.MaxTokens,
		"safety_margin":     p.monitor.SafetyMargin,
		"memory_threshold":  p.config.MemoryCountThreshold,
		"context_threshold": p.config.ContextUsageThreshold,
	}
}
