# Iterative Knowledge Framework Enhancement: Domain-Based Organization

## Overview

This document describes a major enhancement to the iterative knowledge framework that introduces domain-based organization for improved scalability, discoverability, and learning pathway management. This enhancement addresses the limitations of flat session organization as knowledge repositories grow.

## Problem Statement

### Current Structure Limitations

**Flat Organization Issues**:
```
_knowledge/
├── maps/ (concept maps separated from sessions)
├── sessions/ (mixed domains in sequential numbering)
├── PUPIL.md, backlog.yaml, prompts/ (framework files)
```

**Specific Problems**:
1. **Poor Discoverability**: Hard to find sessions by topic or domain
2. **Mixed Domains**: Sessions covering different knowledge areas are intermixed
3. **Separated Artifacts**: Concept maps are isolated from related learning sessions
4. **Sequential Numbering**: session-001, session-002 provides no domain context
5. **Scaling Issues**: Flat structure becomes unwieldy as content grows
6. **Fragmented Learning Paths**: Related sessions scattered across directory

## Solution: Domain-Based Organization

### Enhanced Directory Structure

```
_knowledge/
├── PUPIL.md                          # Framework-level (unchanged)
├── prompts/                          # Framework-level (unchanged)
│   └── (existing prompt files)
└── domains/                          # NEW: Domain-based organization
    ├── {domain-name}/
    │   ├── domain.yaml               # Concept map (moved from maps/)
    │   ├── README.md                 # Domain overview & session index
    │   └── sessions/
    │       ├── {descriptive-name}.md # Sessions with meaningful names
    │       └── {another-session}.md
    └── {another-domain}/
        ├── domain.yaml
        ├── README.md
        └── sessions/
            └── {session-files}.md
```

### Key Design Principles

1. **Domain Cohesion**: Related knowledge grouped together
2. **Co-located Artifacts**: Concept maps live with their sessions
3. **Descriptive Naming**: Session names indicate content, not sequence
4. **Scalable Architecture**: Easy to add new domains without restructuring
5. **Clear Navigation**: Domain → sessions pathway for learners

## Implementation Guide

### 1. Directory Structure Creation

Create the enhanced structure:

```bash
# Create domains directory
mkdir -p _knowledge/domains

# For each knowledge domain, create:
mkdir -p _knowledge/domains/{domain-name}/sessions
```

### 2. Concept Map Migration

Move and rename concept maps:

```bash
# From: _knowledge/maps/{concept-map}.yaml
# To: _knowledge/domains/{domain-name}/domain.yaml
mv _knowledge/maps/vector-systems.yaml _knowledge/domains/vector-databases/domain.yaml
mv _knowledge/maps/memory-processing.yaml _knowledge/domains/memory-systems/domain.yaml
```

### 3. Session Migration

Move and rename sessions:

```bash
# From: _knowledge/sessions/session-001-descriptive-name.md
# To: _knowledge/domains/{domain}/sessions/{descriptive-name}.md
mv _knowledge/sessions/session-001-vector-fundamentals.md _knowledge/domains/vector-databases/sessions/fundamentals.md
```

### 4. Domain README Creation

Each domain requires a `README.md` with:

```markdown
# {Domain Name} Domain

## Overview
Brief description of the domain and its scope.

## Learning Pathway
### Prerequisites
- Other domains or knowledge required

### Sessions in Order
1. **session-name.md** (duration) - Description
2. **another-session.md** (duration) - Description

### Key Concepts Covered
- List of main concepts in this domain

## Integration Points
### Builds On
- Dependencies on other domains

### Provides To  
- What other domains use this knowledge

## Session Status
- ✅ completed-session.md - Brief status
- 📋 planned-session.md - Planned content
- 🔄 future-session.md - Future consideration
```

### 5. Enhanced Session Frontmatter

Update session metadata to include domain information:

