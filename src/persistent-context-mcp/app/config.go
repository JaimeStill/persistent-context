package app

import (
	"fmt"
	"os"
	"time"
)

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Name      string        `mapstructure:"name"`        // MCP server name
	Version   string        `mapstructure:"version"`     // MCP server version
	WebAPIURL string        `mapstructure:"web_api_url"` // Web API URL for HTTP client
	Timeout   time.Duration `mapstructure:"timeout"`     // HTTP client timeout
}

// LoadConfig loads MCP configuration from environment variables with defaults
func LoadConfig() (*MCPConfig, error) {
	cfg := &MCPConfig{
		Name:      getEnvOrDefault("APP_MCP_NAME", "persistent-context-mcp"),
		Version:   getEnvOrDefault("APP_MCP_VERSION", "1.0.0"),
		WebAPIURL: getEnvOrDefault("APP_MCP_WEB_API_URL", "http://localhost:8543"),
		Timeout:   30 * time.Second,
	}

	if err := cfg.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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