# Session 10 Execution Plan: Architecture Refactoring & Memory Enhancement

## Overview

Session 10 focuses on separating the monolithic server into two independent executables (MCP server and web server) to enable Claude Code integration. This requires restructuring the application orchestration logic and updating build standards.

## Session Progress

### 0. Session Start Process (5 minutes) - COMPLETED

- [x] Review latest session handoff from `_context/sessions/session-009.md`
- [x] Review `tasks.md` for current roadmap
- [x] Use plan mode to validate session scope with user
- [x] Create execution-plan.md for Session 10

### 1. Package Restructuring (15 minutes) - IN PROGRESS

**1.1 Create pkg/app directory structure**

- [ ] Create pkg/app directory
- [ ] Move application.go to pkg/app/app.go with base orchestration logic
- [ ] Update import paths and package declarations

**1.2 Create specialized application orchestrators**

- [ ] Create pkg/app/mcp_app.go for MCP-specific orchestration
- [ ] Create pkg/app/web_app.go for web-specific orchestration
- [ ] Implement component filtering logic for each app type

### 2. Create Separate Executables (15 minutes) - PENDING

**2.1 MCP Executable**

- [ ] Create cmd/mcp/main.go using pkg/app/mcp_app.go
- [ ] Implement MCP-specific process lifecycle

**2.2 Web Executable**  

- [ ] Create cmd/web/main.go using pkg/app/web_app.go
- [ ] Implement web-specific process lifecycle

### 3. Configuration & Build System Updates (10 minutes) - PENDING

**3.1 Update Build Standards**

- [ ] Update CLAUDE.md Build Standards section
- [ ] Change from bin/app to bin/web and bin/mcp

**3.2 Docker Configuration**

- [ ] Update Dockerfile for multi-stage builds
- [ ] Simplify Docker compose environment variables
- [ ] Implement consistent default port (8543)

### 4. Integration Testing (5 minutes) - PENDING

- [ ] Test MCP server standalone functionality
- [ ] Test service separation and Docker integration
- [ ] Validate both services start independently

### 5. Session Management and Handoff (5 minutes) - PENDING

- [ ] Update execution plan with final results
- [ ] Archive to `_context/sessions/session-010.md`
- [ ] Clean up execution-plan.md
- [ ] Update tasks.md with accomplishments

## Architecture Design

### Current Monolithic Structure

```
app/
  application.go    # All components initialized
  main.go          # Single executable
```

### Target Multi-Executable Structure

```
pkg/app/
  app.go           # Base application logic (shared)
  mcp_app.go       # MCP server orchestration
  web_app.go       # Web server orchestration

cmd/
  mcp/main.go      # MCP executable
  web/main.go      # Web executable
```

### Component Distribution

- **MCP App**: config, logger, vectorDB, llmClient, journal, mcpServer
- **Web App**: config, logger, vectorDB, llmClient, journal, consolidation, httpServer

## Success Criteria

- Two independent executables build and run successfully
- MCP server communicates with journal via HTTP API
- Web server handles HTTP requests and consolidation
- Architecture ready for Claude Code integration
- Build standards updated for multi-executable pattern

## Session 10 Results

### Current Status

**COMPLETED SUCCESSFULLY**: All Session 10 objectives achieved with enhanced architecture.

### Major Accomplishments

1. **Flexible Application Framework**: Created `internal/app/` with Application interface and Runner for consistent process lifecycle
2. **Service-Specific Applications**: 
   - `internal/app/mcp_app.go` - MCP server with vectorDB, llmClient, journal, mcpServer
   - `internal/app/web_app.go` - Web server with vectorDB, llmClient, journal, consolidation, httpServer
3. **Independent Executables**: 
   - `bin/mcp` and `bin/web` build and run successfully
   - Clean separation of concerns and dependencies
4. **Separate Docker Images**: 
   - `Dockerfile.web` and `Dockerfile.mcp` for independent scaling and deployment
   - Health check dependencies: MCP waits for web server `/ready` endpoint
5. **Configuration Simplification**:
   - Removed unnecessary `MCP.Enabled` and `Consolidation.Enabled` flags
   - Updated default port from 8080 to 8543
   - Clean Docker compose with separate services
6. **Architecture Cleanup**: 
   - Moved app package to `internal/app/` (no need for public API)
   - Removed outdated `server/app/` package
   - Fixed test references and build paths

### Enhanced Features Beyond Original Plan

- **Separate Docker Images**: Web and MCP can be deployed independently
- **Health Check Dependencies**: Proper startup ordering with `/ready` endpoint
- **Configuration Cleanup**: Removed redundant enabled/disabled flags
- **Clean Package Structure**: Everything properly in `internal/` 

### Architecture Ready For

- Claude Code MCP integration (standalone MCP executable)
- Independent scaling of web and MCP services
- Enhanced deployment flexibility with Docker
- Session 11 development priorities

### Issues and Blockers

None - all objectives completed successfully.

### Next Session Priorities

**Session 11 Priority: Enhanced Memory System**
1. Enhanced memory scoring with decay and relevance factors (deferred from Session 9)
2. Memory association tracking system (deferred from Session 9)
3. Test Claude Code integration with standalone MCP server
