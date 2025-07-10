package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// LLMConfig holds LLM configuration
type LLMConfig struct {
	Provider             string        `mapstructure:"provider"`               // "ollama", "openai", etc.
	URL                  string        `mapstructure:"url"`                    // LLM service URL
	EmbeddingModel       string        `mapstructure:"embedding_model"`        // Model for embeddings
	ConsolidationModel   string        `mapstructure:"consolidation_model"`    // Model for consolidation
	CacheEnabled         bool          `mapstructure:"cache_enabled"`          // Enable embedding cache
	CacheTTL             time.Duration `mapstructure:"cache_ttl"`              // Cache TTL
	Timeout              time.Duration `mapstructure:"timeout"`                // Request timeout
	MaxRetries           int           `mapstructure:"max_retries"`            // Max retry attempts
}

// LoadConfig loads configuration from viper
func (c *LLMConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("llm", c)
}

// ValidateConfig validates the configuration
func (c *LLMConfig) ValidateConfig() error {
	if c.Provider == "" {
		return fmt.Errorf("llm provider cannot be empty")
	}
	
	if c.URL == "" {
		return fmt.Errorf("llm URL cannot be empty")
	}
	
	if c.EmbeddingModel == "" {
		return fmt.Errorf("embedding model cannot be empty")
	}
	
	if c.ConsolidationModel == "" {
		return fmt.Errorf("consolidation model cannot be empty")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.CacheTTL <= 0 && c.CacheEnabled {
		return fmt.Errorf("cache TTL must be positive when cache is enabled")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *LLMConfig) GetDefaults() map[string]any {
	return map[string]any{
		"llm.provider":             "ollama",
		"llm.url":                  "http://ollama:11434",
		"llm.embedding_model":      "phi3:mini",
		"llm.consolidation_model":  "phi3:mini",
		"llm.cache_enabled":        true,
		"llm.cache_ttl":            "1h",
		"llm.timeout":              "30s",
		"llm.max_retries":          3,
	}
}