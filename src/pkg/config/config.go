package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Configurable defines the interface for package-specific configurations
type Configurable interface {
	LoadConfig(*viper.Viper) error
	ValidateConfig() error
	GetDefaults() map[string]any
}

// LoadViper creates and configures a Viper instance with common settings
func LoadViper(envPrefix string) *viper.Viper {
	v := viper.New()
	
	// Configure environment variables
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Try to read config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/persistent-context/")
	
	// Read config file if it exists (ignore errors)
	_ = v.ReadInConfig()
	
	return v
}