package services

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/llm"
)

// LLMService wraps LLM implementations as a managed service
type LLMService struct {
	BaseService
	llm    llm.LLM
	config *config.LLMConfig
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.LLMConfig) *LLMService {
	return &LLMService{
		BaseService: NewBaseService("llm"),
		config:      cfg,
	}
}

// Initialize creates the appropriate LLM implementation
func (s *LLMService) Initialize(ctx context.Context) error {
	if s.IsInitialized() {
		return nil
	}

	// Create LLM config
	llmConfig := &llm.Config{
		Provider:           s.config.Provider,
		URL:               s.config.URL,
		APIKey:            s.config.APIKey,
		EmbeddingModel:    s.config.EmbeddingModel,
		ConsolidationModel: s.config.ConsolidationModel,
		CacheEnabled:      s.config.CacheEnabled,
		MaxRetries:        s.config.MaxRetries,
		Timeout:           s.config.Timeout,
	}

	// Create the LLM implementation
	llmImpl, err := llm.NewLLM(llmConfig)
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	// Verify connectivity
	if err := llmImpl.HealthCheck(ctx); err != nil {
		return fmt.Errorf("failed to connect to LLM: %w", err)
	}

	s.llm = llmImpl
	s.SetInitialized(true)
	return nil
}

// Start begins LLM operations
func (s *LLMService) Start(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	s.SetRunning(true)
	return nil
}

// Stop gracefully shuts down the LLM service
func (s *LLMService) Stop(ctx context.Context) error {
	if s.llm != nil {
		s.llm.ClearCache()
	}
	s.SetRunning(false)
	return nil
}

// HealthCheck verifies the LLM is accessible
func (s *LLMService) HealthCheck(ctx context.Context) error {
	if !s.IsInitialized() {
		return fmt.Errorf("service not initialized")
	}

	return s.llm.HealthCheck(ctx)
}

// LLM returns the LLM interface
func (s *LLMService) LLM() llm.LLM {
	return s.llm
}