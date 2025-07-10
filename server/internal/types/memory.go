package types

import (
	"context"
	"time"
)

// MemoryType represents different types of memories
type MemoryType string

const (
	// TypeEpisodic represents specific experiences and events
	TypeEpisodic MemoryType = "episodic"
	
	// TypeSemantic represents general knowledge and facts
	TypeSemantic MemoryType = "semantic"
	
	// TypeProcedural represents learned skills and procedures
	TypeProcedural MemoryType = "procedural"
	
	// TypeMetacognitive represents knowledge about thinking processes
	TypeMetacognitive MemoryType = "metacognitive"
)

// MemoryEntry represents a single memory stored in the system
type MemoryEntry struct {
	ID         string            `json:"id"`
	Type       MemoryType        `json:"type"`
	Content    string            `json:"content"`
	Embedding  []float32         `json:"embedding,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	AccessedAt time.Time         `json:"accessed_at"`
	Strength   float32           `json:"strength"`
}

// Memory represents the base interface for all memory types
type Memory interface {
	// Store saves a memory entry
	Store(ctx context.Context, content string, metadata map[string]any) error
	
	// Retrieve gets a specific memory by ID
	Retrieve(ctx context.Context, id string) (*MemoryEntry, error)
	
	// Query searches memories based on semantic similarity
	Query(ctx context.Context, query string, limit int) ([]*MemoryEntry, error)
	
	// Transform converts this memory type to another (e.g., episodic to semantic)
	Transform(ctx context.Context, targetType MemoryType) (Memory, error)
}

// EpisodicMemory represents raw, time-based experiences
type EpisodicMemory struct {
	// Stores recent experiences with full context
	entries []*MemoryEntry
}

// SemanticMemory represents abstracted knowledge and concepts
type SemanticMemory struct {
	// Stores facts, concepts, and relationships
	concepts map[string]*Concept
}

// Concept represents a semantic knowledge unit
type Concept struct {
	ID           string
	Name         string
	Definition   string
	Relationships map[string]float32 // Related concepts and their strength
	Examples     []string
	CreatedFrom  []string // IDs of episodic memories this was derived from
}

// ProceduralMemory represents learned patterns and behaviors
type ProceduralMemory struct {
	// Stores behavioral patterns and responses
	patterns []*Pattern
}

// Pattern represents a learned behavioral pattern
type Pattern struct {
	ID          string
	Trigger     string   // What activates this pattern
	Actions     []string // Sequence of actions
	Frequency   int      // How often this pattern has been used
	Success     float32  // Success rate of this pattern
	LastUsed    time.Time
}

// MetacognitiveMemory represents self-reflection and learning strategies
type MetacognitiveMemory struct {
	// Stores insights about own thinking and learning
	insights []*Insight
}

// Insight represents a metacognitive observation
type Insight struct {
	ID          string
	Type        string // "learning_strategy", "mistake_pattern", "improvement"
	Description string
	Evidence    []string // Memory IDs that support this insight
	Confidence  float32
	CreatedAt   time.Time
}