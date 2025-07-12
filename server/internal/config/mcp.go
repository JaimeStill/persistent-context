package config

import (
	"fmt"
	"time"

	"github.com/JaimeStill/persistent-context/internal/types"
	"github.com/spf13/viper"
)

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Enabled        bool          `mapstructure:"enabled"`         // Enable MCP server
	Name           string        `mapstructure:"name"`            // MCP server name
	Version        string        `mapstructure:"version"`         // MCP server version
	ServerEndpoint string        `mapstructure:"server_endpoint"` // Journal API endpoint
	CaptureMode    string        `mapstructure:"capture_mode"`    // Active capture profile
	ProfilesDir    string        `mapstructure:"profiles_dir"`    // External profiles directory
	Includes       []string      `mapstructure:"includes"`        // Additional config files
	
	// Performance settings
	BufferSize        int           `mapstructure:"buffer_size"`        // Context buffer size
	BatchWindowMs     int           `mapstructure:"batch_window_ms"`    // Batching time window
	MaxBatchSize      int           `mapstructure:"max_batch_size"`     // Maximum captures per batch
	CacheSize         int           `mapstructure:"cache_size"`         // Embedding cache size
	PriorityQueueSize int           `mapstructure:"priority_queue_size"` // Priority queue capacity
	WorkerCount       int           `mapstructure:"worker_count"`       // Number of worker goroutines
	RetryAttempts     int           `mapstructure:"retry_attempts"`     // Max retry attempts
	RetryDelay        time.Duration `mapstructure:"retry_delay"`        // Delay between retries
	Timeout           time.Duration `mapstructure:"timeout"`            // Processing timeout
	
	// Filter rules and profiles
	FilterRules types.FilterRules          `mapstructure:"filter_rules"` // Capture filtering rules
	Profiles    map[string]*types.Profile `mapstructure:"profiles"`     // Capture mode profiles
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
	
	if c.ServerEndpoint == "" {
		return fmt.Errorf("server endpoint cannot be empty")
	}
	
	if c.CaptureMode == "" {
		return fmt.Errorf("capture mode cannot be empty")
	}
	
	// Validate performance settings
	if c.BufferSize < 0 {
		return fmt.Errorf("buffer size cannot be negative")
	}
	
	if c.BatchWindowMs <= 0 {
		return fmt.Errorf("batch window must be positive")
	}
	
	if c.MaxBatchSize <= 0 {
		return fmt.Errorf("max batch size must be positive")
	}
	
	if c.CacheSize <= 0 {
		return fmt.Errorf("cache size must be positive")
	}
	
	if c.PriorityQueueSize <= 0 {
		return fmt.Errorf("priority queue size must be positive")
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
	
	// Validate filter rules
	if c.FilterRules.FileOperations.MinChangeSize < 0 {
		return fmt.Errorf("file operations min change size cannot be negative")
	}
	
	if c.FilterRules.FileOperations.DebounceMs < 0 {
		return fmt.Errorf("file operations debounce cannot be negative")
	}
	
	if c.FilterRules.FileOperations.MaxFileSize <= 0 {
		return fmt.Errorf("file operations max file size must be positive")
	}
	
	if c.FilterRules.CommandExecution.MaxOutputLines <= 0 {
		return fmt.Errorf("command execution max output lines must be positive")
	}
	
	if c.FilterRules.SearchOperations.MinResults < 0 {
		return fmt.Errorf("search operations min results cannot be negative")
	}
	
	if c.FilterRules.SearchOperations.MaxResults <= 0 {
		return fmt.Errorf("search operations max results must be positive")
	}
	
	if c.FilterRules.SearchOperations.BatchWindowMs <= 0 {
		return fmt.Errorf("search operations batch window must be positive")
	}
	
	// Validate profiles
	for name, profile := range c.Profiles {
		if profile == nil {
			return fmt.Errorf("profile %s cannot be nil", name)
		}
		
		if profile.Name == "" {
			return fmt.Errorf("profile %s name cannot be empty", name)
		}
		
		if profile.DebounceMultiplier <= 0 {
			return fmt.Errorf("profile %s debounce multiplier must be positive", name)
		}
		
		if profile.CaptureThreshold < 0 || profile.CaptureThreshold > 1 {
			return fmt.Errorf("profile %s capture threshold must be between 0 and 1", name)
		}
	}
	
	return nil
}

// GetDefaults returns default configuration values
func (c *MCPConfig) GetDefaults() map[string]any {
	return map[string]any{
		"mcp.enabled":             false,
		"mcp.name":                "persistent-context-mcp",
		"mcp.version":             "1.0.0",
		"mcp.server_endpoint":     "http://localhost:8080",
		"mcp.capture_mode":        "balanced",
		"mcp.profiles_dir":        "~/.config/persistent-context/profiles",
		"mcp.includes":            []string{},
		
		// Performance defaults
		"mcp.buffer_size":         1000,
		"mcp.batch_window_ms":     5000,
		"mcp.max_batch_size":      10,
		"mcp.cache_size":          1000,
		"mcp.priority_queue_size": 100,
		"mcp.worker_count":        4,
		"mcp.retry_attempts":      3,
		"mcp.retry_delay":         "1s",
		"mcp.timeout":             "30s",
		
		// Default filter rules
		"mcp.filter_rules.file_operations.min_change_size":   50,
		"mcp.filter_rules.file_operations.debounce_ms":       2000,
		"mcp.filter_rules.file_operations.ignore_patterns":   []string{"*.tmp", "*.log", "node_modules/**", ".git/**", "bin/**", "data/**", "*.swp", "*.bak"},
		"mcp.filter_rules.file_operations.include_patterns":  []string{},
		"mcp.filter_rules.file_operations.max_file_size":     5242880, // 5MB
		
		"mcp.filter_rules.command_execution.capture_errors":    true,
		"mcp.filter_rules.command_execution.capture_patterns":  []string{"^(ERROR|FAIL|PANIC)", "test.*failed", "build.*error"},
		"mcp.filter_rules.command_execution.ignore_patterns":   []string{},
		"mcp.filter_rules.command_execution.max_output_lines":  5000,
		
		"mcp.filter_rules.search_operations.min_results":     1,
		"mcp.filter_rules.search_operations.max_results":     50,
		"mcp.filter_rules.search_operations.batch_window_ms": 30000,
		
		// Built-in profiles
		"mcp.profiles.conservative.name":                 "conservative",
		"mcp.profiles.conservative.description":          "Minimal capture for production environments",
		"mcp.profiles.conservative.debounce_multiplier":  2.5,
		"mcp.profiles.conservative.filter_strictness":    "high",
		"mcp.profiles.conservative.capture_threshold":    0.8,
		
		"mcp.profiles.balanced.name":                 "balanced",
		"mcp.profiles.balanced.description":          "Default balanced capture mode",
		"mcp.profiles.balanced.debounce_multiplier":  1.0,
		"mcp.profiles.balanced.filter_strictness":    "medium",
		"mcp.profiles.balanced.capture_threshold":    0.5,
		
		"mcp.profiles.aggressive.name":                 "aggressive",
		"mcp.profiles.aggressive.description":          "Maximum capture for debugging",
		"mcp.profiles.aggressive.debounce_multiplier":  0.25,
		"mcp.profiles.aggressive.filter_strictness":    "low",
		"mcp.profiles.aggressive.capture_threshold":    0.2,
	}
}