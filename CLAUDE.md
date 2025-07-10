# Claude Code

## Repository Restrictions

The only restriction within this repository is that any folder prefixed with a `_` is read-only for you. It is intended for me to take notes and keep track of important context artifacts.

## Project: Autonomous LLM Memory Consolidation System

### Session Constraints

- Each work session has approximately 1 hour of available time
- Tasks should be scoped to fit within this time and context window
- Focus on MVP implementation that can be completed incrementally

### Design Principles

- Keep solutions simple and maintainable
- Use idiomatic approaches that follow best practices
- Avoid overly complicated designs requiring substantial debugging
- Design elegance is incredibly important
- Complex enough to properly implement features, but not overengineered
- Do not use emojis in code, documentation, or communication

### Technical Preferences

- Use Docker containers for third-party services where available (e.g., Ollama, Qdrant)
- Prefer containerization over direct installation for better isolation and portability
- Use `any` instead of `interface{}` in Go code for better readability
- Organize Docker volumes under a consistent root (e.g., `./data/`) for cleaner structure
- Implement proper separation of concerns with dedicated packages for each functionality

### Development Approach

- Break down work into discrete, completable tasks
- Maintain clear documentation of progress and decisions
- Update this file with new directives as they are established

### Session Management and Handoff Process

- **Documentation First**: At the start of each session, archive the previous execution plan to `_context/sessions/session-XXX.md` and create a new execution plan for the current session
- **Continuous Updates**: Update execution-plan.md continuously during development, marking tasks as completed with relevant implementation details
- **Living Document**: The execution plan serves as both a progress tracker and a comprehensive handoff document for the next session
- **Clear Handoff**: Each session should end with a complete execution plan showing what was accomplished, any blockers, and next steps

### Configuration Management

- **Configuration First**: All new components must expose their settings through the config package using Viper's mapstructure tags
- **Sensible Defaults**: Provide reasonable default values for all configuration options
- **Documentation**: Document all configuration options in code comments and execution plans
- **Flexibility**: Design components to be runtime-configurable rather than compile-time where possible
- **Nested Structure**: Use nested configuration structures to organize related settings logically

## Projected Repository Structure

As the project scales, the repository structure should evolve to:

```
persistent-context/
├── server/                 # Core Go server for memory consolidation
│   ├── cmd/               # Main applications
│   ├── internal/          # Private application code
│   └── pkg/               # Public libraries
├── client/                # Future: Web/CLI interface for memory analysis
├── mcp/                   # MCP server implementations
│   ├── file-watcher/
│   ├── git-monitor/
│   └── api-monitor/
├── tools/                 # Utility scripts and tools
│   ├── persona-export/
│   └── memory-analyzer/
├── docker/                # Docker configurations
│   └── docker-compose.yml
├── personas/              # Storage for exported personas
├── docs/                  # Project documentation
├── _context/              # Read-only context artifacts
├── claude.md
├── tasks.md
└── execution-plan.md
```

Create directories only as needed. Currently focusing on `server/` for MVP.