```yaml
---
domain: memory-systems           # NEW: Domain identification
name: consolidation             # NEW: Short name for referencing
title: Memory Consolidation Deep Dive
duration: 45
prerequisites: 
  - domains/vector-databases/sessions/fundamentals  # NEW: Domain paths
  - domains/memory-systems/sessions/processing-pipeline
builds_on: [memory-processing-pipeline, event-driven-processing]
unlocks: [memory-scoring, consolidation-triggers, semantic-memory]
complexity: advanced
---
```

## Framework Template Updates

### Updated init/_knowledge/ Template

```
init/_knowledge/
├── PUPIL.md                  # Enhanced with domain structure
├── domains/                  # NEW: Template domain structure
│   └── example-domain/
│       ├── domain.yaml
│       ├── README.md
│       └── sessions/
│           └── README.md
└── prompts/                  # Enhanced prompts
    ├── new-domain.md         # NEW: Creating new domains
    ├── domain-session.md     # NEW: Creating domain sessions
    └── (existing prompts)
```

### Enhanced Prompts

**New Prompt: new-domain.md**
```markdown
You are helping to create a new knowledge domain within the iterative knowledge framework.

Context: {domain name and brief description}

Create:
1. Domain directory structure
2. domain.yaml concept map
3. Domain README.md with learning pathway
4. Initial session plan

Consider:
- Prerequisites from other domains
- Integration points
- Logical session progression
- Key concepts to cover
```

## Migration Strategy for Existing Repositories

### Backward Compatibility Approach

1. **Gradual Migration**: Maintain old structure during transition
2. **Symbolic Links**: Create links from old paths to new locations
3. **Reference Updates**: Update all internal cross-references
4. **Documentation Updates**: Update PUPIL.md and backlog.yaml

### Migration Script Template

```bash
#!/bin/bash
# Domain-based migration script

# 1. Create domain structure
mkdir -p _knowledge/domains/{domain-name}/sessions

# 2. Move concept maps
mv _knowledge/maps/{map-name}.yaml _knowledge/domains/{domain-name}/domain.yaml

# 3. Move sessions with renaming
mv _knowledge/sessions/session-{N}-{name}.md _knowledge/domains/{domain-name}/sessions/{name}.md

# 4. Update session frontmatter (manual)
# 5. Create domain README (manual)
# 6. Update references (manual)

# 7. Clean up old structure
rm -rf _knowledge/maps
rm -rf _knowledge/sessions
```

## Benefits and Expected Outcomes

### Immediate Benefits

1. **Improved Discoverability**: Find sessions by domain quickly
2. **Co-located Knowledge**: Maps and sessions together
3. **Meaningful Names**: Descriptive session names instead of numbers
4. **Clear Pathways**: Domain READMEs guide learning progression

### Long-term Benefits

1. **Scalability**: Easy to add new domains without restructuring
2. **Maintainability**: Clear boundaries between knowledge areas
3. **Collaboration**: Multiple people can work on different domains
4. **Reusability**: Domains can be referenced across projects

### Learning Experience Improvements

1. **Intuitive Navigation**: "I want to learn about X" → go to domains/X/
2. **Progressive Learning**: Domain prerequisites create clear dependency chains
3. **Focused Sessions**: Each domain maintains coherent scope
4. **Cross-Domain Integration**: Explicit integration points documented

## Example Implementation

### Practical Example: Memory Systems Domain

**Before**:
```
_knowledge/
├── maps/memory-processing.yaml
├── sessions/session-002-memory-pipeline.md
├── sessions/session-003-consolidation.md
└── sessions/session-004-associations.md
```

**After**:
```
_knowledge/domains/memory-systems/
├── domain.yaml                    # Concept map
├── README.md                      # Domain guide
└── sessions/
    ├── processing-pipeline.md     # Renamed from session-002
    ├── consolidation.md           # Renamed from session-003
    └── association-tracking.md    # Renamed from session-004
```

**Domain README Example**:
```markdown
# Memory Systems Domain

## Learning Pathway
1. **processing-pipeline.md** (45 min) - Memory capture and async processing
2. **consolidation.md** (45 min) - Advanced consolidation algorithms  
3. **association-tracking.md** (30 min) - Graph-based relationships

## Prerequisites
- domains/vector-databases/sessions/fundamentals

## Key Concepts
- Episodic vs semantic memory transformation
- Event-driven processing pipelines
- Memory scoring and association algorithms
```

