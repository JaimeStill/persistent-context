package mcp

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/mcp"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// TestHarness provides performance testing for the MCP pipeline
type TestHarness struct {
	pipeline *mcp.ProcessingPipeline
	logger   *logger.Logger
	config   config.MCPConfig
}

// TestResult contains the results of a performance test
type TestResult struct {
	Duration        time.Duration `json:"duration"`
	EventsGenerated int           `json:"events_generated"`
	EventsProcessed int           `json:"events_processed"`
	EventsFiltered  int           `json:"events_filtered"`
	EventsFailed    int           `json:"events_failed"`
	Throughput      float64       `json:"throughput"`       // events per second
	AverageLatency  time.Duration `json:"average_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	MinLatency      time.Duration `json:"min_latency"`
}

// EventTemplate defines a template for generating test events
type EventTemplate struct {
	Type     types.EventType
	Source   string
	Content  string
	Metadata map[string]any
}

// NewTestHarness creates a new test harness
func NewTestHarness(cfg config.MCPConfig, log *logger.Logger) *TestHarness {
	// Create a filter for testing
	profile := cfg.Profiles[cfg.CaptureMode]
	if profile == nil {
		profile = cfg.Profiles["balanced"]
	}
	filter := mcp.NewFilterEngine(cfg.FilterRules, profile)
	
	// Create pipeline for testing
	pipeline := mcp.NewProcessingPipeline(cfg, filter, log)
	
	return &TestHarness{
		pipeline: pipeline,
		logger:   log,
		config:   cfg,
	}
}

// RunLoadTest runs a load test with the specified parameters
func (th *TestHarness) RunLoadTest(ctx context.Context, duration time.Duration, eventsPerSecond int) (*TestResult, error) {
	th.logger.Info("Starting MCP load test",
		"duration", duration,
		"events_per_second", eventsPerSecond)
	
	// Test setup
	startTime := time.Now()
	endTime := startTime.Add(duration)
	eventInterval := time.Second / time.Duration(eventsPerSecond)
	
	var (
		eventsGenerated int
		mu              sync.Mutex
		latencies       []time.Duration
	)
	
	// Event templates for testing
	templates := []EventTemplate{
		{
			Type:    types.EventTypeFileWrite,
			Source:  "test_file.go",
			Content: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			Metadata: map[string]any{
				"change_size": 100,
				"file_size":   int64(512),
			},
		},
		{
			Type:    types.EventTypeCommandOutput,
			Source:  "go build",
			Content: "go: building test package\nBuild successful",
			Metadata: map[string]any{
				"is_error":    false,
				"exit_code":   0,
			},
		},
		{
			Type:    types.EventTypeSearchResults,
			Source:  "grep -r \"func\"",
			Content: "Found 15 matches in 8 files",
			Metadata: map[string]any{
				"result_count": 15,
			},
		},
		{
			Type:    types.EventTypeCommandOutput,
			Source:  "npm test",
			Content: "ERROR: Test failed\nAssertion error: expected true, got false",
			Metadata: map[string]any{
				"is_error":  true,
				"exit_code": 1,
			},
		},
	}
	
	// Start event generation
	ticker := time.NewTicker(eventInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			th.logger.Info("Load test cancelled")
			return nil, ctx.Err()
			
		case <-ticker.C:
			if time.Now().After(endTime) {
				goto testComplete
			}
			
			// Generate and process event
			template := templates[rand.Intn(len(templates))]
			event := &types.CaptureEvent{
				Type:      template.Type,
				Source:    fmt.Sprintf("%s_%d", template.Source, eventsGenerated),
				Content:   template.Content,
				Metadata:  template.Metadata,
				Timestamp: time.Now(),
			}
			
			// Measure latency
			eventStart := time.Now()
			err := th.pipeline.ProcessEvent(event)
			latency := time.Since(eventStart)
			
			mu.Lock()
			eventsGenerated++
			if err == nil {
				latencies = append(latencies, latency)
			}
			mu.Unlock()
			
			if err != nil {
				th.logger.Debug("Event processing failed", "error", err)
			}
		}
	}
	
testComplete:
	// Wait a bit for pipeline to finish processing
	time.Sleep(time.Millisecond * 500)
	
	// Collect final metrics
	metrics := th.pipeline.GetMetrics()
	totalDuration := time.Since(startTime)
	
	// Calculate latency statistics
	var avgLatency, maxLatency, minLatency time.Duration
	if len(latencies) > 0 {
		var totalLatency time.Duration
		maxLatency = latencies[0]
		minLatency = latencies[0]
		
		for _, lat := range latencies {
			totalLatency += lat
			if lat > maxLatency {
				maxLatency = lat
			}
			if lat < minLatency {
				minLatency = lat
			}
		}
		avgLatency = totalLatency / time.Duration(len(latencies))
	}
	
	result := &TestResult{
		Duration:        totalDuration,
		EventsGenerated: eventsGenerated,
		EventsProcessed: int(metrics.ProcessedEvents),
		EventsFiltered:  int(metrics.FilteredEvents),
		EventsFailed:    int(metrics.FailedEvents),
		Throughput:      float64(eventsGenerated) / totalDuration.Seconds(),
		AverageLatency:  avgLatency,
		MaxLatency:      maxLatency,
		MinLatency:      minLatency,
	}
	
	th.logger.Info("Load test completed",
		"duration", result.Duration,
		"events_generated", result.EventsGenerated,
		"events_processed", result.EventsProcessed,
		"throughput", fmt.Sprintf("%.2f events/sec", result.Throughput),
		"avg_latency", result.AverageLatency)
	
	return result, nil
}

// RunFilterTest tests the filtering engine with various event types
func (th *TestHarness) RunFilterTest() (map[string]any, error) {
	th.logger.Info("Running filter test")
	
	// Test events with expected filter results
	testCases := []struct {
		name           string
		event          *types.CaptureEvent
		shouldCapture  bool
		expectedPriority types.Priority
	}{
		{
			name: "large_file_write",
			event: &types.CaptureEvent{
				Type:   types.EventTypeFileWrite,
				Source: "important.go",
				Content: "package main\n// Large file content...",
				Metadata: map[string]any{
					"change_size": 300,  // Must be > MinChangeSize*5 = 250 for PriorityHigh
					"file_size":   int64(1024),
				},
			},
			shouldCapture:    true,
			expectedPriority: types.PriorityHigh,
		},
		{
			name: "small_file_write",
			event: &types.CaptureEvent{
				Type:   types.EventTypeFileWrite,
				Source: "small.go",
				Content: "package main",
				Metadata: map[string]any{
					"change_size": 10,
					"file_size":   int64(50),
				},
			},
			shouldCapture:    false,
			expectedPriority: types.PriorityLow,
		},
		{
			name: "error_command",
			event: &types.CaptureEvent{
				Type:   types.EventTypeCommandOutput,
				Source: "go build",
				Content: "ERROR: compilation failed",
				Metadata: map[string]any{
					"is_error":  true,
					"exit_code": 1,
				},
			},
			shouldCapture:    true,
			expectedPriority: types.PriorityCritical,
		},
		{
			name: "ignored_file",
			event: &types.CaptureEvent{
				Type:   types.EventTypeFileWrite,
				Source: "temp.tmp",
				Content: "temporary content",
				Metadata: map[string]any{
					"change_size": 100,
					"file_size":   int64(200),
				},
			},
			shouldCapture:    false,
			expectedPriority: types.PriorityLow,
		},
	}
	
	results := make(map[string]any)
	var passed, failed int
	
	// We need to access the filter from the pipeline - let's add a getter method
	// For now, we'll create a separate filter for testing
	profile := th.config.Profiles[th.config.CaptureMode]
	if profile == nil {
		profile = th.config.Profiles["balanced"]
	}
	filter := mcp.NewFilterEngine(th.config.FilterRules, profile)
	
	for _, tc := range testCases {
		shouldCapture, priority := filter.ShouldCapture(tc.event)
		
		testPassed := shouldCapture == tc.shouldCapture
		if shouldCapture && tc.shouldCapture {
			testPassed = testPassed && priority == tc.expectedPriority
		}
		
		if testPassed {
			passed++
		} else {
			failed++
			th.logger.Warn("Filter test failed",
				"test", tc.name,
				"expected_capture", tc.shouldCapture,
				"actual_capture", shouldCapture,
				"expected_priority", tc.expectedPriority,
				"actual_priority", priority)
		}
		
		results[tc.name] = map[string]any{
			"passed":           testPassed,
			"should_capture":   shouldCapture,
			"priority":         priority,
			"expected_capture": tc.shouldCapture,
			"expected_priority": tc.expectedPriority,
		}
	}
	
	results["summary"] = map[string]any{
		"total_tests": len(testCases),
		"passed":      passed,
		"failed":      failed,
		"success_rate": float64(passed) / float64(len(testCases)),
	}
	
	th.logger.Info("Filter test completed",
		"passed", passed,
		"failed", failed,
		"success_rate", fmt.Sprintf("%.2f%%", 100.0*float64(passed)/float64(len(testCases))))
	
	return results, nil
}

// Shutdown shuts down the test harness
func (th *TestHarness) Shutdown() error {
	return th.pipeline.Shutdown()
}