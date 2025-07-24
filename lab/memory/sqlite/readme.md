# SQLite FTS5 Experiment - Architectural Roadmap

## Experiment Objective

Evaluate SQLite FTS5 with BM25 scoring as a zero-dependency memory system that can handle associative memory formation, cognitive time-based decay, and fast context retrieval without external services.

## Architectural Foundation

### Core Hypothesis

SQLite's mature FTS5 extension with built-in BM25 ranking can provide sufficient search quality and performance for memory retrieval while maintaining complete autonomy and minimal operational overhead.

### Key Technical Bets

- **FTS5 Performance**: Sub-50ms queries even with 100k+ memories
- **JSON Flexibility**: Metadata and associations stored as JSON columns
- **BM25 Relevance**: Built-in scoring sufficient for memory ranking
- **Association Storage**: Graph relationships via relational joins
- **Cognitive Time Indexing**: Custom time-based decay using SQLite functions

## Implementation Phases

### Phase 1: Foundation Setup

**Goal**: Establish core SQLite schema and basic memory operations

**Key Components**:

- Database schema design for memories, associations, and metadata
- Basic CRUD operations using the shared memory interface
- FTS5 table configuration with custom tokenizers if needed
- Connection pooling and transaction management

**Success Criteria**:

- Can store and retrieve individual memories
- FTS5 search returns ranked results
- Database handles concurrent operations safely

**Questions to Resolve**:

- Optimal FTS5 configuration parameters
- JSON schema for memory metadata and associations
- Transaction boundaries for memory + association creation

### Phase 2: Association Formation

**Goal**: Implement automatic association detection and storage

**Key Components**:

- Text similarity algorithms for semantic associations
- Time-proximity detection for temporal associations  
- Reference parsing for causal associations
- Bidirectional association storage and querying

**Success Criteria**:

- New memories automatically form associations with existing ones
- Association queries return related memories efficiently
- Association strength scoring works intuitively

**Questions to Resolve**:

- How to efficiently compute similarity without embeddings
- Optimal association storage schema (separate table vs JSON)
- Association strength calculation and thresholds

### Phase 3: Cognitive Time and Decay

**Goal**: Implement token-based time advancement and memory decay

**Key Components**:

- Cognitive time tracking and advancement logic
- Memory strength decay algorithms using SQL functions
- Frequency-based reinforcement on memory access
- Cleanup of extremely weak memories

**Success Criteria**:

- Memory strength decreases over cognitive time
- Frequently accessed memories maintain strength
- Query results naturally prefer recent/strong memories
- Database size remains manageable through decay

**Questions to Resolve**:

- Decay curve parameters and tuning
- When and how to perform memory cleanup
- Balancing decay speed vs memory preservation

### Phase 4: Query Optimization

**Goal**: Achieve target performance for real-time memory retrieval

**Key Components**:

- Query optimization and index tuning
- Result ranking that combines FTS5 scores with memory strength
- Caching strategies for frequent queries
- Performance monitoring and bottleneck identification

**Success Criteria**:

- Queries complete in <1 second consistently  
- Results relevance feels natural and useful
- Performance scales reasonably with memory count
- Memory usage remains acceptable

**Questions to Resolve**:

- Custom ranking formulas combining BM25 + memory strength
- Caching strategy without external dependencies
- Index maintenance overhead vs query speed

### Phase 5: Scale Testing and Tuning

**Goal**: Validate performance and behavior at realistic scales

**Key Components**:

- Large-scale memory simulation and loading
- Performance benchmarking across different memory counts
- Memory pressure testing and degradation analysis
- Parameter tuning based on observed behavior

**Success Criteria**:

- System handles 10k+ memories without major degradation
- Memory associations remain meaningful at scale
- Decay and cleanup maintain reasonable database size
- Performance characteristics are predictable

**Questions to Resolve**:

- Scale limits and degradation patterns
- Optimal decay parameters for different usage patterns
- Maintenance procedures for long-running systems

## Technical Architecture

### Database Schema Approach

