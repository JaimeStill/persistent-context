package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Configurable defines the interface for package-specific configurations
type Configurable interface {
	LoadConfig(*viper.Viper) error
	ValidateConfig() error
	GetDefaults() map[string]any
}

// Config holds all application configuration
type Config struct {
	HTTP          HTTPConfig          `mapstructure:"server"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	VectorDB      VectorDBConfig      `mapstructure:"vectordb"`
	LLM           LLMConfig           `mapstructure:"llm"`
	Journal       JournalConfig       `mapstructure:"journal"`
	MCP           MCPConfig           `mapstructure:"mcp"`
	Persona       PersonaConfig       `mapstructure:"persona"`
	Consolidation ConsolidationConfig `mapstructure:"consolidation"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()
	
	// Set defaults from all packages
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
	
	// Load package-specific configurations
	if err := config.loadPackageConfigs(v); err != nil {
		return nil, fmt.Errorf("failed to load package configs: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// loadPackageConfigs loads configuration for all packages
func (c *Config) loadPackageConfigs(v *viper.Viper) error {
	configurables := []Configurable{
		&c.HTTP,
		&c.Logging,
		&c.VectorDB,
		&c.LLM,
		&c.Journal,
		&c.MCP,
		&c.Persona,
		&c.Consolidation,
	}
	
	for _, configurable := range configurables {
		if err := configurable.LoadConfig(v); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}
	
	return nil
}

// setDefaults sets default configuration values from all packages
func setDefaults(v *viper.Viper) {
	// Load defaults from all packages
	configurables := []Configurable{
		&HTTPConfig{},
		&LoggingConfig{},
		&VectorDBConfig{},
		&LLMConfig{},
		&JournalConfig{},
		&MCPConfig{},
		&PersonaConfig{},
		&ConsolidationConfig{},
	}
	
	for _, configurable := range configurables {
		defaults := configurable.GetDefaults()
		for key, value := range defaults {
			v.SetDefault(key, value)
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate all package configurations
	configurables := []Configurable{
		&c.HTTP,
		&c.Logging,
		&c.VectorDB,
		&c.LLM,
		&c.Journal,
		&c.MCP,
		&c.Persona,
		&c.Consolidation,
	}
	
	for _, configurable := range configurables {
		if err := configurable.ValidateConfig(); err != nil {
			return err
		}
	}
	
	return nil
}