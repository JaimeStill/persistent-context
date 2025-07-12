package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// JournalConfig holds journal processing configuration
type JournalConfig struct {
	BatchSize             uint64        `mapstructure:"batch_size"`             // Batch size for processing
	RetentionDays         int           `mapstructure:"retention_days"`         // Days to retain memories
	ConsolidationInterval time.Duration `mapstructure:"consolidation_interval"` // How often to consolidate
	MaxMemorySize         uint64        `mapstructure:"max_memory_size"`        // Max memories to keep
	StrengthThreshold     float32       `mapstructure:"strength_threshold"`     // Minimum strength to keep
}

// LoadConfig loads configuration from viper
func (c *JournalConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("journal", c)
}

// ValidateConfig validates the configuration
func (c *JournalConfig) ValidateConfig() error {
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
func (c *JournalConfig) GetDefaults() map[string]any {
	return map[string]any{
		"journal.batch_size":             100,
		"journal.retention_days":         30,
		"journal.consolidation_interval": "6h",
		"journal.max_memory_size":        10000,
		"journal.strength_threshold":     0.1,
	}
}