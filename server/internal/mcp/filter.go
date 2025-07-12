package mcp

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/JaimeStill/persistent-context/internal/types"
)

// FilterEngine handles capture filtering based on rules and profiles
type FilterEngine struct {
	rules   types.FilterRules
	profile *types.Profile
	
	// Pattern caches for performance
	fileIgnorePatterns    []string
	fileIncludePatterns   []string
	commandCaptureRegex   []*regexp.Regexp
	commandIgnoreRegex    []*regexp.Regexp
	
	// Debouncing state
	lastFileEvents map[string]time.Time
	debounceTimers map[string]*time.Timer
}

// NewFilterEngine creates a new filtering engine
func NewFilterEngine(rules types.FilterRules, profile *types.Profile) *FilterEngine {
	fe := &FilterEngine{
		rules:          rules,
		profile:        profile,
		lastFileEvents: make(map[string]time.Time),
		debounceTimers: make(map[string]*time.Timer),
	}
	
	// Compile regex patterns
	fe.compilePatterns()
	
	return fe
}

// compilePatterns pre-compiles regex patterns for performance
func (fe *FilterEngine) compilePatterns() {
	// Compile command capture patterns
	for _, pattern := range fe.rules.CommandExecution.CapturePatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			fe.commandCaptureRegex = append(fe.commandCaptureRegex, regex)
		}
	}
	
	// Compile command ignore patterns
	for _, pattern := range fe.rules.CommandExecution.IgnorePatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			fe.commandIgnoreRegex = append(fe.commandIgnoreRegex, regex)
		}
	}
	
	// Store file patterns
	fe.fileIgnorePatterns = fe.rules.FileOperations.IgnorePatterns
	fe.fileIncludePatterns = fe.rules.FileOperations.IncludePatterns
}

// ShouldCapture determines if an event should be captured
func (fe *FilterEngine) ShouldCapture(event *types.CaptureEvent) (bool, types.Priority) {
	switch event.Type {
	case types.EventTypeFileRead, types.EventTypeFileWrite, types.EventTypeFileDelete:
		return fe.shouldCaptureFile(event)
	case types.EventTypeCommandRun, types.EventTypeCommandOutput:
		return fe.shouldCaptureCommand(event)
	case types.EventTypeSearchQuery, types.EventTypeSearchResults:
		return fe.shouldCaptureSearch(event)
	default:
		return false, types.PriorityLow
	}
}

// shouldCaptureFile determines if a file operation should be captured
func (fe *FilterEngine) shouldCaptureFile(event *types.CaptureEvent) (bool, types.Priority) {
	// Check file size limits
	if size, ok := event.Metadata["file_size"].(int64); ok {
		if size > fe.rules.FileOperations.MaxFileSize {
			return false, types.PriorityLow
		}
	}
	
	// Check if file should be filtered
	if fe.isFileFiltered(event.Source) {
		return false, types.PriorityLow
	}
	
	// Apply debouncing for file operations
	if event.Type == types.EventTypeFileWrite {
		if !fe.checkDebounce(event.Source) {
			return false, types.PriorityLow
		}
	}
	
	// Check change size for file writes
	if event.Type == types.EventTypeFileWrite {
		if changeSize, ok := event.Metadata["change_size"].(int); ok {
			if changeSize < fe.rules.FileOperations.MinChangeSize {
				return false, types.PriorityLow
			}
		}
	}
	
	// Determine priority
	priority := types.PriorityMedium
	if event.Type == types.EventTypeFileWrite {
		if changeSize, ok := event.Metadata["change_size"].(int); ok {
			if changeSize > fe.rules.FileOperations.MinChangeSize*5 {
				priority = types.PriorityHigh
			}
		}
	}
	
	return true, priority
}

// isFileFiltered checks if a file path should be filtered out
func (fe *FilterEngine) isFileFiltered(path string) bool {
	for _, pattern := range fe.fileIgnorePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			// Check if explicitly included
			if fe.matchesAnyIncludePattern(path) {
				return false // Not filtered - explicitly included
			}
			return true // Filtered - matches ignore but not include
		}
	}
	return false // Not filtered - doesn't match any ignore patterns
}

