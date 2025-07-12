package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// PersonaConfig holds persona storage configuration
type PersonaConfig struct {
	Path string `mapstructure:"path"` // Path to store persona data
}

// LoadConfig loads configuration from viper
func (c *PersonaConfig) LoadConfig(v *viper.Viper) error {
	return nil
}

// ValidateConfig validates the configuration
func (c *PersonaConfig) ValidateConfig() error {
	if c.Path == "" {
		return fmt.Errorf("persona path cannot be empty")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *PersonaConfig) GetDefaults() map[string]any {
	return map[string]any{
		"persona.path": "/data/personas",
	}
}