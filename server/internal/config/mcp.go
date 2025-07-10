package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Enabled       bool          `mapstructure:"enabled"`        // Enable MCP server
	Name          string        `mapstructure:"name"`           // MCP server name
	Version       string        `mapstructure:"version"`        // MCP server version
	BufferSize    int           `mapstructure:"buffer_size"`    // Context buffer size
	WorkerCount   int           `mapstructure:"worker_count"`   // Number of worker goroutines
	RetryAttempts int           `mapstructure:"retry_attempts"` // Max retry attempts
	RetryDelay    time.Duration `mapstructure:"retry_delay"`    // Delay between retries
	Timeout       time.Duration `mapstructure:"timeout"`        // Processing timeout
}

// LoadConfig loads configuration from viper
func (c *MCPConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("mcp", c)
}

// ValidateConfig validates the configuration
func (c *MCPConfig) ValidateConfig() error {
	if c.Name == "" {
		return fmt.Errorf("mcp name cannot be empty")
	}
	
	if c.Version == "" {
		return fmt.Errorf("mcp version cannot be empty")
	}
	
	if c.BufferSize < 0 {
		return fmt.Errorf("buffer size cannot be negative")
	}
	
	if c.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be positive")
	}
	
	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}
	
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *MCPConfig) GetDefaults() map[string]any {
	return map[string]any{
		"mcp.enabled":        false,
		"mcp.name":           "persistent-context",
		"mcp.version":        "1.0.0",
		"mcp.buffer_size":    1000,
		"mcp.worker_count":   2,
		"mcp.retry_attempts": 3,
		"mcp.retry_delay":    "1s",
		"mcp.timeout":        "30s",
	}
}