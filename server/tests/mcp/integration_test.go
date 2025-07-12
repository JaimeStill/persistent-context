package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/JaimeStill/persistent-context/internal/config"
	"github.com/JaimeStill/persistent-context/internal/logger"
	"github.com/JaimeStill/persistent-context/internal/types"
)

func TestMCPIntegration(t *testing.T) {
	// Create test configuration manually
	cfg := config.MCPConfig{
		Name:              "test-mcp",
		Version:           "1.0.0",
		ServerEndpoint:    "http://localhost:8543",
		CaptureMode:       "balanced",
		WorkerCount:       2,
		BatchWindowMs:     1000,
		MaxBatchSize:      5,
		CacheSize:         100,
		PriorityQueueSize: 50,
		BufferSize:        100,
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		Timeout:           30 * time.Second,
		
		FilterRules: types.FilterRules{
			FileOperations: types.FileOperationRules{
				MinChangeSize:   50,  // Higher threshold to match test expectations
				DebounceMs:      500,
				IgnorePatterns:  []string{"*.tmp"},
				IncludePatterns: []string{},
				MaxFileSize:     1024,
			},
			CommandExecution: types.CommandExecutionRules{
				CaptureErrors:   true,
				CapturePatterns: []string{"ERROR", "FAIL"},
				IgnorePatterns:  []string{},
				MaxOutputLines:  100,
			},
			SearchOperations: types.SearchOperationRules{
				MinResults:    1,
				MaxResults:    50,
				BatchWindowMs: 5000,
			},
		},
		
		Profiles: map[string]*types.Profile{
			"balanced": {
				Name:               "balanced",
				DebounceMultiplier: 1.0,
				FilterStrictness:   types.FilterStrictnessMedium,
				CaptureThreshold:   0.5,
			},
		},
	}
	
	// Create test config and logger
	testConfig := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
	log := logger.New(testConfig)
	
	// Create test harness
	harness := NewTestHarness(cfg, log)
	defer harness.Shutdown()
	
	t.Run("FilterTest", func(t *testing.T) {
		results, err := harness.RunFilterTest()
		if err != nil {
			t.Fatalf("Filter test failed: %v", err)
		}
		
		summary := results["summary"].(map[string]any)
		successRate := summary["success_rate"].(float64)
		
		if successRate < 0.75 { // 75% success rate threshold
			t.Errorf("Filter test success rate too low: %.2f%%", successRate*100)
		}
		
		t.Logf("Filter test completed with %.2f%% success rate", successRate*100)
	})
	
	t.Run("LoadTest", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Run a modest load test
		result, err := harness.RunLoadTest(ctx, 5*time.Second, 20) // 20 events/second for 5 seconds
		if err != nil {
			t.Fatalf("Load test failed: %v", err)
		}
		
		// Validate performance metrics
		if result.Throughput < 15 { // Should handle at least 15 events/second
			t.Errorf("Throughput too low: %.2f events/sec", result.Throughput)
		}
		
		if result.AverageLatency > 50*time.Millisecond { // Should be under 50ms average
			t.Errorf("Average latency too high: %v", result.AverageLatency)
		}
		
		if result.EventsFailed > result.EventsGenerated/10 { // Less than 10% failure rate
			t.Errorf("Too many failed events: %d/%d", result.EventsFailed, result.EventsGenerated)
		}
		
		t.Logf("Load test: %.2f events/sec, %v avg latency, %d/%d events processed", 
			result.Throughput, result.AverageLatency, result.EventsProcessed, result.EventsGenerated)
	})
	
	t.Run("EventTypes", func(t *testing.T) {
		// Test different event types
		testEvents := []*types.CaptureEvent{
			{
				Type:    types.EventTypeFileWrite,
				Source:  "test.go",
				Content: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
				Metadata: map[string]any{
					"change_size": 100,
					"file_size":   int64(200),
				},
			},
			{
				Type:    types.EventTypeCommandOutput,
				Source:  "go test",
				Content: "PASS\nok  \ttest\t0.123s",
				Metadata: map[string]any{
					"is_error":  false,
					"exit_code": 0,
				},
			},
			{
				Type:    types.EventTypeSearchResults,
				Source:  "grep -r \"main\"",
				Content: "Found 5 matches",
				Metadata: map[string]any{
					"result_count": 5,
				},
			},
		}
		
		for _, event := range testEvents {
			err := harness.pipeline.ProcessEvent(event)
			if err != nil {
				t.Errorf("Failed to process %s event: %v", event.Type, err)
			}
		}
		
		// Wait for processing
		time.Sleep(2 * time.Second)
		
		metrics := harness.pipeline.GetMetrics()
		if metrics.TotalEvents < int64(len(testEvents)) {
			t.Errorf("Not all events were counted: %d < %d", metrics.TotalEvents, len(testEvents))
		}
		
		t.Logf("Processed %d events, %d filtered, %d failed", 
			metrics.ProcessedEvents, metrics.FilteredEvents, metrics.FailedEvents)
	})
}