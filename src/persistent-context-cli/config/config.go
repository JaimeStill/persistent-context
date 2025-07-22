package config

import (
	"time"
)

// Config represents the CLI tool configuration
type Config struct {
	// Connection settings
	WebURL     string `mapstructure:"web_url"`
	DirectMode bool   `mapstructure:"direct_mode"`
	
	// Direct mode connection settings (for future use)
	QdrantURL  string `mapstructure:"qdrant_url"`
	QdrantPort int    `mapstructure:"qdrant_port"`
	OllamaURL  string `mapstructure:"ollama_url"`
	
	// Performance settings
	Timeout time.Duration `mapstructure:"timeout"`
	
	// Display settings
	OutputFormat string `mapstructure:"output_format"` // "text", "json"
	Verbose      bool   `mapstructure:"verbose"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WebURL:       "http://localhost:8543",
		DirectMode:   false, // Use web service by default
		QdrantURL:    "localhost",
		QdrantPort:   6334,
		OllamaURL:    "http://localhost:11434",
		Timeout:      30 * time.Second,
		OutputFormat: "text",
		Verbose:      false,
	}
}