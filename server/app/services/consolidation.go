package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/consolidation"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/memory"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// ConsolidationService wraps the consolidation engine as a managed service
type ConsolidationService struct {
	BaseService
	engine *consolidation.Engine
	config *config.ConsolidationConfig
	logger *slog.Logger
}

// NewConsolidationService creates a new consolidation service
func NewConsolidationService(cfg *config.ConsolidationConfig, logger *slog.Logger) *ConsolidationService {
	return &ConsolidationService{
		BaseService: NewBaseService("consolidation", "memory", "llm"), // Depends on memory and llm
		config:      cfg,
		logger:      logger,
	}
}

// Initialize prepares the consolidation service
func (s *ConsolidationService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// Dependencies will be injected by the orchestrator
	s.SetInitialized(true)
	return nil
}

// InitializeWithDependencies initializes the consolidation service with injected dependencies
func (s *ConsolidationService) InitializeWithDependencies(memoryStore *memory.MemoryStore, llmClient llm.LLM) error {
	if s.IsInitialized() {
		return nil
	}

	// Create consolidation engine
	s.engine = consolidation.NewEngine(memoryStore, llmClient, s.config)
	
	s.SetInitialized(true)
	s.logger.Info("Consolidation service initialized",
		"max_tokens", s.config.MaxTokens,
		"safety_margin", s.config.SafetyMargin,
		"enabled", s.config.Enabled,
	)
	return nil
}

// Start begins consolidation operations
func (s *ConsolidationService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("consolidation service not initialized")
	}

	if !s.config.Enabled {
		s.logger.Info("Consolidation service disabled by configuration")
		s.SetRunning(true)
		return nil
	}

	// Start the consolidation engine
	if err := s.engine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start consolidation engine: %w", err)
	}

	s.SetRunning(true)
	s.logger.Info("Consolidation service started")
	return nil
}

// Stop gracefully shuts down the consolidation service
func (s *ConsolidationService) Stop(ctx context.Context) error {
	if !s.IsRunning() {
		return nil
	}

	if s.engine != nil {
		if err := s.engine.Stop(); err != nil {
			s.logger.Error("Error stopping consolidation engine", "error", err)
			return err
		}
	}

	s.SetRunning(false)
	s.logger.Info("Consolidation service stopped")
	return nil
}

// HealthCheck verifies the consolidation service is operational
func (s *ConsolidationService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("consolidation service not initialized")
	}

	if !s.config.Enabled {
		return nil // Service is healthy but disabled
	}

	if !s.IsRunning() {
		return fmt.Errorf("consolidation service not running")
	}

	// Check if engine is responsive by getting stats
	stats := s.engine.GetStats()
	if stats == nil {
		return fmt.Errorf("consolidation engine not responding")
	}

	return nil
}

// Engine returns the consolidation engine
func (s *ConsolidationService) Engine() *consolidation.Engine {
	return s.engine
}

// TriggerConsolidation triggers a consolidation event
func (s *ConsolidationService) TriggerConsolidation(eventType consolidation.EventType, trigger string, memories []*types.MemoryEntry) error {
	if !s.IsInitialized() || !s.IsRunning() {
		return fmt.Errorf("consolidation service not ready")
	}

	if !s.config.Enabled {
		s.logger.Debug("Consolidation trigger ignored - service disabled", "event_type", eventType.String())
		return nil
	}

	return s.engine.TriggerEvent(eventType, trigger, memories)
}

// UpdateContextUsage updates the context usage monitoring
func (s *ConsolidationService) UpdateContextUsage(tokens int) {
	if s.engine != nil && s.config.Enabled {
		s.engine.UpdateContextUsage(tokens)
	}
}

// GetStats returns consolidation statistics
func (s *ConsolidationService) GetStats() map[string]any {
	if s.engine == nil {
		return map[string]any{
			"enabled": s.config.Enabled,
			"initialized": s.IsInitialized(),
			"running": s.IsRunning(),
		}
	}
	
	stats := s.engine.GetStats()
	stats["enabled"] = s.config.Enabled
	stats["initialized"] = s.IsInitialized()
	stats["running"] = s.IsRunning()
	
	return stats
}