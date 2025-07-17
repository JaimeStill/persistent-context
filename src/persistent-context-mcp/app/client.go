package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/models"
)

// Client provides HTTP client functionality for the journal API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new HTTP client for the journal API
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// CaptureContext captures context via HTTP API
func (c *Client) CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) (*models.MemoryEntry, error) {
	req := models.CaptureMemoryRequest{
		Source:   source,
		Content:  content,
		Metadata: metadata,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/journal", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("request failed: %s - %s", errResp.Error, errResp.Message)
	}

	var captureResp models.CaptureMemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&captureResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return a minimal memory entry with the ID
	return &models.MemoryEntry{
		ID: captureResp.ID,
	}, nil
}

// GetMemories retrieves memories via HTTP API
func (c *Client) GetMemories(ctx context.Context, limit uint32) ([]*models.MemoryEntry, error) {
	url := fmt.Sprintf("%s/api/v1/journal?limit=%d", c.baseURL, limit)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var getResp models.GetMemoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return getResp.Memories, nil
}


// QuerySimilarMemories searches for similar memories via HTTP API
func (c *Client) QuerySimilarMemories(ctx context.Context, content string, memoryType models.MemoryType, limit uint64) ([]*models.MemoryEntry, error) {
	req := models.SearchMemoriesRequest{
		Content:    content,
		MemoryType: string(memoryType),
		Limit:      limit,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/journal/search", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var searchResp models.SearchMemoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResp.Memories, nil
}

// ConsolidateMemories triggers memory consolidation via HTTP API
func (c *Client) ConsolidateMemories(ctx context.Context, memories []*models.MemoryEntry) error {
	memoryIDs := make([]string, len(memories))
	for i, memory := range memories {
		memoryIDs[i] = memory.ID
	}

	req := models.ConsolidateRequest{
		MemoryIDs: memoryIDs,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/journal/consolidate", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed: %s - %s", errResp.Error, errResp.Message)
	}

	return nil
}

// GetMemoryStats retrieves memory statistics via HTTP API
func (c *Client) GetMemoryStats(ctx context.Context) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v1/journal/stats", c.baseURL)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var statsResp models.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return statsResp.Stats, nil
}

// HealthCheck checks if the web server is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/ready", c.baseURL)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}


