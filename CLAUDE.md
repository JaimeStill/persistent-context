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

### Build Standards

- **Consistent Build Output**: Always output Go builds to `bin/server` for consistent .gitignore management
- **Docker Build Path**: Use `bin/server` as the build target in Dockerfiles
- **Binary Location**: Ensure all build scripts and processes use the standardized `bin/server` output path

### Development Approach

- Break down work into discrete, completable tasks
- Maintain clear documentation of progress and decisions
- Update this file with new directives as they are established

### Session Management and Handoff Process

Every development session MUST follow this exact structure:

**Part 0: Documentation Setup (First Task of Every Session)**

- Archive current execution-plan.md to `_context/sessions/session-XXX.md`
- Create new execution-plan.md with current session agenda and progress tracking
- Update CLAUDE.md with any new directives or improvements

**During Session: Continuous Updates**

- Update execution-plan.md continuously during development
- Mark tasks as completed with relevant implementation details
- Document any blockers or issues discovered
- The execution plan serves as both progress tracker and handoff document

**Part N: Documentation Cleanup (Final Task of Every Session)**

- Update execution-plan.md with session results and accomplishments
- Update tasks.md with any new tasks or issues discovered
- Note improvements and next steps for future sessions
- Ensure clean handoff state with complete documentation

This documentation flow is MANDATORY for every session and takes precedence over all other tasks.

### Configuration Management

- **Configuration First**: All new components must expose their settings through the config package using Viper's mapstructure tags
- **Sensible Defaults**: Provide reasonable default values for all configuration options
- **Documentation**: Document all configuration options in code comments and execution plans
- **Flexibility**: Design components to be runtime-configurable rather than compile-time where possible
- **Nested Structure**: Use nested configuration structures to organize related settings logically
