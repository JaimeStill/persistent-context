# Prompt: Generate New Learning Session

Use this prompt to create a new learning session from a concept with empty sessions in domain.yaml files or a newly identified knowledge gap.

## Prompt Template

```
Generate a learning session for [CONCEPT NAME] following the Iterative Knowledge framework.

Context:
- My experience level: [From PUPIL.md or specific context]
- Why I need this: [Current project need or curiosity]
- Related concepts I understand: [List relevant prior knowledge]
- Time available: [30, 45, or 60 minutes]

Session Requirements:
- Use the session.md template structure
- Include practical, runnable examples
- Connect to my existing knowledge
- Provide real-world applications
- Include comprehension checkpoints

Specific Focus:
[Any particular aspect you want emphasized]

Environment:
[Your development environment constraints]
```

## Example Usage

```
Generate a learning session for "vector databases" following the Iterative Knowledge framework.

Context:
- My experience level: Comfortable with traditional databases, new to vector concepts
- Why I need this: Building a memory system that needs similarity search
- Related concepts I understand: SQL databases, key-value stores, basic embeddings
- Time available: 45 minutes

Session Requirements:
- Use the session.md template structure
- Include practical, runnable examples
- Connect to my existing knowledge
- Provide real-world applications
- Include comprehension checkpoints

Specific Focus:
Focus on practical usage with Qdrant and when to choose vector DB over traditional DB.

Environment:
Linux with Docker, Python preferred for examples.
```

## Tips

- Be specific about your current knowledge level
- Mention any tools/libraries you're already using
- Specify if you prefer theory-first or hands-on-first
- Include any specific use cases you're targeting