```sql
-- Core memories table with FTS5
CREATE VIRTUAL TABLE memories_fts USING fts5(
    content, metadata, tokenize='porter'
);

-- Memory metadata and associations
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- episodic, semantic, procedural, metacognitive  
    strength REAL NOT NULL DEFAULT 1.0,
    cognitive_time INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    last_accessed INTEGER NOT NULL,
    metadata JSON,
    content_id TEXT -- Reference to FTS5 table
);

-- Association relationships
CREATE TABLE associations (
    id TEXT PRIMARY KEY,
    from_memory_id TEXT NOT NULL,
    to_memory_id TEXT NOT NULL,
    association_type TEXT NOT NULL, -- temporal, semantic, causal, contextual
    strength REAL NOT NULL,
    created_cognitive_time INTEGER NOT NULL,
    FOREIGN KEY (from_memory_id) REFERENCES memories(id),
    FOREIGN KEY (to_memory_id) REFERENCES memories(id)
);
```

### Query Strategy

- **Primary Search**: FTS5 for content matching with BM25 scoring
- **Memory Filtering**: Combine FTS5 results with memory strength and recency
- **Association Expansion**: Follow associations from primary results
- **Result Ranking**: Custom scoring combining search relevance + memory strength

### Association Formation Strategy

- **Text Similarity**: n-gram overlap, common keywords, TF-IDF without vectors
- **Temporal Proximity**: Memories within similar cognitive time windows
- **Reference Detection**: Parse content for explicit references to other memories
- **Contextual Grouping**: Memories from same files, sessions, or topics

## CLI Interface Design

### Basic Operations

```bash
# Store new memory
sqlite-memory store --type=episodic --content="Implemented JWT auth middleware"

# Query memories  
sqlite-memory query "authentication implementation"

# Show associations
sqlite-memory associations <memory-id>

# Advance cognitive time
sqlite-memory advance-time --input-tokens=1500 --output-tokens=800

# Performance testing
sqlite-memory benchmark --memories=10000 --queries=1000
```

### Simulation Commands

```bash
# Load test dataset
sqlite-memory simulate --scenario=auth-development --memories=500

# Run scale test
sqlite-memory scale-test --max-memories=100000 --query-samples=100

# Memory analysis
sqlite-memory analyze --decay-report --association-graph
```

## Evaluation Criteria

### Performance Benchmarks

- **Memory Storage**: < 200ms including association formation
- **Query Response**: < 1s for typical queries, < 5s for complex association queries
- **Database Size**: Reasonable growth patterns with effective decay
- **Concurrent Access**: Handle multiple simultaneous operations

### Quality Assessments  

- **Search Relevance**: Manual evaluation of query result quality
- **Association Accuracy**: Do related memories connect appropriately?
- **Memory Evolution**: Does the system improve over time?
- **Context Preservation**: Important memories survive appropriately

### Operational Characteristics

- **Zero Dependencies**: No external services or complex setup
- **Resource Usage**: Memory and CPU consumption patterns
- **Reliability**: Error handling and data integrity
- **Maintainability**: Schema migrations and system updates

## Risk Assessment

### Potential Challenges

- **Text-Only Similarity**: Limited semantic understanding without embeddings
- **BM25 Limitations**: May not capture nuanced relevance for code contexts
- **Association Quality**: Text-based associations might miss semantic relationships
- **Scale Limitations**: SQLite performance with very large datasets

### Mitigation Strategies

- **Enhanced Text Processing**: Custom tokenizers, stemming, domain-specific parsing
- **Hybrid Scoring**: Combine multiple similarity signals beyond just text
- **Incremental Optimization**: Start simple, add complexity based on observed needs
- **Performance Monitoring**: Early detection of scale-related issues

## Success Metrics

### Minimum Viable Performance

- Store 1000 memories in < 200s total
- Query typical memory set in < 1s
- Form meaningful associations between related memories
- Handle realistic coding session simulation

### Ideal Performance Targets

- Sub-100ms memory storage including associations
- Sub-500ms query response for complex searches
- Maintain performance up to 50k+ memories
- Intuitive and accurate memory associations

## Next Phase Dependencies

This experiment's results will inform:

- Comparative analysis against Voyage AI approach
- Integration strategy for persistent-context project
- Performance optimization priorities
- Operational deployment requirements

The implementation should focus on establishing baseline functionality quickly, then iterating based on observed performance and quality characteristics.
