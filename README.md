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
- Go 1.24+ with `$HOME/go/bin` in PATH
- VS Code with Claude Code extension (for MCP integration)

**Setup Go PATH** (if not already configured):

```bash
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**VS Code Terminal Configuration**: This repository includes `.vscode/settings.json` with cross-platform terminal profiles configured to use login shells (bash/zsh with `-l` flag). This ensures that environment variables from your shell profile are properly loaded in VS Code terminals, which is required for Claude Code MCP integration to find globally installed binaries.

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

### 3. Build and Install MCP Server

**Build locally** (creates `bin/persistent-context-mcp`):

```bash
cd src
go build -o ../bin/persistent-context-mcp ./persistent-context-mcp/
```

**Install globally** (creates `$GOPATH/bin/persistent-context-mcp`):

```bash
go install ./persistent-context-mcp
```

### 4. Connect Claude Code

The repository includes `.mcp.json` configuration that references the globally installed `persistent-context-mcp` binary. After installing the MCP server (step 3) and ensuring Go's bin directory is in your PATH, simply launch Claude Code from the repository root directory.

**Environment Troubleshooting**: If Claude Code can't find the `persistent-context-mcp` command:

1. Ensure `$HOME/go/bin` is in your PATH (see Prerequisites)
2. The workspace includes `.vscode/settings.json` that configures terminal profiles to use login shells, ensuring environment variables are properly loaded
3. Restart VS Code after PATH changes to ensure Claude Code inherits the updated environment

### 5. Test MCP Server Manually (Optional)

You can test the MCP server directly before connecting to Claude Code:

```bash
# Test initialize command (using local binary)
echo '{"id": "test", "method": "initialize", "params": {}}' | ./bin/persistent-context-mcp --stdio

# Test tools list (using global binary)
echo '{"id": "test2", "method": "tools/list", "params": {}}' | persistent-context-mcp --stdio
```

Expected responses:

- Initialize: Returns server name and version
- Tools list: Returns 10 available MCP tools

### 6. Test Integration

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
