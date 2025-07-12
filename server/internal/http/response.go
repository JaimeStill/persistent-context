package http

import (
	"github.com/JaimeStill/persistent-context/internal/types"
)

// Response models for journal API endpoints

// CaptureMemoryResponse represents the response after capturing a memory
type CaptureMemoryResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// GetMemoriesResponse represents the response for memory retrieval
type GetMemoriesResponse struct {
	Memories []*types.MemoryEntry `json:"memories"`
	Count    int                  `json:"count"`
	Limit    uint64               `json:"limit"`
}

// SearchMemoriesResponse represents the response for memory search
type SearchMemoriesResponse struct {
	Memories []*types.MemoryEntry `json:"memories"`
	Query    string               `json:"query"`
	Count    int                  `json:"count"`
	Limit    uint64               `json:"limit"`
}

// ConsolidateResponse represents the response after memory consolidation
type ConsolidateResponse struct {
	Message        string `json:"message"`
	ProcessedCount int    `json:"processed_count"`
}

// StatsResponse represents the response for memory statistics
type StatsResponse struct {
	Stats map[string]any `json:"stats"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}