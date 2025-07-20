# Memory Systems Domain

## Overview

Memory systems transform raw conversation context into structured, searchable, and evolving knowledge. This domain covers the complete lifecycle from initial memory capture through consolidation into semantic knowledge, including the sophisticated algorithms that determine memory importance and relationships.

## Learning Pathway

### Prerequisites
- **Required**: `vector-databases/sessions/fundamentals` - Understanding vector storage and similarity
- Basic familiarity with event-driven architectures
- Go interfaces and dependency injection concepts

### Sessions in Order
1. **processing-pipeline.md** (45 min) - How context becomes persistent memory
   - Memory capture through the complete system pipeline
   - Asynchronous processing with Go channels and goroutines
   - Integration between HTTP APIs, LLM services, and vector storage

2. **consolidation.md** (45 min) - Advanced memory consolidation algorithms
   - Event-driven consolidation triggers and memory selection
   - LLM-powered transformation from episodic to semantic memories
   - Context window management and memory scoring algorithms

3. **association-tracking.md** (30 min) - Graph-based memory relationships
   - Four types of memory associations (temporal, semantic, causal, contextual)
   - Bidirectional indexing and strength-based filtering
   - Association-enhanced consolidation and retrieval

### Key Concepts Covered
- **Memory Types**: Episodic, semantic, procedural, and metacognitive
- **Processing Pipeline**: Async event queues and background processing
- **Consolidation**: LLM-powered knowledge extraction and synthesis
- **Association Graphs**: Relationship tracking and strength calculation
- **Scoring Algorithms**: Time decay, frequency, and relevance weighting

## Integration Points

### **Builds On**
- `vector-databases` - Requires understanding of similarity search and embeddings
- `go-patterns` - Uses interfaces, channels, and concurrency patterns

### **Integrates With**
- `integration` - Memory systems expose HTTP APIs for external access

### **Provides To**
- Session 14 development work - Core backend features for completion

## Domain Architecture

```
Context Input → Processing Pipeline → Vector Storage
     ↓              ↓                    ↓
Event Queue → Consolidation → Semantic Memory
     ↓              ↓                    ↓
Associations ← Memory Graph ← Scoring Algorithms
```

## Real-World Analogies

- **Processing Pipeline**: Like a manufacturing assembly line for thoughts
- **Consolidation**: Like how sleep transforms daily experiences into learning
- **Associations**: Like building a knowledge graph in your brain

## Domain Artifacts

- **domain.yaml** - Concept map showing memory processing relationships
- **sessions/** - Progressive learning from basic pipeline to advanced algorithms
- **README.md** - This overview and learning pathway guide

## Session Status

- ✅ processing-pipeline.md - Memory capture and async processing
- ✅ consolidation.md - Advanced consolidation algorithms  
- ✅ association-tracking.md - Graph-based memory relationships
- 🔄 Future sessions - Enhanced scoring, meta-consolidation as needed

Start with `processing-pipeline.md` to understand how your system transforms conversations into searchable memories, then progress through consolidation and association tracking to master the complete memory lifecycle.