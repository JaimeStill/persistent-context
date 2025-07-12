package types

import "time"

// EventType represents the type of capture event
type EventType string

const (
	EventTypeFileRead      EventType = "file_read"
	EventTypeFileWrite     EventType = "file_write"
	EventTypeFileDelete    EventType = "file_delete"
	EventTypeCommandRun    EventType = "command_run"
	EventTypeCommandOutput EventType = "command_output"
	EventTypeSearchQuery   EventType = "search_query"
	EventTypeSearchResults EventType = "search_results"
)

// Priority represents capture priority levels
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// CaptureEvent represents an event that might be captured
type CaptureEvent struct {
	Type        EventType              `json:"type"`        // Event type
	Source      string                 `json:"source"`      // Source identifier (file path, command, etc.)
	Content     string                 `json:"content"`     // Event content
	Metadata    map[string]any        `json:"metadata"`    // Additional metadata
	Timestamp   time.Time             `json:"timestamp"`   // Event timestamp
	Priority    Priority              `json:"priority"`    // Event priority
}

// FilterStrictness represents the strictness level for filtering
type FilterStrictness string

const (
	FilterStrictnessLow    FilterStrictness = "low"
	FilterStrictnessMedium FilterStrictness = "medium"
	FilterStrictnessHigh   FilterStrictness = "high"
)

// FilterRules defines capture filtering rules
type FilterRules struct {
	FileOperations   FileOperationRules   `mapstructure:"file_operations"`
	CommandExecution CommandExecutionRules `mapstructure:"command_execution"`
	SearchOperations SearchOperationRules  `mapstructure:"search_operations"`
}

// FileOperationRules defines file operation filtering
type FileOperationRules struct {
	MinChangeSize   int      `mapstructure:"min_change_size"`   // Minimum lines changed to trigger capture
	DebounceMs      int      `mapstructure:"debounce_ms"`       // Quiet period before capture
	IgnorePatterns  []string `mapstructure:"ignore_patterns"`   // Glob patterns to ignore
	IncludePatterns []string `mapstructure:"include_patterns"`  // Explicit include patterns
	MaxFileSize     int64    `mapstructure:"max_file_size"`     // Maximum file size to capture
}

// CommandExecutionRules defines command execution filtering
type CommandExecutionRules struct {
	CaptureErrors   bool     `mapstructure:"capture_errors"`    // Always capture error outputs
	CapturePatterns []string `mapstructure:"capture_patterns"`  // Regex patterns to capture
	IgnorePatterns  []string `mapstructure:"ignore_patterns"`   // Regex patterns to ignore
	MaxOutputLines  int      `mapstructure:"max_output_lines"`  // Maximum output lines to capture
}

// SearchOperationRules defines search operation filtering
type SearchOperationRules struct {
	MinResults    int `mapstructure:"min_results"`     // Minimum results to capture
	MaxResults    int `mapstructure:"max_results"`     // Maximum results to capture
	BatchWindowMs int `mapstructure:"batch_window_ms"` // Group searches within window
}

// Profile represents a capture mode profile
type Profile struct {
	Name               string           `mapstructure:"name"`                // Profile name
	Description        string           `mapstructure:"description"`         // Profile description
	Base               string           `mapstructure:"base"`                // Base profile to inherit from
	DebounceMultiplier float64          `mapstructure:"debounce_multiplier"` // Multiplier for debounce times
	FilterStrictness   FilterStrictness `mapstructure:"filter_strictness"`   // Filter strictness level
	CaptureThreshold   float64          `mapstructure:"capture_threshold"`   // Threshold for capture decision
	FilterRules        *FilterRules     `mapstructure:"filter_rules"`        // Override filter rules
	ScoringWeights     *ScoringWeights  `mapstructure:"scoring_weights"`     // Memory scoring weights
}

// ScoringWeights defines memory scoring weights
type ScoringWeights struct {
	Recency   float64 `mapstructure:"recency"`   // Weight for memory recency
	Frequency float64 `mapstructure:"frequency"` // Weight for access frequency
	Relevance float64 `mapstructure:"relevance"` // Weight for context relevance
}