# Persistent Context Architecture Documentation

## Overview

The Persistent Context system is an autonomous LLM memory consolidation service built with Go, designed to capture, process, and consolidate contextual information from various sources. The architecture implements a multi-layered service-oriented design with comprehensive configuration management, event-driven consolidation, and middleware-based processing pipelines.

## Overall Service Behavior

### Core Function

The service operates as a continuous memory consolidation system that:

1. **Captures Context**: Receives memory inputs through MCP server or HTTP API
2. **Processes Memories**: Applies middleware pipeline for validation, enrichment, and routing
3. **Stores Vectorized Data**: Generates embeddings and stores in vector database collections
4. **Monitors Thresholds**: Tracks memory count, context usage, and embedding size
5. **Triggers Consolidation**: Event-driven consolidation when thresholds are exceeded
6. **Manages Lifecycle**: Transforms episodic memories into semantic knowledge over time

### Service Lifecycle

```
Initialization → Runtime Monitoring → Event Processing → Graceful Shutdown
      ↓              ↓                    ↓               ↓
   Dependency     Health Checks      Consolidation    Resource Cleanup
   Resolution     Memory Capture     Context Safety   Service Shutdown
```

### Memory Types and Flow

- **Episodic**: Raw time-based experiences → stored immediately
- **Semantic**: Abstracted knowledge ← consolidated from episodic memories
- **Procedural**: Learned patterns ← derived from repeated semantic patterns
- **Metacognitive**: Self-reflection ← analysis of learning and memory patterns

## Configuration System

The system uses Viper-based configuration with hierarchical loading from environment variables, YAML files, and sensible defaults.

### Core Configuration Structure

#### HTTP Server Configuration (`server.*`)

```yaml
server:
  port: "8080"                    # HTTP server port
  read_timeout: 10                # Request read timeout (seconds)
  write_timeout: 10               # Response write timeout (seconds)
  shutdown_timeout: 30            # Graceful shutdown timeout (seconds)
```

#### Logging Configuration (`logging.*`)

```yaml
logging:
  level: "info"                   # Log level: debug, info, warn, error
  format: "json"                  # Log format: json, text
```

#### Vector Database Configuration (`vectordb.*`)

```yaml
vectordb:
  provider: "qdrant"              # Database provider
  url: "http://qdrant:6333"       # Connection URL
  vector_dimension: 1536          # Embedding vector dimension
  on_disk_payload: true           # Use disk storage for payloads
  timeout: "30s"                  # Connection timeout
  collection_names:
    episodic: "episodic_memories"
    semantic: "semantic_memories"
    procedural: "procedural_memories"
    metacognitive: "metacognitive_memories"
```

#### LLM Configuration (`llm.*`)

```yaml
llm:
  provider: "ollama"              # LLM provider
  url: "http://ollama:11434"      # Service URL
  embedding_model: "phi3:mini"    # Model for embeddings
  consolidation_model: "phi3:mini" # Model for consolidation
  cache_enabled: true             # Enable embedding cache
  cache_ttl: "1h"                 # Cache time-to-live
  timeout: "30s"                  # Request timeout
  max_retries: 3                  # Maximum retry attempts
```

#### Memory Configuration (`memory.*`)

```yaml
memory:
  batch_size: 100                 # Processing batch size
  retention_days: 30              # Memory retention period
  consolidation_interval: "6h"    # Consolidation frequency
  max_memory_size: 10000          # Maximum memory entries
  strength_threshold: 0.1         # Minimum memory strength
```

#### MCP Configuration (`mcp.*`)

```yaml
mcp:
  enabled: false                  # Enable MCP server
  name: "persistent-context"      # MCP server name
  version: "1.0.0"               # MCP server version
  buffer_size: 1000              # Context buffer size
  worker_count: 2                # Worker goroutines
  retry_attempts: 3              # Max retry attempts
  retry_delay: "1s"              # Retry delay
  timeout: "30s"                 # Processing timeout
```

#### Consolidation Configuration (`consolidation.*`)

```yaml
consolidation:
  max_tokens: 128000             # Maximum context window size
  safety_margin: 0.7             # Context window safety margin (70%)
  memory_count_threshold: 50     # Memory count trigger
  embedding_size_threshold: 1048576 # Embedding size trigger (1MB)
  context_usage_threshold: 0.8   # Context usage trigger (80%)
  decay_factor: 0.01             # Time decay factor
  access_weight: 2.0             # Access frequency weight
  relevance_weight: 1.5          # Semantic relevance weight
  enabled: true                  # Enable consolidation
```

