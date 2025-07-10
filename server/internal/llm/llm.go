package llm

import (
	"context"
	"fmt"
	"time"
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

// Config holds configuration for LLM implementations
type Config struct {
	Provider           string        `mapstructure:"provider"`
	URL               string        `mapstructure:"url"`
	APIKey            string        `mapstructure:"api_key"`
	EmbeddingModel    string        `mapstructure:"embedding_model"`
	ConsolidationModel string       `mapstructure:"consolidation_model"`
	CacheEnabled      bool          `mapstructure:"cache_enabled"`
	MaxRetries        int           `mapstructure:"max_retries"`
	Timeout           time.Duration `mapstructure:"timeout"`
}

// NewLLM creates a new LLM implementation based on the provider
func NewLLM(config *Config) (LLM, error) {
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