## Conclusion

The domain-based organization enhancement transforms the iterative knowledge framework from a flat, sequential structure to a scalable, intuitive, domain-organized system. This improvement maintains all existing functionality while dramatically improving discoverability, learning pathway clarity, and long-term maintainability.

The enhancement is backward compatible and can be implemented gradually, making it suitable for immediate adoption in existing repositories while becoming the standard structure for new implementations.

This enhancement addresses real user pain points observed in practice and provides a foundation for continued growth and sophistication of knowledge repositories built with the iterative knowledge framework.

## Domain-Based Backlog Management

### Elimination of backlog.yaml

The framework removes the separate `backlog.yaml` file in favor of integrated backlog management within domain structures. This change eliminates redundancy and improves maintainability.

**Rationale**: The `backlog.yaml` file duplicated information already captured in domain concept maps, creating unnecessary maintenance overhead and potential inconsistencies.

### Backlog Integration in domain.yaml

**Backlog Representation**: Concepts with empty `sessions: []` arrays serve as backlog items within each domain.

**Example**:
```yaml
concepts:
  - id: "performance-optimization"
    name: "Vector Database Performance Optimization"
    level: advanced
    description: "Performance tuning and indexing strategies"
    sessions: []  # Empty = backlog item
    
  - id: "fundamentals"
    name: "Vector Database Fundamentals"
    level: foundational
    description: "Understanding high-dimensional data storage"
    sessions: ["fundamentals"]  # Populated = completed
```

### Benefits of Domain-Based Backlogs

1. **Co-located Planning**: Backlog items exist alongside related concepts and sessions
2. **Eliminates Duplication**: Single source of truth for each domain's roadmap
3. **Domain Ownership**: Each domain manages its own future work
4. **Natural Prioritization**: Concept relationships inform priority
5. **Simplified Maintenance**: One fewer file to keep synchronized

### Migration Strategy

1. **Audit existing backlog.yaml**: Identify all concepts not yet captured in domain files
2. **Add missing concepts**: Create appropriate concept entries in relevant domain.yaml files
3. **Verify coverage**: Ensure all backlog items have been migrated
4. **Remove backlog.yaml**: Delete the file after successful migration

## Prompt Infrastructure Updates

### Required Prompt Modifications

The domain-based approach requires updates to several framework prompts:

**`prompts/add-to-backlog.md`**:
- Change from: "Format according to backlog.yaml structure"
- Change to: "Add concept to appropriate domain.yaml concepts section"
- Update AI instructions to suggest domain placement and relationships

**`prompts/new-session.md`**:
- Change from: "from a concept in your backlog"
- Change to: "from a concept with empty sessions in domain.yaml files"

**`prompts/README.md`**:
- Update descriptions to reference domain.yaml files instead of central backlog
- Clarify that add-to-backlog now adds to domain backlogs

### Template Structure Updates

**Remove from all templates**:
```
├── backlog.yaml
```

**Document in templates** that domain.yaml serves dual purpose:
- **Current state**: concepts with sessions populated
- **Planning/backlog**: concepts with empty sessions arrays

## Enhanced Conclusion

The domain-based organization enhancement, combined with integrated backlog management, creates a comprehensive improvement to the iterative knowledge framework:

### Key Improvements

1. **Structural Enhancement**: Domain-based organization improves discoverability and scalability
2. **Eliminated Redundancy**: Removal of backlog.yaml streamlines maintenance
3. **Integrated Planning**: Backlog items co-located with domain knowledge
4. **Updated Workflows**: Prompt infrastructure adapted for domain-based approach

### Implementation Benefits

- **Reduced Complexity**: Fewer files to maintain and synchronize
- **Better Organization**: Natural grouping of related concepts and planning
- **Scalable Architecture**: Domains can be developed independently
- **Maintainable Workflows**: Simplified prompt-based interactions

This comprehensive enhancement maintains backward compatibility while providing a robust foundation for knowledge repository growth and sophistication.