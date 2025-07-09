package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JaimeStill/persistent-context/internal/memory"
)

// MemoryStore implements the MCP MemoryStore interface
type MemoryStore struct {
	// In-memory storage for Session 1 (will be replaced with Qdrant in Session 2)
	episodic []memory.MemoryEntry
	counter  int
}

// NewMemoryStore creates a new memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		episodic: make([]memory.MemoryEntry, 0),
		counter:  0,
	}
}

// CaptureContext implements the MCP interface for capturing context
func (ms *MemoryStore) CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) error {
	ms.counter++
	
	// Create memory entry
	entry := memory.MemoryEntry{
		ID:         fmt.Sprintf("mem_%d", ms.counter),
		Type:       memory.TypeEpisodic,
		Content:    content,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Strength:   1.0, // New memories start with full strength
	}
	
	// Add source to metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}
	entry.Metadata["source"] = source
	
	// Store in episodic memory
	ms.episodic = append(ms.episodic, entry)
	
	log.Printf("Memory captured: source=%s, id=%s, content_length=%d", 
		source, entry.ID, len(content))
	
	return nil
}

// GetMemories retrieves memories (placeholder for Session 2)
func (ms *MemoryStore) GetMemories(ctx context.Context, limit int) ([]memory.MemoryEntry, error) {
	if limit <= 0 || limit > len(ms.episodic) {
		limit = len(ms.episodic)
	}
	
	// Return most recent memories first
	result := make([]memory.MemoryEntry, 0, limit)
	for i := len(ms.episodic) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, ms.episodic[i])
	}
	
	return result, nil
}

// GetMemoryByID retrieves a specific memory by ID
func (ms *MemoryStore) GetMemoryByID(ctx context.Context, id string) (*memory.MemoryEntry, error) {
	for _, entry := range ms.episodic {
		if entry.ID == id {
			// Update access time
			entry.AccessedAt = time.Now()
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("memory with id %s not found", id)
}

// GetMemoryStats returns basic statistics about stored memories
func (ms *MemoryStore) GetMemoryStats() map[string]any {
	totalMemories := len(ms.episodic)
	
	// Calculate average content length
	totalContentLength := 0
	if totalMemories > 0 {
		for _, entry := range ms.episodic {
			totalContentLength += len(entry.Content)
		}
	}
	
	avgContentLength := 0
	if totalMemories > 0 {
		avgContentLength = totalContentLength / totalMemories
	}
	
	// Find oldest and newest memories
	var oldestTime, newestTime time.Time
	if totalMemories > 0 {
		oldestTime = ms.episodic[0].CreatedAt
		newestTime = ms.episodic[0].CreatedAt
		
		for _, entry := range ms.episodic {
			if entry.CreatedAt.Before(oldestTime) {
				oldestTime = entry.CreatedAt
			}
			if entry.CreatedAt.After(newestTime) {
				newestTime = entry.CreatedAt
			}
		}
	}
	
	return map[string]any{
		"total_memories":        totalMemories,
		"episodic_memories":     totalMemories,
		"semantic_memories":     0, // Placeholder for Session 3
		"procedural_memories":   0, // Placeholder for Session 3
		"metacognitive_memories": 0, // Placeholder for Session 3
		"average_content_length": avgContentLength,
		"oldest_memory":         oldestTime.Format(time.RFC3339),
		"newest_memory":         newestTime.Format(time.RFC3339),
	}
}

// HealthCheck implements the HealthChecker interface
func (ms *MemoryStore) HealthCheck(ctx context.Context) error {
	// For in-memory storage, we're always healthy
	// In Session 2, this will check Qdrant connectivity
	return nil
}