# MCP Server Design Document

## Overview

The MCP (Model Context Protocol) server acts as a bridge between Claude Code and the persistent-context memory system. It enables seamless session continuity by capturing relevant interactions and storing them as memories for later retrieval and consolidation.

## MVP Architecture (Post-Review Simplified)

### Deployment Model

1. **Standalone Process**: The MCP server runs as a separate process from the main persistent-context application
2. **Stdio Communication**: Uses stdin/stdout for JSON-RPC communication with Claude Code
3. **HTTP Backend**: Communicates with the journal service via HTTP API for all storage operations

### Communication Flow

```
Claude Code <--stdio--> MCP Server <--HTTP--> Web Service <---> {VectorDB, LLM}
```

### Project Structure (Refactored)

```
persistent-context/
├── cmd/
│   ├── persistent-context-mcp/     # Local MCP binary
│   │   └── main.go
│   └── internal/                   # MCP-specific internals
│       ├── config/                 # MCP configuration
│       └── client/                 # HTTP client to web service
├── web/
│   ├── persistent-context-svc/     # Containerized web service
│   │   └── main.go
│   └── internal/                   # Web-specific internals
│       ├── journal/               # Journal implementation
│       └── http/                  # HTTP server
├── pkg/                           # Shared packages
│   ├── types/                     # Common types
│   ├── config/                    # Shared config
│   └── logger/                    # Shared logging
└── docker-compose.yml             # Web service stack
```

## Primary Use Case: Session Continuity

The MCP server enables seamless transitions between Claude Code sessions by:
1. Capturing context during active coding sessions
2. Storing memories persistently across session boundaries
3. Retrieving relevant context when Claude Code restarts
4. Building cumulative knowledge over time

## Essential MCP Tools (Simplified)

For MVP focus, the MCP server provides 5 essential tools:

1. **`capture_memory`** - Core memory capture functionality
   - Captures context with metadata
   - Stores memories for future retrieval
   - Essential for session continuity

2. **`get_memories`** - Memory retrieval for session continuity
   - Retrieves recent memories
   - Enables context restoration across sessions
   - Supports pagination for large memory sets

3. **`trigger_consolidation`** - Demonstrate memory evolution
   - Triggers consolidation of episodic memories
   - Shows transformation to semantic knowledge
   - Demonstrates learning over time

4. **`get_stats`** - Validation and monitoring
   - Provides memory statistics
   - Validates system health
   - Monitors memory growth and consolidation

5. **`search_memories`** - Enhanced continuity demonstration
   - Searches memories by content similarity
   - Enables context-aware retrieval
   - Supports session-specific queries

## Configuration System (Simplified)

### Essential Configuration

For MVP, configuration focuses on core functionality:

```yaml
# Essential MCP configuration
web_api_url: "http://localhost:8543"
capture_mode: "balanced"
consolidation_threshold: 100
timeout_seconds: 30
```

### Environment Variables

```bash
# Core MCP settings (using APP_ prefix)
APP_MCP_WEB_API_URL=http://localhost:8543
APP_MCP_TIMEOUT=30s
APP_MCP_NAME=persistent-context-mcp
APP_MCP_VERSION=1.0.0
```

## Build and Installation

### Build MCP Binary

```bash
# Build locally
go build -o bin/persistent-context-mcp ./cmd/persistent-context-mcp/

# Install globally
go install ./cmd/persistent-context-mcp/
```

### Start Web Service Stack

```bash
# Start containerized services
docker-compose up -d
```

### Verify Connection

```bash
# Check web service health
curl http://localhost:8543/health

# Test MCP server manually
echo '{"id": "test", "method": "initialize", "params": {}}' | persistent-context-mcp --stdio
```

## Session Continuity Workflow

### Typical Usage Pattern

1. **Start Web Services**: `docker-compose up -d`
2. **Install MCP Server**: `go install ./cmd/persistent-context-mcp/`
3. **Configure Claude Code**: MCP server auto-connects via `.mcp.json`
4. **Work in Session 1**: Context automatically captured
5. **Exit Claude Code**: Memories persist in web service
6. **Restart Claude Code**: Previous context automatically restored
7. **Continue in Session 2**: Build on previous knowledge

## MVP Implementation Focus

For the MVP, the MCP server focuses on core functionality:

- **Simplified Architecture**: Direct HTTP communication with web service
- **Essential Tools Only**: 5 core tools for session continuity
- **Basic Configuration**: Minimal settings for reliable operation
- **Integration Testing**: Manual verification through actual usage

## Future Enhancements

Post-MVP features that may be added:

- **Advanced Filtering**: Configurable capture rules and patterns
- **Multiple Personas**: Support for different memory contexts
- **Performance Optimization**: Async pipelines and caching
- **Cloud Storage**: Remote memory persistence options