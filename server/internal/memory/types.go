package memory

import (
	"context"
	"time"
)

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

// MemoryType represents the different types of memory
type MemoryType string

const (
	TypeEpisodic      MemoryType = "episodic"
	TypeSemantic      MemoryType = "semantic"
	TypeProcedural    MemoryType = "procedural"
	TypeMetacognitive MemoryType = "metacognitive"
)

// MemoryEntry represents a single memory with metadata
type MemoryEntry struct {
	ID         string         `json:"id"`
	Type       MemoryType     `json:"type"`
	Content    string         `json:"content"`
	Embedding  []float32      `json:"embedding,omitempty"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
	AccessedAt time.Time      `json:"accessed_at"`
	Strength   float32        `json:"strength"` // 0.0 to 1.0
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