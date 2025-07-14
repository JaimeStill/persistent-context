package app

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/JaimeStill/persistent-context/pkg/config"
)

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port            int `mapstructure:"port"`
	ReadTimeout     int `mapstructure:"read_timeout"`
	WriteTimeout    int `mapstructure:"write_timeout"`
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
}

// LoadConfig loads configuration from viper
func (c *HTTPConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("server", c)
}

// ValidateConfig validates the configuration
func (c *HTTPConfig) ValidateConfig() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Port)
	}
	
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}
	
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}
	
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be positive")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *HTTPConfig) GetDefaults() map[string]any {
	return map[string]any{
		"server.port":             8543,
		"server.read_timeout":     30,
		"server.write_timeout":    30,
		"server.shutdown_timeout": 30,
	}
}

// PersonaConfig holds persona management configuration
type PersonaConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	StoragePath   string `mapstructure:"storage_path"`
	MaxPersonas   int    `mapstructure:"max_personas"`
	MaxVersions   int    `mapstructure:"max_versions"`
}

// LoadConfig loads configuration from viper
func (c *PersonaConfig) LoadConfig(v *viper.Viper) error {
	return v.UnmarshalKey("persona", c)
}

// ValidateConfig validates the configuration
func (c *PersonaConfig) ValidateConfig() error {
	if c.MaxPersonas < 0 {
		return fmt.Errorf("max_personas cannot be negative")
	}
	
	if c.MaxVersions < 0 {
		return fmt.Errorf("max_versions cannot be negative")
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *PersonaConfig) GetDefaults() map[string]any {
	return map[string]any{
		"persona.enabled":       true,
		"persona.storage_path":  "./data/personas/",
		"persona.max_personas":  100,
		"persona.max_versions":  10,
	}
}

// Config holds all web service configuration
type Config struct {
	HTTP     HTTPConfig             `mapstructure:"server"`
	Logging  config.LoggingConfig   `mapstructure:"logging"`
	VectorDB config.VectorDBConfig  `mapstructure:"vectordb"`
	LLM      config.LLMConfig       `mapstructure:"llm"`
	Journal  config.JournalConfig   `mapstructure:"journal"`
	Persona  PersonaConfig          `mapstructure:"persona"`
	Memory   config.MemoryConfig    `mapstructure:"memory"`
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
	
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Load package-specific configurations
	if err := cfg.loadPackageConfigs(v); err != nil {
		return nil, fmt.Errorf("failed to load package configs: %w", err)
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &cfg, nil
}

// loadPackageConfigs loads configuration for all packages
func (c *Config) loadPackageConfigs(v *viper.Viper) error {
	configurables := []config.Configurable{
		&c.HTTP,
		&c.Logging,
		&c.VectorDB,
		&c.LLM,
		&c.Journal,
		&c.Persona,
		&c.Memory,
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
	configurables := []config.Configurable{
		&HTTPConfig{},
		&config.LoggingConfig{},
		&config.VectorDBConfig{},
		&config.LLMConfig{},
		&config.JournalConfig{},
		&PersonaConfig{},
		&config.MemoryConfig{},
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
	configurables := []config.Configurable{
		&c.HTTP,
		&c.Logging,
		&c.VectorDB,
		&c.LLM,
		&c.Journal,
		&c.Persona,
		&c.Memory,
	}
	
	for _, configurable := range configurables {
		if err := configurable.ValidateConfig(); err != nil {
			return err
		}
	}
	
	return nil
}