package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig loads configuration from viper
func (c *LoggingConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("logging", c)
}

// ValidateConfig validates the configuration
func (c *LoggingConfig) ValidateConfig() error {
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !slices.Contains(validLogLevels, c.Level) {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)", 
			c.Level, strings.Join(validLogLevels, ", "))
	}
	
	validLogFormats := []string{"json", "text"}
	if !slices.Contains(validLogFormats, c.Format) {
		return fmt.Errorf("invalid log format: %s (must be one of: %s)", 
			c.Format, strings.Join(validLogFormats, ", "))
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *LoggingConfig) GetDefaults() map[string]any {
	return map[string]any{
		"logging.level":  "info",
		"logging.format": "json",
	}
}