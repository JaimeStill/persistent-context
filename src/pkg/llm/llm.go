package llm

import (
	"context"
	"fmt"

	"github.com/JaimeStill/persistent-context/pkg/config"
)

// LLM defines the interface for Large Language Model operations
type LLM interface {
	// GenerateEmbedding creates vector embeddings for the given text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// ConsolidateMemories uses the LLM to consolidate multiple memories into semantic knowledge
	ConsolidateMemories(ctx context.Context, memories []string) (string, error)

	// HealthCheck verifies the LLM service is accessible
	HealthCheck(ctx context.Context) error

	// ClearCache clears any internal caches
	ClearCache()
}

// NewLLM creates a new LLM implementation based on the provider
func NewLLM(config *config.LLMConfig) (LLM, error) {
	switch config.Provider {
	case "ollama":
		return NewOllamaLLM(config)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	case "localai":
		return nil, fmt.Errorf("LocalAI provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}