// matchesAnyIncludePattern checks if a path matches any include pattern
func (fe *FilterEngine) matchesAnyIncludePattern(path string) bool {
	for _, includePattern := range fe.fileIncludePatterns {
		if includeMatched, _ := filepath.Match(includePattern, path); includeMatched {
			return true
		}
	}
	return false
}

// shouldCaptureCommand determines if a command operation should be captured
func (fe *FilterEngine) shouldCaptureCommand(event *types.CaptureEvent) (bool, types.Priority) {
	content := event.Content
	
	// Always capture errors if enabled
	if fe.rules.CommandExecution.CaptureErrors {
		if isError, ok := event.Metadata["is_error"].(bool); ok && isError {
			return true, types.PriorityCritical
		}
	}
	
	// Check ignore patterns first
	for _, regex := range fe.commandIgnoreRegex {
		if regex.MatchString(content) {
			return false, types.PriorityLow
		}
	}
	
	// Check capture patterns
	for _, regex := range fe.commandCaptureRegex {
		if regex.MatchString(content) {
			return true, types.PriorityHigh
		}
	}
	
	// Check output size limits
	lines := strings.Count(content, "\n") + 1
	if lines > fe.rules.CommandExecution.MaxOutputLines {
		// Too many lines, don't capture
		return false, types.PriorityLow
	}
	
	// Default: don't capture routine command output
	return false, types.PriorityLow
}

// shouldCaptureSearch determines if a search operation should be captured
func (fe *FilterEngine) shouldCaptureSearch(event *types.CaptureEvent) (bool, types.Priority) {
	if resultCount, ok := event.Metadata["result_count"].(int); ok {
		if resultCount < fe.rules.SearchOperations.MinResults {
			return false, types.PriorityLow
		}
		
		if resultCount > fe.rules.SearchOperations.MaxResults {
			// Too many results, might not be useful
			return false, types.PriorityLow
		}
		
		// Determine priority based on result count
		if resultCount <= 5 {
			return true, types.PriorityHigh // Precise searches are valuable
		} else if resultCount <= 20 {
			return true, types.PriorityMedium
		} else {
			return true, types.PriorityLow
		}
	}
	
	return true, types.PriorityMedium
}

// checkDebounce implements debouncing for file operations
func (fe *FilterEngine) checkDebounce(source string) bool {
	now := time.Now()
	debounceMs := time.Duration(fe.rules.FileOperations.DebounceMs) * time.Millisecond
	
	// Apply profile multiplier
	if fe.profile != nil {
		debounceMs = time.Duration(float64(debounceMs) * fe.profile.DebounceMultiplier)
	}
	
	// Check if we're within debounce period
	if lastTime, exists := fe.lastFileEvents[source]; exists {
		if now.Sub(lastTime) < debounceMs {
			// Update the timer but don't capture yet
			fe.updateDebounceTimer(source, debounceMs)
			return false
		}
	}
	
	// Update last event time
	fe.lastFileEvents[source] = now
	fe.updateDebounceTimer(source, debounceMs)
	
	return true
}

// updateDebounceTimer updates or creates a debounce timer
func (fe *FilterEngine) updateDebounceTimer(source string, duration time.Duration) {
	// Cancel existing timer
	if timer, exists := fe.debounceTimers[source]; exists {
		timer.Stop()
	}
	
	// Create new timer
	fe.debounceTimers[source] = time.AfterFunc(duration, func() {
		// Clean up after debounce period
		delete(fe.lastFileEvents, source)
		delete(fe.debounceTimers, source)
	})
}

// UpdateRules updates the filter rules
func (fe *FilterEngine) UpdateRules(rules types.FilterRules) {
	fe.rules = rules
	fe.compilePatterns()
}

// UpdateProfile updates the active profile
func (fe *FilterEngine) UpdateProfile(profile *types.Profile) {
	fe.profile = profile
}

// GetStats returns filtering statistics
func (fe *FilterEngine) GetStats() map[string]any {
	return map[string]any{
		"active_debounce_timers":   len(fe.debounceTimers),
		"tracked_file_events":      len(fe.lastFileEvents),
		"command_capture_patterns": len(fe.commandCaptureRegex),
		"command_ignore_patterns":  len(fe.commandIgnoreRegex),
		"file_ignore_patterns":     len(fe.fileIgnorePatterns),
		"file_include_patterns":    len(fe.fileIncludePatterns),
	}
}