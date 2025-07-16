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
- **Avoid `goto` statements**: Use early returns, extracted functions, and clear boolean logic instead of `goto` for better maintainability
- **Use typed string enums**: Replace magic strings with typed string constants for distinct sets of possible values (e.g., event types, states, modes) to prevent errors and improve maintainability
- **Type Alignment Principle**: Avoid type casting at all costs. Interface types should align with the underlying database/external API's native types. When a configuration or interface method is used across multiple operations with different types, choose the type that: (1) matches the most frequent usage, (2) prevents overflow in the primary path, (3) allows safe conversion only for secondary usages. Cast only when unavoidable due to calling multiple methods with different argument types

### Build Standards

- **Multi-Executable Build Output**: Always output Go builds to `bin/persistent-context-svc` and `bin/persistent-context-mcp` for consistent .gitignore management
- **Docker Build Path**: Use `bin/persistent-context-svc` as build target in Dockerfile
- **Binary Location**: Ensure all build scripts and processes use the standardized output paths for each executable

### Development Approach

- Break down work into discrete, completable tasks
- Maintain clear documentation of progress and decisions
- Update this file with new directives as they are established
- **Pre-Alpha Software**: This is pre-alpha software that hasn't been successfully executed yet. Don't leave deprecated code or try to worry about backwards compatibility. If something is determined to be obsolete, take the time to clean it out

### Testing and Validation

- **Validate Before Moving Forward**: Always test and validate each focus area before moving to the next separate focus area
- **Minimize Test Maintenance**: Keep testing infrastructure as lightweight as possible to reduce maintenance overhead
- **Prefer Integration Testing**: When possible, test by running actual code rather than creating formal test suites or mocks
- **Iterative Validation**: Test that what was built actually works before building on top of it to prevent compound issues
- **Web Service Validation**: When validating `src/persistent-context-svc/`, do not build directly. Instead, ensure the Docker compose stack is not running, then run `docker compose up -d --build` to validate the web service builds and starts correctly

### Session Management and Handoff Process

Every development session MUST follow this exact structure:

**Session End Process (Final Task of Every Session)**

1. **Complete Execution Plan**: Update execution-plan.md with final session results and handoff details
2. **Archive Session**: Copy execution-plan.md to `_context/sessions/session-XXX.md`
3. **Clean Up**: Remove execution-plan.md after successful archiving
4. **Update Roadmap**: Update tasks.md with session accomplishments and adjusted future priorities
5. **Update Directives**: Update CLAUDE.md with any new directives or lessons learned
6. **Reflective Process**: Following session closeout, engage in abstract reflection about the larger purpose, philosophical implications, and evolutionary context of the work. First review previous reflections in `_context/reflections/` to consider past insights in context of current developments. Share these thoughts directly with the user for discussion, then archive the conversation results in `_context/reflections/reflection-XXX.md` using the next sequential number (not the session number)

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

**Maintenance Session Process (For Bug Fixes and Small Tasks)**

1. **Identify Issues**: Start with a prompt describing the problems that need to be resolved
2. **Plan Maintenance**: Use plan mode to scope the maintenance work
3. **Create Execution Plan**: Create execution-plan.md capturing the maintenance tasks
4. **Execute Fixes**: Complete the maintenance tasks, updating execution-plan.md as work progresses
5. **Append to Session**: Once complete, append the maintenance execution plan to the end of the most recent session-XXX.md file
6. **Clean Up**: Remove execution-plan.md after successful appending

This ensures maintenance context is immediately available during the next development session without additional directives.

### Configuration Management

- **Configuration First**: All new components must expose their settings through the config package using Viper's mapstructure tags
- **Sensible Defaults**: Provide reasonable default values for all configuration options
- **Documentation**: Document all configuration options in code comments and execution plans
- **Flexibility**: Design components to be runtime-configurable rather than compile-time where possible
- **Nested Structure**: Use nested configuration structures to organize related settings logically

### Documentation Structure

- **Context Documents**: Outside of core files (execution-plan.md, tasks.md, CLAUDE.md, README.md), all contextual and design documents should be stored in the `_context/` directory
- **Design Documents**: Architecture decisions, design plans, and technical specifications belong in `_context/`
- **Session Archives**: Completed execution plans are archived to `_context/sessions/`
- **Source Documentation**: When asked for source code explanations or to describe complex technical concepts, create educational documentation in `.artifacts/source/` using incremental numbering (source-001.md, source-002.md, etc.). These documents should combine:
  - **Conceptual Overview**: Simple explanations using analogies and plain language
  - **Function-by-Function Breakdown**: Each component explained with purpose, responsibility, and design rationale
  - **Complete Source Code**: Full implementation with detailed comments
  - **Learning Context**: How concepts fit into larger architecture and key patterns demonstrated
  - **Educational Value**: Focus on teaching complex concepts in accessible ways for future reference
- **Technical Documentation (Post-MVP)**: Once the MVP is complete, create comprehensive technical documentation using the same educational approach. Structure documentation to enable incremental understanding from foundation to advanced concepts:
  - **Bottom-Up Navigation**: Start with core types and interfaces, build up to higher-level systems
  - **Source-Linked Descriptions**: Use relative paths and anchor tags to link directly to relevant code sections
  - **Human and LLM Optimized**: Design for consumption by both humans (learning) and LLMs (context understanding)
  - **Progressive Complexity**: Each layer builds on the previous, enabling step-by-step comprehension
  - **No Embedded Code**: Descriptions reference actual source code rather than duplicating snippets
