package persona

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/JaimeStill/persistent-context/internal/types"
)

// Persona represents a snapshot of memory state with metadata
type Persona struct {
	ID          string                 `json:"id"`           // Unique persona identifier
	Name        string                 `json:"name"`         // Human-readable name
	Description string                 `json:"description"`  // Description of the persona
	Version     int                    `json:"version"`      // Version number for tracking changes
	ParentID    string                 `json:"parent_id"`    // ID of parent persona (for versioning)
	CreatedAt   time.Time              `json:"created_at"`   // Creation timestamp
	UpdatedAt   time.Time              `json:"updated_at"`   // Last update timestamp
	Metadata    map[string]any         `json:"metadata"`     // Additional metadata
	MemoryCount int                    `json:"memory_count"` // Number of memories in this persona
	Tags        []string               `json:"tags"`         // Tags for categorization
}

// PersonaExport represents the full export data for a persona
type PersonaExport struct {
	Persona      *Persona                      `json:"persona"`
	Memories     []*types.MemoryEntry          `json:"memories"`
	Associations []*types.MemoryAssociation    `json:"associations"`
	ExportedAt   time.Time                     `json:"exported_at"`
	Format       string                        `json:"format"` // Format version
}

// PersonaManager handles persona operations
type PersonaManager struct {
	personas map[string]*Persona // In-memory storage (could be persisted later)
}

// NewPersonaManager creates a new persona manager
func NewPersonaManager() *PersonaManager {
	return &PersonaManager{
		personas: make(map[string]*Persona),
	}
}

// CreatePersona creates a new persona with given name and description
func (pm *PersonaManager) CreatePersona(name, description string, metadata map[string]any) *Persona {
	persona := &Persona{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    metadata,
		Tags:        []string{},
	}
	
	pm.personas[persona.ID] = persona
	return persona
}

// GetPersona retrieves a persona by ID
func (pm *PersonaManager) GetPersona(id string) (*Persona, error) {
	persona, exists := pm.personas[id]
	if !exists {
		return nil, fmt.Errorf("persona not found: %s", id)
	}
	return persona, nil
}

// ListPersonas returns all personas
func (pm *PersonaManager) ListPersonas() []*Persona {
	personas := make([]*Persona, 0, len(pm.personas))
	for _, p := range pm.personas {
		personas = append(personas, p)
	}
	return personas
}

// UpdatePersona updates persona metadata
func (pm *PersonaManager) UpdatePersona(id string, updates map[string]any) error {
	persona, exists := pm.personas[id]
	if !exists {
		return fmt.Errorf("persona not found: %s", id)
	}
	
	// Update fields
	if name, ok := updates["name"].(string); ok {
		persona.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		persona.Description = desc
	}
	if tags, ok := updates["tags"].([]string); ok {
		persona.Tags = tags
	}
	
	persona.UpdatedAt = time.Now()
	return nil
}

// CreateVersion creates a new version of an existing persona
func (pm *PersonaManager) CreateVersion(parentID string, name, description string) (*Persona, error) {
	parent, exists := pm.personas[parentID]
	if !exists {
		return nil, fmt.Errorf("parent persona not found: %s", parentID)
	}
	
	// Create new version
	newPersona := &Persona{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Version:     parent.Version + 1,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    parent.Metadata, // Inherit metadata
		Tags:        parent.Tags,     // Inherit tags
	}
	
	pm.personas[newPersona.ID] = newPersona
	return newPersona, nil
}

// ExportPersona exports a persona with all associated data
func (pm *PersonaManager) ExportPersona(persona *Persona, memories []*types.MemoryEntry, associations []*types.MemoryAssociation) (*PersonaExport, error) {
	export := &PersonaExport{
		Persona:      persona,
		Memories:     memories,
		Associations: associations,
		ExportedAt:   time.Now(),
		Format:       "1.0", // Version format for compatibility
	}
	
	// Update memory count
	persona.MemoryCount = len(memories)
	
	return export, nil
}

// SerializePersona converts persona export to JSON
func (pm *PersonaManager) SerializePersona(export *PersonaExport) ([]byte, error) {
	return json.MarshalIndent(export, "", "  ")
}

// DeserializePersona converts JSON to persona export
func (pm *PersonaManager) DeserializePersona(data []byte) (*PersonaExport, error) {
	var export PersonaExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to deserialize persona: %w", err)
	}
	return &export, nil
}

// ImportPersona imports a persona from export data
func (pm *PersonaManager) ImportPersona(export *PersonaExport) error {
	// Store the persona
	pm.personas[export.Persona.ID] = export.Persona
	
	// In a real implementation, we would also:
	// 1. Import memories into the vector database
	// 2. Recreate associations
	// 3. Update any references
	
	return nil
}

// GetVersionHistory returns all versions of a persona lineage
func (pm *PersonaManager) GetVersionHistory(personaID string) ([]*Persona, error) {
	persona, exists := pm.personas[personaID]
	if !exists {
		return nil, fmt.Errorf("persona not found: %s", personaID)
	}
	
	// Collect all versions in the lineage
	versions := []*Persona{persona}
	
	// Walk up the parent chain
	currentID := persona.ParentID
	for currentID != "" {
		parent, exists := pm.personas[currentID]
		if !exists {
			break
		}
		versions = append(versions, parent)
		currentID = parent.ParentID
	}
	
	// Walk down to find any children
	for _, p := range pm.personas {
		if p.ParentID == personaID && p.ID != persona.ID {
			versions = append(versions, p)
		}
	}
	
	return versions, nil
}

// ComparePersonas compares two personas and returns differences
func (pm *PersonaManager) ComparePersonas(id1, id2 string) (map[string]any, error) {
	persona1, err := pm.GetPersona(id1)
	if err != nil {
		return nil, err
	}
	
	persona2, err := pm.GetPersona(id2)
	if err != nil {
		return nil, err
	}
	
	diff := map[string]any{
		"persona1": persona1,
		"persona2": persona2,
		"changes": map[string]any{
			"name_changed":        persona1.Name != persona2.Name,
			"description_changed": persona1.Description != persona2.Description,
			"version_diff":        persona2.Version - persona1.Version,
			"time_diff":           persona2.CreatedAt.Sub(persona1.CreatedAt),
		},
	}
	
	return diff, nil
}