# Memory System Experiments

## Overview

This repository contains standalone experiments to evaluate two memory system approaches for the persistent-context project. These experiments serve as focused prototypes to test performance, scalability, and effectiveness before committing to a full architectural implementation.

## Background

The persistent-context project requires a memory system that can:

- Capture and evolve context across Claude Code sessions
- Store mixed content types (code, documentation, conversations)
- Form associations between related memories
- Provide fast, relevant context retrieval for autonomous LLM usage
- Scale to handle months/years of development history

Previous attempts with local LLM consolidation proved unviable due to performance bottlenecks (14+ second processing times). These experiments focus on association-based memory systems without LLM consolidation.

## Experimental Approaches

### 1. SQLite FTS5 + BM25 Approach

- **Storage**: SQLite with Full-Text Search extension
- **Relevance**: BM25 scoring algorithm
- **Associations**: Text-based similarity and metadata relationships
- **Dependencies**: Zero external services
- **Target**: Self-hosted, zero-cost solution

### 2. Voyage AI + Qdrant Approach  

- **Storage**: Qdrant vector database
- **Embeddings**: Voyage AI API (voyage-3-lite model)
- **Associations**: Vector similarity and semantic relationships
- **Dependencies**: External API service, Docker containers
- **Target**: High-quality semantic understanding

## Memory System Model

Both experiments implement the same cognitive memory model:

### Memory Types

- **Episodic**: Specific events and interactions ("Fixed auth bug in session.go")
- **Semantic**: General knowledge and patterns ("JWT tokens expire after 1 hour")  
- **Procedural**: How-to knowledge ("Steps to debug connection timeouts")
- **Metacognitive**: Self-awareness about processes ("I work better with smaller functions")

### Association Formation

- **Temporal**: Memories from the same cognitive time period
- **Semantic**: Memories about similar topics/concepts
- **Causal**: Memories where one led to another
- **Contextual**: Memories from the same source/project area

### Cognitive Time Model

Unlike wall-clock time, cognitive time advances based on token processing:

- Input tokens processed by the LLM
- Output tokens generated by the LLM (weighted higher)
- Interaction cycles and context switches

Memory decay and reinforcement operate on this cognitive timeline.

## Test Scenarios

### Scale Testing

- **Small**: 100 memories (single session)
- **Medium**: 1,000 memories (week of work)  
- **Large**: 10,000 memories (months of history)
- **Extreme**: 100,000 memories (years of development)

### Usage Patterns

- **Morning Context**: Loading relevant memories from previous work
- **Deep Focus**: Repeated queries about specific subsystems
- **Context Switching**: Moving between unrelated features  
- **Problem Solving**: Finding similar patterns and solutions

### Memory Content

Realistic development artifacts with natural interconnections:

- Code implementation and modifications
- Architecture discussions and decisions
- Debugging sessions and discoveries
- Planning conversations and strategy notes
- Documentation updates and insights

## Success Criteria

### Performance Requirements

- **Memory Storage**: < 200ms per memory capture
- **Query Response**: < 1 second for real-time assistance
- **Association Formation**: Automatic and efficient
- **Scale Handling**: Graceful degradation up to context limits

### Quality Metrics  

- **Retrieval Relevance**: Memories match query intent
- **Association Accuracy**: Related memories properly connected
- **Context Preservation**: Important information survives over time
- **Memory Evolution**: System adapts and improves with usage

## Repository Structure

```
memory/
├── readme.md                    # This file
├── shared/                      # Core memory package used by both experiments
│   ├── memory/                  # Memory types and interfaces
│   ├── associations/            # Association formation algorithms  
│   ├── decay/                   # Memory decay and reinforcement
│   └── simulation/              # Test data generation and scenarios
├── sqlite/                      # SQLite FTS5 + BM25 implementation
│   ├── README.md                # SQLite experiment roadmap
│   ├── storage/                 # SQLite-specific storage layer
│   └── cmd/                     # CLI application
├── voyage/                      # Voyage AI + Qdrant implementation  
│   ├── readme.md                # Voyage AI experiment roadmap
│   ├── storage/                 # Qdrant + Voyage AI storage layer
│   ├── docker-compose.yml       # Qdrant container setup
│   └── cmd/                     # CLI application
└── results/                     # Experiment results and comparisons
    ├── performance/             # Benchmark results
    ├── quality/                 # Retrieval quality assessments
    └── analysis/                # Comparative analysis and conclusions
```

## Decision Framework

After completing both experiments, the architectural decision will be based on:

1. **Performance**: Which approach meets real-time requirements?
2. **Quality**: Which provides better context retrieval and associations?  
3. **Reliability**: Which is more stable and consistent?
4. **Operational Simplicity**: Preference for zero-dependency solutions when performance is comparable
5. **Scale Characteristics**: How do they handle growing memory bases?

## Next Steps

1. Implement the shared memory package with core abstractions
2. Build SQLite experiment following its architectural roadmap
3. Build Voyage AI experiment following its architectural roadmap  
4. Execute comparative testing across all scenarios
5. Analyze results and make architectural decision for persistent-context integration

The winning approach will become the foundation for the persistent-context memory system, enabling Claude Code sessions with continuous, evolving contextual memory.
