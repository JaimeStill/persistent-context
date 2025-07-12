# Persistent Context

Autonomous memory consolidation for LLMs. Continuously captures context and distills experiences into persistent, hierarchical memories.

## What It Does

Gives LLMs memory that persists between conversations. The system:

- Captures context automatically
- Consolidates memories in the background
- Maintains different types of memories
- Exports/imports memory "personas"

## Quick Start

### Prerequisites

- Docker and Docker Compose
- VS Code with Claude Code extension (for MCP integration)

### 1. Start Services

```bash
docker compose up -d
```

This starts:

- **Qdrant**: Vector database (ports 6333-6334)
- **Ollama**: Local LLM service (port 11434)
- **Web Server**: Memory API (port 8543)

### 2. Verify Services

```bash
# Check all services are healthy
docker compose ps

# Test web server
curl http://localhost:8543/ready
```

### 3. Build MCP Server

```bash
cd server
go build -o bin/mcp ./cmd/mcp/
```

### 4. Connect Claude Code

Connect Claude to the MCP server

```sh
claude mcp add persistent-context /{path-to-repo}/server/bin/mcp
```

Then launch Claude Code.

### 5. Test Integration

Ask Claude Code:

- "Can you capture this conversation as a memory?"
- "What memories do you have from our previous work?"
- "Search for memories related to Docker configuration"

## Architecture

```
Claude Code → MCP Server → Web Server → {Vector DB, LLM}
```

- **MCP Server**: Protocol translation between Claude Code and HTTP API
- **Web Server**: Memory operations, scoring, associations, consolidation
- **Vector DB**: Semantic storage and similarity search
- **LLM**: Embeddings and memory consolidation
