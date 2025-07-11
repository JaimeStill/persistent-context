package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaimeStill/persistent-context/app/middleware"
	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/consolidation"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/memory"
	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// MemoryService wraps memory storage operations as a managed service
type MemoryService struct {
	BaseService
	store         *memory.MemoryStore
	config        *config.MemoryConfig
	pipeline      *middleware.Pipeline
	consolidation *ConsolidationService
}

// NewMemoryService creates a new memory service
func NewMemoryService(cfg *config.MemoryConfig) *MemoryService {
	return &MemoryService{
		BaseService: NewBaseService("memory", "vectordb", "llm"), // Depends on vectordb and llm
		config:      cfg,
	}
}

// Initialize creates the memory store with its dependencies
func (s *MemoryService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// This will be handled by the registry - dependencies should be injected
	s.SetInitialized(true)
	return nil
}

// InitializeWithDependencies initializes the memory service with injected dependencies
func (s *MemoryService) InitializeWithDependencies(vdb vectordb.VectorDB, llmClient llm.LLM) error {
	if s.IsInitialized() {
		return nil
	}

	// Create memory store dependencies
	deps := &memory.Dependencies{
		VectorDB:  vdb,
		LLMClient: llmClient,
		Config:    s.config,
	}

	// Create memory store
	s.store = memory.NewMemoryStore(deps)
	
	// Create and configure middleware pipeline
	s.pipeline = middleware.NewPipeline(slog.Default())
	s.setupPipeline()
	
	s.SetInitialized(true)
	return nil
}

// SetConsolidationService sets the consolidation service for middleware integration
func (s *MemoryService) SetConsolidationService(consolidation *ConsolidationService) {
	s.consolidation = consolidation
	// Re-setup pipeline with consolidation integration
	if s.pipeline != nil {
		s.setupPipeline()
	}
}

// setupPipeline configures the memory processing pipeline
func (s *MemoryService) setupPipeline() {
	// Add middleware in order of execution
	s.pipeline.Use(middleware.LoggingMiddleware(slog.Default()))
	s.pipeline.Use(middleware.ValidationMiddleware)
	s.pipeline.Use(middleware.EnrichmentMiddleware)
	
	// Add consolidation middleware (if consolidation service is available)
	// This triggers consolidation events based on memory processing
	consolidationTrigger := func(ctx context.Context, memCtx *middleware.MemoryContext) error {
		if s.consolidation != nil {
			// Trigger threshold-based consolidation check
			if err := s.consolidation.TriggerConsolidation(
				consolidation.ThresholdReached,
				"memory_processing",
				[]*types.MemoryEntry{memCtx.Memory},
			); err != nil {
				slog.Error("Failed to trigger consolidation", "error", err, "memory_id", memCtx.Memory.ID)
				return err
			}
		} else {
			slog.Debug("Consolidation service not available", "memory_id", memCtx.Memory.ID)
		}
		return nil
	}
	s.pipeline.Use(middleware.ConsolidationMiddleware(consolidationTrigger))
}

// Start begins memory operations
func (s *MemoryService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	s.SetRunning(true)
	return nil
}

// Stop gracefully shuts down the memory service
func (s *MemoryService) Stop(ctx context.Context) error {
	s.SetRunning(false)
	return nil
}

// HealthCheck verifies the memory service is operational
func (s *MemoryService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	return s.store.HealthCheck(ctx)
}

// Store returns the memory store
func (s *MemoryService) Store() *memory.MemoryStore {
	return s.store
}

// Pipeline returns the middleware pipeline
func (s *MemoryService) Pipeline() *middleware.Pipeline {
	return s.pipeline
}

// ProcessMemory processes a memory through the middleware pipeline
func (s *MemoryService) ProcessMemory(ctx context.Context, memory *types.MemoryEntry, source string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("memory service not initialized")
	}
	
	// Process through middleware pipeline
	return s.pipeline.Process(ctx, memory, source)
}