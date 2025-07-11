package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaimeStill/persistent-context/app/middleware"
	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/memory"
	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// MemoryService wraps memory storage operations as a managed service
type MemoryService struct {
	BaseService
	store    *memory.MemoryStore
	config   *config.MemoryConfig
	pipeline *middleware.Pipeline
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

// setupPipeline configures the memory processing pipeline
func (s *MemoryService) setupPipeline() {
	// Add middleware in order of execution
	s.pipeline.Use(middleware.LoggingMiddleware(slog.Default()))
	s.pipeline.Use(middleware.ValidationMiddleware)
	s.pipeline.Use(middleware.EnrichmentMiddleware)
	
	// Add consolidation middleware (if consolidation engine is available)
	// This would trigger consolidation events based on memory processing
	s.pipeline.Use(middleware.ConsolidationMiddleware(func(ctx context.Context, memCtx *middleware.MemoryContext) error {
		// TODO: Integrate with consolidation engine when available
		slog.Info("Consolidation trigger placeholder", "memory_id", memCtx.Memory.ID)
		return nil
	}))
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