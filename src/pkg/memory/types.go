package memory

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

// AssociationType represents different types of memory associations
type AssociationType string

const (
	// AssociationTemporal represents memories that occurred close in time
	AssociationTemporal AssociationType = "temporal"
	
	// AssociationSemantic represents memories with similar content/meaning
	AssociationSemantic AssociationType = "semantic"
	
	// AssociationCausal represents cause-and-effect relationships
	AssociationCausal AssociationType = "causal"
	
	// AssociationContextual represents memories from similar contexts
	AssociationContextual AssociationType = "contextual"
)

// MemoryAssociation represents a relationship between two memories
type MemoryAssociation struct {
	ID         string          `json:"id"`           // Unique association ID
	SourceID   string          `json:"source_id"`    // Source memory ID
	TargetID   string          `json:"target_id"`    // Target memory ID
	Type       AssociationType `json:"type"`         // Type of association
	Strength   float64         `json:"strength"`     // Association strength (0.0-1.0)
	CreatedAt  time.Time       `json:"created_at"`   // When association was created
	UpdatedAt  time.Time       `json:"updated_at"`   // Last update time
	Metadata   map[string]any  `json:"metadata"`     // Additional association data
}

// MemoryScore represents enhanced scoring for memory importance
type MemoryScore struct {
	BaseImportance    float64   `json:"base_importance"`     // Original importance (0.0-1.0)
	DecayFactor       float64   `json:"decay_factor"`        // Time-based decay (0.0-1.0)
	AccessFrequency   int       `json:"access_frequency"`    // Number of access events
	LastAccessed      time.Time `json:"last_accessed"`       // Most recent access time
	RelevanceScore    float64   `json:"relevance_score"`     // Semantic relevance score
	CompositeScore    float64   `json:"composite_score"`     // Final calculated score
}

// MemoryEntry represents a single memory stored in the system
type MemoryEntry struct {
	ID            string            `json:"id"`
	Type          MemoryType        `json:"type"`
	Content       string            `json:"content"`
	Embedding     []float32         `json:"embedding,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	AccessedAt    time.Time         `json:"accessed_at"`
	Strength      float32           `json:"strength"`
	Score         MemoryScore       `json:"score"`               // Enhanced scoring
	AssociationIDs []string         `json:"association_ids"`     // Related memory references
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