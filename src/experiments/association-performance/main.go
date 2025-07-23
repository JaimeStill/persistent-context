package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TestMemory represents our test data structure
type TestMemory struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Source    string         `json:"source"`
	Type      string         `json:"type"`
	Size      string         `json:"size"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// CaptureRequest matches our service API
type CaptureRequest struct {
	Source   string         `json:"source"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

// CaptureResponse matches our service API
type CaptureResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
}

// PerformanceResult stores timing measurements
type PerformanceResult struct {
	MemoryID      string        `json:"memory_id"`
	ContentSize   int           `json:"content_size_bytes"`
	CaptureTime   time.Duration `json:"capture_time_ms"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

func main() {
	fmt.Println("=== Association-Based Memory Performance Test ===")
	fmt.Println("Testing memory capture speed WITHOUT LLM processing")
	fmt.Println("Target: < 200ms per memory capture with associations")
	
	// Load test data
	memories, err := loadTestData("../test-data/dataset_medium.json")
	if err != nil {
		fmt.Printf("Error loading test data: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded %d test memories\n", len(memories))
	
	// Test different batch sizes to find performance characteristics
	batchSizes := []int{1, 5, 10, 20, 50}
	
	for _, batchSize := range batchSizes {
		if batchSize > len(memories) {
			continue
		}
		
		fmt.Printf("\n--- Testing %d memories ---\n", batchSize)
		batch := memories[:batchSize]
		
		results := testMemoryCapture(batch)
		analyzeResults(results, batchSize)
		
		// Small delay between batches
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println("\n=== Summary ===")
	fmt.Println("This test validates that pure association-based memory")
	fmt.Println("can achieve < 200ms capture times without LLM processing.")
}

func loadTestData(filename string) ([]TestMemory, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var memories []TestMemory
	err = json.Unmarshal(data, &memories)
	return memories, err
}

func testMemoryCapture(memories []TestMemory) []PerformanceResult {
	results := make([]PerformanceResult, len(memories))
	
	for i, mem := range memories {
		start := time.Now()
		
		// Make capture request to our service (this will build associations)
		reqBody := CaptureRequest{
			Source:   mem.Source,
			Content:  mem.Content,
			Metadata: mem.Metadata,
		}
		
		jsonData, _ := json.Marshal(reqBody)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8543/api/v1/journal", bytes.NewBuffer(jsonData))
		if err != nil {
			results[i] = PerformanceResult{
				MemoryID:    mem.ID,
				ContentSize: len(mem.Content),
				CaptureTime: time.Since(start),
				Success:     false,
				Error:       fmt.Sprintf("request creation failed: %v", err),
				Timestamp:   time.Now(),
			}
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			results[i] = PerformanceResult{
				MemoryID:    mem.ID,
				ContentSize: len(mem.Content),
				CaptureTime: time.Since(start),
				Success:     false,
				Error:       fmt.Sprintf("request failed: %v", err),
				Timestamp:   time.Now(),
			}
			continue
		}
		defer resp.Body.Close()
		
		respBody, _ := io.ReadAll(resp.Body)
		captureTime := time.Since(start)
		
		success := resp.StatusCode == 200 || resp.StatusCode == 201
		var errorMsg string
		var capturedID string
		
		if success {
			var captureResp CaptureResponse
			if err := json.Unmarshal(respBody, &captureResp); err == nil {
				capturedID = captureResp.ID
			}
		} else {
			errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			if len(respBody) > 0 && len(respBody) < 200 {
				errorMsg += ": " + string(respBody)
			}
		}
		
		results[i] = PerformanceResult{
			MemoryID:    capturedID,
			ContentSize: len(mem.Content),
			CaptureTime: captureTime,
			Success:     success,
			Error:       errorMsg,
			Timestamp:   time.Now(),
		}
		
		// Progress indicator
		if success {
			fmt.Printf("Memory %d: ✓ %s (Size: %d bytes)\n", 
				i+1, captureTime.Truncate(time.Millisecond), len(mem.Content))
		} else {
			fmt.Printf("Memory %d: ✗ %s\n", i+1, errorMsg)
		}
	}
	
	return results
}

func analyzeResults(results []PerformanceResult, batchSize int) {
	var totalTime time.Duration
	var successCount int
	var minTime, maxTime time.Duration
	var totalBytes int
	
	minTime = time.Hour // Initialize to large value
	
	for _, result := range results {
		if result.Success {
			successCount++
			totalTime += result.CaptureTime
			totalBytes += result.ContentSize
			
			if result.CaptureTime < minTime {
				minTime = result.CaptureTime
			}
			if result.CaptureTime > maxTime {
				maxTime = result.CaptureTime
			}
		}
	}
	
	fmt.Printf("\n--- Batch Results ---\n")
	fmt.Printf("Success Rate: %d/%d (%.1f%%)\n", successCount, len(results), 
		float64(successCount)/float64(len(results))*100)
	
	if successCount > 0 {
		avgTime := totalTime / time.Duration(successCount)
		fmt.Printf("Average Time: %s\n", avgTime.Truncate(time.Millisecond))
		fmt.Printf("Min Time: %s\n", minTime.Truncate(time.Millisecond))
		fmt.Printf("Max Time: %s\n", maxTime.Truncate(time.Millisecond))
		fmt.Printf("Total Data: %d KB\n", totalBytes/1024)
		fmt.Printf("Throughput: %.1f memories/second\n", 
			float64(successCount)/totalTime.Seconds())
		
		// Check if we meet target
		if avgTime < 200*time.Millisecond {
			fmt.Printf("✅ MEETS TARGET: < 200ms per memory\n")
		} else {
			fmt.Printf("❌ EXCEEDS TARGET: > 200ms per memory\n")
		}
		
		// Check for scalability
		memoriesPerHour := float64(successCount) / totalTime.Hours()
		fmt.Printf("Projected Rate: %.0f memories/hour\n", memoriesPerHour)
		
		if memoriesPerHour > 1000 {
			fmt.Printf("✅ SCALABLE: Can handle large Claude Code sessions\n")
		} else {
			fmt.Printf("⚠️  LIMITED: May struggle with intensive sessions\n")
		}
	}
}