## Core Service Components

### 1. Type System (`internal/types/`)

**Purpose**: Foundation type definitions and interfaces

**Key Components**:

- `MemoryType`: Enumeration for memory classification
- `MemoryEntry`: Core memory structure with content, embedding, and metadata
- `Memory`: Interface for memory operations (Store, Retrieve, Query, Transform)
- Specialized memory types: `EpisodicMemory`, `SemanticMemory`, `ProceduralMemory`, `MetacognitiveMemory`

**Integration Pattern**: Used throughout the system as the canonical type definitions

### 2. Memory Store (`internal/memory/`)

**Purpose**: Central memory management with vectorDB and LLM integration

**Key Operations**:

```go
// Core memory operations
CaptureContext(ctx context.Context, content string, metadata map[string]any) (*types.MemoryEntry, error)
QuerySimilarMemories(ctx context.Context, query string, memoryType types.MemoryType, limit int) ([]*types.MemoryEntry, error)
ConsolidateMemories(ctx context.Context, memories []*types.MemoryEntry) (*types.MemoryEntry, error)
BatchStoreMemories(ctx context.Context, memories []*types.MemoryEntry) error
```

**Dependencies**: VectorDB interface, LLM interface, MemoryConfig

**Behavior**:

- Generates embeddings for all content using LLM service
- Routes memories to appropriate vector collections by type
- Implements batch processing for efficiency
- Handles consolidation through LLM-powered memory transformation

### 3. Vector Database Layer (`internal/vectordb/`)

**Architecture**: Interface-based abstraction with Qdrant implementation

**Interface Contract**:

```go
type VectorDB interface {
    Initialize(ctx context.Context) error
    Store(ctx context.Context, memory *types.MemoryEntry) error
    Query(ctx context.Context, vector []float32, collection string, limit int) ([]*types.MemoryEntry, error)
    Retrieve(ctx context.Context, id string, collection string) (*types.MemoryEntry, error)
    HealthCheck(ctx context.Context) error
}
```

**Qdrant Implementation Features**:

- Collection management per memory type
- Vector similarity search with configurable dimensions
- Payload storage with rich metadata support
- Connection pooling and health monitoring

### 4. LLM Layer (`internal/llm/`)

**Architecture**: Interface-based abstraction with Ollama implementation

**Interface Contract**:

```go
type LLM interface {
    GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
    ConsolidateMemories(ctx context.Context, memories []*types.MemoryEntry) (*types.MemoryEntry, error)
    HealthCheck(ctx context.Context) error
}
```

**Ollama Implementation Features**:

- HTTP client with connection pooling
- Embedding generation with TTL-based caching
- Memory consolidation through structured prompting
- Retry logic with exponential backoff
- Health monitoring with service availability checks

### 5. Consolidation Engine (`internal/consolidation/`)

**Purpose**: Event-driven memory consolidation with context window safety

**Core Components**:

**Context Monitor**:

```go
type ContextMonitor struct {
    maxTokens         int
    safetyMargin     float32
    currentUsage     int
    estimatedCost    int
}
```

**Event System**:

```go
type ConsolidationEvent struct {
    Type         EventType  // ContextInit, NewContext, ThresholdReached, ConversationEnd
    Trigger      string
    Memories     []Memory
    ContextState ContextState
    Timestamp    time.Time
}
```

**Memory Scoring**:

```go
type MemoryScore struct {
    AccessCount       int
    LastAccessed      time.Time
    SemanticRelevance float32
    DecayFactor       float32
    TotalScore        float32
}
```

**Behavior**:

- Monitors context window usage with configurable safety margins
- Triggers consolidation based on memory count, embedding size, or context usage
- Implements importance-based memory selection for consolidation
- Ensures consolidation never exceeds context window limits

### 6. Application Services (`app/services/`)

**Architecture**: Service wrapper pattern with lifecycle management

**Base Service Pattern**:

```go
type Service interface {
    Name() string
    Dependencies() []string
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck(ctx context.Context) error
}
```

**Service Implementations**:

**Memory Service**:

- Wraps memory store with middleware pipeline integration
- Provides `ProcessMemory()` method for pipeline-based processing
- Manages memory store lifecycle and dependencies

**VectorDB Service**:

