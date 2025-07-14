package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JaimeStill/persistent-context/pkg/config"
)

// OllamaLLM implements LLM operations using Ollama
type OllamaLLM struct {
	config *config.LLMConfig
	client *http.Client
	cache  map[string][]float32
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response from embedding generation
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// GenerateRequest represents a request for text generation
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResponse represents the response from text generation
type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// NewOllamaLLM creates a new Ollama LLM implementation
func NewOllamaLLM(config *config.LLMConfig) (*OllamaLLM, error) {
	return &OllamaLLM{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		cache: make(map[string][]float32),
	}, nil
}

// GenerateEmbedding generates embeddings for the given text
func (c *OllamaLLM) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Check cache first if enabled
	if c.config.CacheEnabled {
		if embedding, exists := c.cache[text]; exists {
			return embedding, nil
		}
	}

	// Prepare request
	reqBody := EmbeddingRequest{
		Model:  c.config.EmbeddingModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request with retries
	var embedding []float32
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		embedding, lastErr = c.makeEmbeddingRequest(ctx, jsonData)
		if lastErr == nil {
			break
		}

		if attempt < c.config.MaxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
				// Exponential backoff
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
	}

	// Cache the result if enabled
	if c.config.CacheEnabled {
		c.cache[text] = embedding
	}

	return embedding, nil
}

// makeEmbeddingRequest makes a single embedding request
func (c *OllamaLLM) makeEmbeddingRequest(ctx context.Context, jsonData []byte) ([]float32, error) {
	url := c.config.URL + "/api/embeddings"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embeddingResp.Embedding, nil
}

// ConsolidateMemories uses the LLM to consolidate multiple memories into semantic knowledge
func (c *OllamaLLM) ConsolidateMemories(ctx context.Context, memories []string) (string, error) {
	prompt := c.buildConsolidationPrompt(memories)

	reqBody := GenerateRequest{
		Model:  c.config.ConsolidationModel,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request with retries
	var result string
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		result, lastErr = c.makeGenerateRequest(ctx, jsonData)
		if lastErr == nil {
			break
		}

		if attempt < c.config.MaxRetries {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
				// Exponential backoff
			}
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
	}

	return result, nil
}

// makeGenerateRequest makes a single generation request
func (c *OllamaLLM) makeGenerateRequest(ctx context.Context, jsonData []byte) (string, error) {
	url := c.config.URL + "/api/generate"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var generateResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return generateResp.Response, nil
}

// buildConsolidationPrompt creates a prompt for memory consolidation
func (c *OllamaLLM) buildConsolidationPrompt(memories []string) string {
	prompt := "You are a memory consolidation system. Analyze the following episodic memories and extract the key semantic knowledge, patterns, and insights. "
	prompt += "Consolidate them into concise, meaningful knowledge that can be stored as semantic memory.\n\n"
	prompt += "Episodic memories to analyze:\n"

	for i, memory := range memories {
		prompt += fmt.Sprintf("%d. %s\n", i+1, memory)
	}

	prompt += "\nPlease provide a consolidated summary that captures the essential knowledge and patterns from these memories:"

	return prompt
}

// HealthCheck checks if Ollama is accessible
func (c *OllamaLLM) HealthCheck(ctx context.Context) error {
	url := c.config.URL + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// ClearCache clears the embedding cache
func (c *OllamaLLM) ClearCache() {
	if c.config.CacheEnabled {
		c.cache = make(map[string][]float32)
	}
}
