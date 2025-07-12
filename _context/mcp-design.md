# MCP Server Design Document

## Overview

The MCP (Model Context Protocol) server acts as a bridge between Claude Code and the persistent-context memory system. It captures relevant interactions and stores them as memories for later retrieval and consolidation.

## Architecture

### Deployment Model

1. **Standalone Process**: The MCP server runs as a separate process from the main persistent-context application
2. **Stdio Communication**: Uses stdin/stdout for JSON-RPC communication with Claude Code
3. **HTTP Backend**: Communicates with the journal service via HTTP API for all storage operations

### Communication Flow

```
Claude Code <--stdio--> MCP Server <--HTTP--> Journal API <---> Vector DB
```

## Configuration System

### Hierarchical Configuration

The MCP server supports a hierarchical configuration system that merges settings from multiple sources, with later sources overriding earlier ones:

1. **Built-in Defaults** (Embedded in code)
2. **System Configuration** (`/etc/persistent-context/config.yaml`)
3. **User Configuration** (`~/.config/persistent-context/config.yaml`)
4. **Workspace Configuration** (`.persistent-context/config.yaml` in project root)
5. **Environment Variables** (PC_* prefix)
6. **Command-line Arguments** (Highest priority)

### Configuration File Structure

```yaml
# Main configuration
mcp:
  server_endpoint: "http://localhost:8080"
  capture_mode: "balanced"                  # Can reference a profile
  profiles_dir: "~/.config/persistent-context/profiles"  # Additional profiles location
  
  # Include other configuration files
  includes:
    - "~/.config/persistent-context/mcp-custom.yaml"
    - ".persistent-context/mcp-local.yaml"

# Persona configuration
persona:
  storage_type: "local"                     # local, s3, gcs, azure
  local:
    base_path: "~/.local/share/persistent-context/personas"
    active_persona: "default"
    auto_export: true
    export_format: "parquet"                # parquet, sqlite, json
    
  # Future: Cloud storage configuration
  # s3:
  #   bucket: "my-personas"
  #   region: "us-west-2"
  #   prefix: "personas/"
```

### Profile System

#### Profile Locations

Profiles are loaded from multiple locations in order:

1. **Built-in Profiles** (Embedded in binary)
2. **System Profiles** (`/etc/persistent-context/profiles/*.yaml`)
3. **User Profiles** (`~/.config/persistent-context/profiles/*.yaml`)
4. **Workspace Profiles** (`.persistent-context/profiles/*.yaml`)

#### Profile File Format

```yaml
# ~/.config/persistent-context/profiles/research-mode.yaml
name: "research-mode"
description: "Optimized for code exploration and research"
base: "balanced"                          # Inherit from another profile

# Override specific settings
debounce_multiplier: 0.5                 # Faster captures during research
filter_rules:
  file_operations:
    min_change_size: 10                   # Capture smaller changes
    include_patterns:
      - "**/*.md"                         # Always capture documentation
      - "**/*.go"
      - "**/*.py"
      - "**/*.js"
      - "**/*.ts"
    
  command_execution:
    capture_patterns:
      - ".*"                              # Capture all command output
    max_output_lines: 10000               # Higher limit for research
    
  search_operations:
    max_results: 100                      # More results during exploration
    batch_window_ms: 60000                # Longer window for research sessions

# Custom scoring weights for this profile
scoring_weights:
  recency: 0.3
  frequency: 0.2
  relevance: 0.5                          # Higher relevance weight for research
```

## Capture Strategy

### Event Types

1. **File Operations**
   - `file_read`: Reading file contents
   - `file_write`: Creating or modifying files
   - `file_delete`: Removing files
   
2. **Command Execution**
   - `command_run`: Bash command execution
   - `command_output`: Command results (filtered)
   
3. **Search Operations**
   - `search_query`: Grep/find operations
   - `search_results`: Relevant findings

### Capture Filtering System

The capture filtering system is designed to be highly flexible and configurable through profiles and rules.

#### Filter Rules Structure

