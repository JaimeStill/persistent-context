package services

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
	"github.com/JaimeStill/persistent-context/internal/memory"
	"github.com/JaimeStill/persistent-context/internal/vectordb"
)

// MemoryService wraps memory storage operations as a managed service
type MemoryService struct {
	BaseService
	store  *memory.MemoryStore
	config *config.MemoryConfig
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
	s.SetInitialized(true)
	return nil
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