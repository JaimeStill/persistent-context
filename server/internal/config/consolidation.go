package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ConsolidationConfig holds consolidation engine configuration
type ConsolidationConfig struct {
	MaxTokens              int     `mapstructure:"max_tokens"`               // Maximum context window size
	SafetyMargin           float64 `mapstructure:"safety_margin"`            // Safety margin for context window (0.0-1.0)
	MemoryCountThreshold   int     `mapstructure:"memory_count_threshold"`   // Number of memories to trigger consolidation
	EmbeddingSizeThreshold uint64  `mapstructure:"embedding_size_threshold"` // Total embedding size threshold
	ContextUsageThreshold  float64 `mapstructure:"context_usage_threshold"`  // Context usage percentage threshold
	DecayFactor            float64 `mapstructure:"decay_factor"`             // Time decay factor for memory importance
	AccessWeight           float64 `mapstructure:"access_weight"`            // Weight for access frequency in scoring
	RelevanceWeight        float64 `mapstructure:"relevance_weight"`         // Weight for semantic relevance in scoring
	Enabled                bool    `mapstructure:"enabled"`                  // Enable/disable consolidation
}

// LoadConfig loads configuration from viper
func (c *ConsolidationConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("consolidation", c)
}

// ValidateConfig validates the configuration
func (c *ConsolidationConfig) ValidateConfig() error {
	if c.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	
	if c.SafetyMargin < 0.0 || c.SafetyMargin > 1.0 {
		return fmt.Errorf("safety_margin must be between 0.0 and 1.0")
	}
	
	if c.MemoryCountThreshold <= 0 {
		return fmt.Errorf("memory_count_threshold must be positive")
	}
	
	if c.EmbeddingSizeThreshold == 0 {
		return fmt.Errorf("embedding_size_threshold must be positive")
	}
	
	if c.ContextUsageThreshold < 0.0 || c.ContextUsageThreshold > 1.0 {
		return fmt.Errorf("context_usage_threshold must be between 0.0 and 1.0")
	}
	
	if c.DecayFactor < 0.0 {
		return fmt.Errorf("decay_factor must be non-negative")
	}
	
	if c.AccessWeight < 0.0 || c.RelevanceWeight < 0.0 {
		return fmt.Errorf("access_weight and relevance_weight must be non-negative")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *ConsolidationConfig) GetDefaults() map[string]any {
	return map[string]any{
		"consolidation.max_tokens":                128000,  // Typical large context window
		"consolidation.safety_margin":             0.7,     // Use 70% of context window
		"consolidation.memory_count_threshold":    50,      // Consolidate after 50 memories
		"consolidation.embedding_size_threshold":  1048576, // 1MB of embeddings
		"consolidation.context_usage_threshold":   0.8,     // Trigger at 80% usage
		"consolidation.decay_factor":              0.01,    // Gentle time decay
		"consolidation.access_weight":             2.0,     // Weight access frequency highly
		"consolidation.relevance_weight":          1.5,     // Weight semantic relevance
		"consolidation.enabled":                   true,    // Enable by default
	}
}