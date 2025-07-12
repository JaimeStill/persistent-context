# Claude Code

## Repository Restrictions

- Any folder prefixed with a `_` is read-only for you. Unless explicitly directed by me or one of your directives in this document, you are only allowed to read these files.
- Any folder prefixed with `.` is private and not accessible by you. Unless explicitly directed by me or one of your directives in this document, you are not allowed to access or modify these files.

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
- When removing code/functions, do not leave orphaned comments - comments should only exist when attached to actual code

### Build Standards

- **Consistent Build Output**: Always output Go builds to `bin/app` for consistent .gitignore management
- **Docker Build Path**: Use `bin/app` as the build target in Dockerfiles
- **Binary Location**: Ensure all build scripts and processes use the standardized `bin/app` output path

### Development Approach

- Break down work into discrete, completable tasks
- Maintain clear documentation of progress and decisions
- Update this file with new directives as they are established

### Session Management and Handoff Process

Every development session MUST follow this exact structure:

**Session End Process (Final Task of Every Session)**

1. **Complete Execution Plan**: Update execution-plan.md with final session results and handoff details
2. **Archive Session**: Copy execution-plan.md to `_context/sessions/session-XXX.md`
3. **Clean Up**: Remove execution-plan.md after successful archiving
4. **Update Roadmap**: Update tasks.md with session accomplishments and adjusted future priorities
5. **Update Directives**: Update CLAUDE.md with any new directives or lessons learned

**Session Start Process (First Task of Every Session)**

1. **Review Context**: Read latest `_context/sessions/session-XXX.md` for handoff, `_context/` for full project details, and `tasks.md` for current roadmap
2. **Plan Session**: Use plan mode to brainstorm and validate session scope and goals
3. **Write Execution Plan**: Create new execution-plan.md for current session and begin work

**During Session: Continuous Updates**

- Update execution-plan.md continuously during development
- Mark tasks as completed with relevant implementation details
- Document any blockers or issues discovered
- The execution plan serves as both progress tracker and handoff document

This documentation flow is MANDATORY for every session and takes precedence over all other tasks.

### Configuration Management

- **Configuration First**: All new components must expose their settings through the config package using Viper's mapstructure tags
- **Sensible Defaults**: Provide reasonable default values for all configuration options
- **Documentation**: Document all configuration options in code comments and execution plans
- **Flexibility**: Design components to be runtime-configurable rather than compile-time where possible
- **Nested Structure**: Use nested configuration structures to organize related settings logically
