package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Name      string        `mapstructure:"name"`        // MCP server name
	Version   string        `mapstructure:"version"`     // MCP server version
	WebAPIURL string        `mapstructure:"web_api_url"` // Web API URL for HTTP client
	Timeout   time.Duration `mapstructure:"timeout"`     // HTTP client timeout
}

// LoadConfig processes configuration after main unmarshaling
// The main config.Load() already unmarshals all values including environment variables.
// This method is for post-processing only and currently has no additional behavior needed.
func (c *MCPConfig) LoadConfig(v *viper.Viper) error {
	return nil
}

// ValidateConfig validates the configuration
func (c *MCPConfig) ValidateConfig() error {
	if c.Name == "" {
		return fmt.Errorf("mcp name cannot be empty")
	}
	
	if c.Version == "" {
		return fmt.Errorf("mcp version cannot be empty")
	}
	
	if c.WebAPIURL == "" {
		return fmt.Errorf("web API URL cannot be empty")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *MCPConfig) GetDefaults() map[string]any {
	return map[string]any{
		"mcp.name":        "persistent-context-mcp",
		"mcp.version":     "1.0.0",
		"mcp.web_api_url": "http://localhost:8543",
		"mcp.timeout":     "30s",
	}
}