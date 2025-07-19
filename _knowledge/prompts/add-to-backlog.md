# Prompt: Add Concept to Backlog

Use this prompt to capture new concepts for future learning when you don't have time for a full session.

## Prompt Template

```
Add to my learning backlog: [CONCEPT/TOPIC]

Context: [Where this came up - what were you working on?]

Priority: [high/medium/low] because [reason]

Related to: [Any existing concepts in your backlog or completed sessions]

Specific aspects: [What particular aspects interest you or seem important]

Notes: [Any additional context, resources, or thoughts]
```

## Example Usage

```
Add to my learning backlog: RAFT consensus algorithm

Context: Designing distributed memory consolidation system and realized I don't understand how to handle network partitions safely

Priority: high because this affects data consistency in my current project

Related to: distributed systems, eventual consistency (both in my backlog)

Specific aspects: 
- Leader election mechanism
- How it handles split-brain scenarios
- Performance implications vs simpler approaches

Notes: Found mentions in etcd documentation, seems like the gold standard for distributed consensus
```

## Quick Capture Version

For rapid capture during focused work:

```
Backlog: [concept] - [one line context] - [priority]
```

Example:

```
Backlog: Circuit breaker patterns - microservices keep failing in cascade - high
```

## Integration

The AI will:

1. Format your input according to backlog.yaml structure
2. Suggest appropriate tags
3. Identify relationships to existing concepts
4. Recommend priority based on your current work
