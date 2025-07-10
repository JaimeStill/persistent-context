package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// StorageConfig holds storage configuration
type StorageConfig struct {
	PersonaPath string `mapstructure:"persona_path"` // Path to store persona data
}

// LoadConfig loads configuration from viper
func (c *StorageConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("storage", c)
}

// ValidateConfig validates the configuration
func (c *StorageConfig) ValidateConfig() error {
	if c.PersonaPath == "" {
		return fmt.Errorf("persona path cannot be empty")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *StorageConfig) GetDefaults() map[string]any {
	return map[string]any{
		"storage.persona_path": "/data/personas",
	}
}