- Wraps VectorDB interface with service lifecycle
- No dependencies - foundation service

**LLM Service**:

- Wraps LLM interface with service lifecycle  
- No dependencies - foundation service

**HTTP Service**:

- Manages HTTP server with dependency injection for health checks
- Depends on VectorDB and LLM for comprehensive health reporting

**MCP Service**:

- Conditional service based on configuration
- Depends on Memory service for context capture

### 7. Middleware Pipeline (`app/middleware/`)

**Architecture**: Extensible processing pipeline with context passing

**Pipeline Structure**:

```go
type Pipeline struct {
    stages []MiddlewareFunc
    logger *slog.Logger
}

type MiddlewareFunc func(ctx context.Context, memCtx *MemoryContext, next func(context.Context, *MemoryContext) error) error

type MemoryContext struct {
    Memory    *types.MemoryEntry
    Source    string
    Stage     string
    Metadata  map[string]any
    StartTime time.Time
}
```

**Middleware Components**:

**Validation Middleware**: Input validation and sanitization
**Enrichment Middleware**: Metadata enrichment and timestamp addition
**Consolidation Middleware**: Consolidation event triggering
**Logging Middleware**: Structured logging of processing stages

**Integration**: Used by Memory Service for all memory processing operations

### 8. Application Orchestrator (`app/orchestrator.go`)

**Purpose**: Central application coordinator and dependency injection manager

**Key Responsibilities**:

```go
type Orchestrator struct {
    registry *lifecycle.Registry
    config   *config.Config
    logger   *logger.Logger
    services map[string]services.Service
}
```

**Orchestration Flow**:

1. **Service Registration**: All services registered with dependency declarations
2. **Dependency Resolution**: Topological sort determines initialization order
3. **Base Service Initialization**: Foundation services (VectorDB, LLM) initialized first
4. **Dependency Injection**: Complex services receive their dependencies
5. **Service Startup**: All services started in dependency order
6. **Health Monitoring**: Continuous health checks across all services
7. **Graceful Shutdown**: Services stopped in reverse dependency order

**Dependency Injection Pattern**:

```go
func (o *Orchestrator) injectDependencies() error {
    // Get foundation services
    vectordbService := o.services["vectordb"].(*services.VectorDBService)
    llmService := o.services["llm"].(*services.LLMService)
    
    // Inject into dependent services
    memoryService := o.services["memory"].(*services.MemoryService)
    memoryService.InitializeWithDependencies(vectordbService.Client(), llmService.Client())
    
    // Continue injection chain...
}
```

## Integration Patterns and Data Flow

### Memory Processing Flow

```
Input (MCP/HTTP) → Middleware Pipeline → Memory Store → Vector Database
       ↓                    ↓                ↓              ↓
   Validation         Enrichment       Embedding        Storage
   Logging           Consolidation     Generation       Retrieval
```

### Event-Driven Consolidation Flow

```
Memory Operations → Threshold Monitoring → Event Generation → Consolidation
       ↓                    ↓                   ↓               ↓
   Count/Size           Context Safety      Event Queue     LLM Processing
   Tracking             Margin Check        Processing      Memory Transform
```

### Service Dependency Graph

```
VectorDB Service (foundation)
LLM Service (foundation)
    ↓
Memory Service (depends: VectorDB, LLM)
    ↓
HTTP Service (depends: VectorDB, LLM)
MCP Service (depends: Memory)
```

### Configuration Loading Hierarchy

```
Environment Variables (APP_*) → YAML Files → Package Defaults → Validation
```

## Key Design Patterns

### 1. Interface-Based Architecture

- VectorDB and LLM interfaces enable provider swapping
- Service interfaces enable testing and mocking
- Clear contracts between components

### 2. Dependency Injection

- Manual dependency injection in orchestrator
- Interface-based dependencies for flexibility
- Service-specific initialization methods

### 3. Middleware Pattern

- Extensible memory processing pipeline
- Context-aware processing with metadata propagation
- Configurable middleware chains

### 4. Event-Driven Processing

- Consolidation triggered by configurable events
- Queue-based event processing with safety checks
- Context-aware event handling

### 5. Service-Oriented Architecture

- Independent service lifecycle management
- Comprehensive health monitoring
- Graceful shutdown with dependency awareness

This architecture provides a robust, scalable foundation for autonomous memory consolidation with clear separation of concerns, comprehensive configuration management, and extensive monitoring capabilities.
