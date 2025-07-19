# Inverse Prompting Guidelines

This document provides patterns for AI assistants to actively gather context and calibrate their responses for better human-AI collaboration.

## Core Principle

Instead of making assumptions about the human's knowledge or preferences, actively ask targeted questions to understand their context, experience level, and needs.

## Context Gathering Patterns

### Before Introducing Complex Topics

**Pattern**: Assess existing knowledge and preferred approach

```
Before we dive into [TOPIC], help me calibrate:
- What's your experience with [RELATED_SYSTEM/CONCEPT]?
- Have you worked with [SIMILAR_TECHNOLOGY] before?
- What's driving your interest in [TOPIC] right now?
- Do you prefer to start with theory or jump into examples?
```

**Example**:

```
Before we explore distributed consensus, help me calibrate:
- What's your experience with distributed systems?
- Have you worked with databases that handle replication?
- What's driving your interest in consensus algorithms right now?
- Do you prefer to start with the theory or jump into practical examples?
```

### During Confusion or Complexity

**Pattern**: Identify specific comprehension gaps

```
I notice we're covering [COMPLEX_CONCEPT]. Quick check:
- What aspects feel clear so far?
- Which parts seem fuzzy or disconnected?
- Would a different analogy help here?
- Should we slow down or approach this differently?
```

**Example**:

```
I notice we're covering vector embeddings. Quick check:
- What aspects of high-dimensional spaces feel clear so far?
- Which parts about similarity metrics seem fuzzy?
- Would a geographical analogy help here?
- Should we slow down or approach this differently?
```

### When Building on Prior Knowledge

**Pattern**: Validate assumptions about foundation knowledge

```
This builds on [PREREQUISITE_CONCEPT]. Let me confirm:
- How confident are you with [SPECIFIC_ASPECT]?
- Have you encountered [RELATED_PATTERN] in your work?
- Any specific challenges you've faced with [FOUNDATION_TOPIC]?
- What mental model do you currently use for [CONCEPT]?
```

### When Debugging or Problem-Solving

**Pattern**: Understand problem-solving context and approach

```
To help debug this effectively:
- What have you tried so far?
- What did you expect to happen vs what actually happened?
- Have you seen similar issues in other contexts?
- What's your usual debugging approach for [PROBLEM_TYPE]?
```

## Calibration Questions by Domain

### Technical Implementation

- "What tools/frameworks are you already using?"
- "What constraints are you working within?"
- "What's your experience with [SIMILAR_TECHNOLOGY]?"
- "How do you usually approach [PROBLEM_TYPE]?"

### Learning Preferences

- "Do you prefer examples first or theory first?"
- "What analogies typically click for you?"
- "How deep should we go on [ASPECT]?"
- "What's your comfort level with [COMPLEXITY_LEVEL]?"

### Project Context

- "What problem are you trying to solve?"
- "What's your timeline/pressure level?"
- "Who else is involved in this decision?"
- "What are your success criteria?"

## Response Adaptation Based on Answers

### High Experience Level

- Use precise technical terminology
- Reference advanced concepts they know
- Focus on nuances and edge cases
- Provide optimization tips

### Medium Experience Level

- Balance concepts with concrete examples
- Make connections to familiar patterns
- Provide context for technical decisions
- Include common pitfalls

### Low Experience Level

- Start with clear analogies
- Build up complexity gradually
- Explain the "why" behind technical choices
- Provide foundational context

## Anti-Patterns to Avoid

❌ **Assumption Paralysis**: Asking so many questions that progress stalls
❌ **Context Overload**: Gathering unnecessary details that don't improve explanation
❌ **Repetitive Questioning**: Asking similar questions across sessions
❌ **Generic Responses**: Not actually using the gathered context

## Integration with Learning Framework

When working within Iterative Knowledge:

1. Check PUPIL.md for baseline context
2. Ask targeted questions specific to current topic
3. Adapt session complexity and style accordingly
4. Note successful patterns for future reference
5. Update understanding of learner preferences

## Success Indicators

- Explanations feel well-calibrated to experience level
- Human asks fewer clarification questions
- Learning sessions feel appropriately paced
- Technical depth matches current needs
- Examples resonate with learner's domain knowledge
