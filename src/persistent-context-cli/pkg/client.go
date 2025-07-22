package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/models"
)

// Client represents the HTTP client for interacting with the persistent context web service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new HTTP client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetMemories retrieves all memories from the system
func (c *Client) GetMemories(limit int) ([]*models.MemoryEntry, error) {
	url := fmt.Sprintf("%s/api/v1/journal", c.baseURL)
	if limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", url, limit)
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response models.GetMemoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Memories, nil
}

// GetMemory retrieves a specific memory by ID
func (c *Client) GetMemory(id string) (*models.MemoryEntry, error) {
	url := fmt.Sprintf("%s/api/v1/journal/%s", c.baseURL, id)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var memory models.MemoryEntry
	if err := json.NewDecoder(resp.Body).Decode(&memory); err != nil {
		return nil, fmt.Errorf("failed to decode memory: %w", err)
	}

	return &memory, nil
}

// TriggerConsolidation triggers the consolidation process
func (c *Client) TriggerConsolidation() (*models.ConsolidateResponse, error) {
	url := fmt.Sprintf("%s/api/v1/journal/consolidate", c.baseURL)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("failed to trigger consolidation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response models.ConsolidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetStats retrieves system statistics
func (c *Client) GetStats() (*models.StatsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/journal/stats", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response models.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CheckHealth retrieves service health status
func (c *Client) CheckHealth() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// CheckReady retrieves service readiness status
func (c *Client) CheckReady() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/ready", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check readiness: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}