package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port            string `mapstructure:"port"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// LoadConfig processes configuration after main unmarshaling
// The main config.Load() already unmarshals all values including environment variables.
// This method is for post-processing only and currently has no additional behavior needed.
func (c *HTTPConfig) LoadConfig(v *viper.Viper) error {
	return nil
}

// ValidateConfig validates the configuration
func (c *HTTPConfig) ValidateConfig() error {
	if c.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}
	
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *HTTPConfig) GetDefaults() map[string]any {
	return map[string]any{
		"server.port":             "8543",
		"server.read_timeout":     10,
		"server.write_timeout":    10,
		"server.shutdown_timeout": 30,
	}
}