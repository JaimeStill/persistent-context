package http

// Request models for journal API endpoints

// CaptureMemoryRequest represents the request for creating a new memory
type CaptureMemoryRequest struct {
	Source   string         `json:"source" binding:"required"`
	Content  string         `json:"content" binding:"required"`
	Metadata map[string]any `json:"metadata"`
}

// GetMemoriesRequest represents query parameters for retrieving memories
type GetMemoriesRequest struct {
	Limit uint64 `form:"limit"`
}

// SearchMemoriesRequest represents the request for searching memories
type SearchMemoriesRequest struct {
	Content    string `json:"content" binding:"required"`
	MemoryType string `json:"memory_type"`
	Limit      uint64 `json:"limit"`
}

// ConsolidateRequest represents the request for consolidating memories
type ConsolidateRequest struct {
	MemoryIDs []string `json:"memory_ids" binding:"required"`
}