# Source Description Request

This prompt automates the creation of educational source documentation for complex technical concepts.

## Instructions

When you encounter a complex technical implementation that would benefit from educational documentation, use this prompt:

---

**Request**: Create a source description for the proposed implementation following the educational documentation standards established in CLAUDE.md.

**Requirements**:

1. Store the documentation at `.artifacts/source/source-XXX.md` where XXX is the next sequential number
2. Follow the format established in source-001.md:
   - **Overview**: Simple explanation using analogies and plain language
   - **Architecture Context**: How this fits into the larger system
   - **Function-by-Function Breakdown**: Each component explained with purpose and responsibility
   - **Complete Source Code**: Full implementation with detailed comments
   - **Key Design Patterns**: Important concepts and patterns demonstrated
   - **Learning Points**: Educational value and future enhancement opportunities

**Current Implementation Context**:

- **File**: `[FILE_PATH]`
- **Purpose**: `[BRIEF_DESCRIPTION]`
- **Complexity Level**: `[HIGH/MEDIUM/LOW]`
- **Key Concepts**: `[LIST_KEY_CONCEPTS]`

**Specific Focus Areas**:

- Explain any algorithms or mathematical concepts in simple terms
- Use analogies to make complex systems relatable
- Highlight design decisions and their rationale
- Explain how this integrates with existing architecture
- Provide clear examples of usage patterns

**Output Format**:
Generate comprehensive educational documentation that serves as both a learning resource and technical reference, optimized for consumption by both humans and LLMs.

---

## Usage Example

To request source documentation for the association tracking system:

```
Request: Create a source description for the proposed implementation following the educational documentation standards established in CLAUDE.md.

File: src/pkg/journal/associations.go
Purpose: Memory association tracking system for related memories
Complexity Level: HIGH
Key Concepts: Graph-based relationships, similarity algorithms, temporal analysis, semantic embeddings

Focus on explaining the association algorithms, relationship management, and how this enables intelligent memory connections.
```
