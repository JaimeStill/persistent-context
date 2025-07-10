package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MemoryConfig holds memory processing configuration
type MemoryConfig struct {
	BatchSize             uint64        `mapstructure:"batch_size"`             // Batch size for processing
	RetentionDays         int           `mapstructure:"retention_days"`         // Days to retain memories
	ConsolidationInterval time.Duration `mapstructure:"consolidation_interval"` // How often to consolidate
	MaxMemorySize         uint64        `mapstructure:"max_memory_size"`        // Max memories to keep
	StrengthThreshold     float32       `mapstructure:"strength_threshold"`     // Minimum strength to keep
}

// LoadConfig loads configuration from viper
func (c *MemoryConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("memory", c)
}

// ValidateConfig validates the configuration
func (c *MemoryConfig) ValidateConfig() error {
	if c.BatchSize == 0 {
		return fmt.Errorf("batch size must be positive")
	}
	
	if c.RetentionDays < 0 {
		return fmt.Errorf("retention days cannot be negative")
	}
	
	if c.ConsolidationInterval <= 0 {
		return fmt.Errorf("consolidation interval must be positive")
	}
	
	if c.MaxMemorySize == 0 {
		return fmt.Errorf("max memory size must be positive")
	}
	
	if c.StrengthThreshold < 0 || c.StrengthThreshold > 1 {
		return fmt.Errorf("strength threshold must be between 0 and 1")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *MemoryConfig) GetDefaults() map[string]any {
	return map[string]any{
		"memory.batch_size":             100,
		"memory.retention_days":         30,
		"memory.consolidation_interval": "6h",
		"memory.max_memory_size":        10000,
		"memory.strength_threshold":     0.1,
	}
}