```yaml
filter_rules:
  file_operations:
    min_change_size: 50          # Minimum lines changed to trigger capture
    debounce_ms: 2000           # Quiet period before capture
    ignore_patterns:            # Glob patterns to ignore
      - "*.tmp"
      - "*.log"
      - "node_modules/**"
      - ".git/**"
    include_patterns:           # Explicit include patterns (override ignores)
      - "*.md"
      - "*.go"
    max_file_size: 5242880      # 5MB default
    
  command_execution:
    capture_errors: true        # Always capture error outputs
    capture_patterns:           # Regex patterns to capture
      - "^(ERROR|FAIL|PANIC)"
      - "test.*failed"
      - "build.*error"
    ignore_patterns:            # Regex patterns to ignore
      - "^\\+"                  # Git diff additions
      - "^-"                    # Git diff deletions
    max_output_lines: 5000      # Default 5000 lines
    
  search_operations:
    min_results: 1              # Minimum results to capture
    max_results: 50             # Maximum results to capture
    batch_window_ms: 30000      # Group searches within window
```

### Built-in Capture Mode Profiles

1. **Conservative** (Default)
   ```yaml
   name: "conservative"
   debounce_multiplier: 2.5
   filter_strictness: "high"
   capture_threshold: 0.8
   ```

2. **Balanced**
   ```yaml
   name: "balanced"
   debounce_multiplier: 1.0
   filter_strictness: "medium"
   capture_threshold: 0.5
   ```

3. **Aggressive**
   ```yaml
   name: "aggressive"
   debounce_multiplier: 0.25
   filter_strictness: "low"
   capture_threshold: 0.2
   ```

## Persona Management

### Local Storage Structure

```
~/.local/share/persistent-context/personas/
├── default/
│   ├── metadata.yaml           # Persona metadata
│   ├── memories.parquet       # Vector embeddings and content
│   ├── journal.db             # SQLite for structured data
│   └── checkpoints/           # Version snapshots
│       ├── 2024-01-15_123456.tar.gz
│       └── 2024-01-16_091234.tar.gz
├── work/
│   ├── metadata.yaml
│   ├── memories.parquet
│   └── journal.db
└── research/
    ├── metadata.yaml
    ├── memories.parquet
    └── journal.db
```

### Persona Metadata Format

```yaml
# metadata.yaml
name: "default"
created_at: "2024-01-15T12:34:56Z"
updated_at: "2024-01-16T09:12:34Z"
version: "1.0.0"
description: "Default persona for general development"

statistics:
  total_memories: 15234
  total_size_bytes: 52428800
  last_consolidation: "2024-01-16T08:00:00Z"
  
configuration:
  capture_mode: "balanced"
  consolidation_threshold: 1000
  vector_dimensions: 3072
```

## Performance Optimizations

### Async Pipeline

- Non-blocking capture queue
- Background processing workers
- Configurable worker pool size

### Batching

- Group related captures within time windows
- Batch HTTP requests to journal API
- Combine similar operations

### Caching

- Cache generated embeddings
- Reuse embeddings for similar content
- LRU cache with configurable size

### Priority Queue

- High priority: Errors, test failures
- Medium priority: File edits, search results
- Low priority: Routine reads, status checks

## Default Configuration

```yaml
# Built-in defaults (lowest priority)
mcp:
  server_endpoint: "http://localhost:8080"
  capture_mode: "balanced"
  
  # Performance settings
  batch_window_ms: 5000
  max_batch_size: 10
  cache_size: 1000
  priority_queue_size: 100
  worker_count: 4
  
  # Default filter rules
  filter_rules:
    file_operations:
      min_change_size: 50
      debounce_ms: 2000
      ignore_patterns:
        - "*.tmp"
        - "*.log"
        - "node_modules/**"
        - ".git/**"
        - "bin/**"
        - "data/**"
      max_file_size: 5242880              # 5MB
      
    command_execution:
      capture_errors: true
      capture_patterns:
        - "^(ERROR|FAIL|PANIC)"
        - "test.*failed"
        - "build.*error"
      max_output_lines: 5000
      
    search_operations:
      min_results: 1
      max_results: 50
      batch_window_ms: 30000
```

## Implementation Plan

1. **Phase 1**: Basic capture pipeline with configurable filtering
2. **Phase 2**: Hierarchical configuration system with file includes
3. **Phase 3**: Profile system with inheritance and overrides
4. **Phase 4**: Persona management with local storage
5. **Phase 5**: Performance optimization and cloud storage support

## Performance Targets

- Capture latency: <100ms @ 95th percentile
- Memory overhead: <50MB for MCP server process
- Throughput: 100 captures/minute sustained
- Embedding cache hit rate: >80%
- Zero blocking of Claude Code operations
- Configuration load time: <50ms including all files