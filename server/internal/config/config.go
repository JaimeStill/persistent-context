package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Qdrant  QdrantConfig  `mapstructure:"qdrant"`
	Ollama  OllamaConfig  `mapstructure:"ollama"`
	Storage StorageConfig `mapstructure:"storage"`
	Logging LoggingConfig `mapstructure:"logging"`
	MCP     MCPConfig     `mapstructure:"mcp"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string `mapstructure:"port"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// QdrantConfig holds Qdrant vector database configuration
type QdrantConfig struct {
	URL            string `mapstructure:"url"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     int    `mapstructure:"vector_size"`
}

// OllamaConfig holds Ollama LLM configuration
type OllamaConfig struct {
	URL   string `mapstructure:"url"`
	Model string `mapstructure:"model"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	PersonaPath string `mapstructure:"persona_path"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()
	
	// Set defaults
	setDefaults(v)
	
	// Configure environment variables
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Try to read config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/persistent-context/")
	
	// Read config file if it exists (ignore errors)
	_ = v.ReadInConfig()
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)
	v.SetDefault("server.shutdown_timeout", 30)
	
	// Qdrant defaults
	v.SetDefault("qdrant.url", "http://qdrant:6333")
	v.SetDefault("qdrant.collection_name", "memories")
	v.SetDefault("qdrant.vector_size", 1536)
	
	// Ollama defaults
	v.SetDefault("ollama.url", "http://ollama:11434")
	v.SetDefault("ollama.model", "phi3:mini")
	
	// Storage defaults
	v.SetDefault("storage.persona_path", "/data/personas")
	
	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	
	// MCP defaults
	v.SetDefault("mcp.enabled", false)
	v.SetDefault("mcp.name", "persistent-context")
	v.SetDefault("mcp.version", "1.0.0")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}
	
	if c.Qdrant.URL == "" {
		return fmt.Errorf("qdrant URL cannot be empty")
	}
	
	if c.Ollama.URL == "" {
		return fmt.Errorf("ollama URL cannot be empty")
	}
	
	if c.Storage.PersonaPath == "" {
		return fmt.Errorf("persona path cannot be empty")
	}
	
	if c.Qdrant.VectorSize <= 0 {
		return fmt.Errorf("vector size must be positive")
	}
	
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)", 
			c.Logging.Level, strings.Join(validLogLevels, ", "))
	}
	
	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, c.Logging.Format) {
		return fmt.Errorf("invalid log format: %s (must be one of: %s)", 
			c.Logging.Format, strings.Join(validLogFormats, ", "))
	}
	
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}