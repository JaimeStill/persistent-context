package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// VectorDBConfig holds vector database configuration
type VectorDBConfig struct {
	Provider               string            `mapstructure:"provider"`               // "qdrant", "weaviate", etc.
	URL                    string            `mapstructure:"url"`                    // Database URL
	MemoryCollections      map[string]string `mapstructure:"memory_collections"`      // Memory type -> collection name
	AssociationsCollection string            `mapstructure:"associations_collection"` // Association collection name
	VectorDimension        int               `mapstructure:"vector_dimension"`       // Vector embedding dimension
	OnDiskPayload          bool              `mapstructure:"on_disk_payload"`        // Use disk storage for payloads
	Timeout                time.Duration     `mapstructure:"timeout"`                // Connection timeout
	Insecure               bool              `mapstructure:"insecure"`               // Disable TLS for development
}

// LoadConfig loads configuration from viper
func (c *VectorDBConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("vectordb", c)
}

// ValidateConfig validates the configuration
func (c *VectorDBConfig) ValidateConfig() error {
	if c.Provider == "" {
		return fmt.Errorf("vectordb provider cannot be empty")
	}

	if c.URL == "" {
		return fmt.Errorf("vectordb URL cannot be empty")
	}

	if c.VectorDimension <= 0 {
		return fmt.Errorf("vector dimension must be positive")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate memory collections
	if len(c.MemoryCollections) == 0 {
		return fmt.Errorf("memory collections cannot be empty")
	}

	return nil
}

// GetDefaults returns default configuration values
func (c *VectorDBConfig) GetDefaults() map[string]any {
	return map[string]any{
		"vectordb.provider":         "qdrant",
		"vectordb.url":              "qdrant:6334",
		"vectordb.insecure":         true,
		"vectordb.vector_dimension": 3072,
		"vectordb.on_disk_payload":  true,
		"vectordb.timeout":          "30s",
		"vectordb.memory_collections": map[string]string{
			"episodic":      "episodic_memories",
			"semantic":      "semantic_memories",
			"procedural":    "procedural_memories",
			"metacognitive": "metacognitive_memories",
		},
		"vectordb.associations_collection": "associations",
	}
}
