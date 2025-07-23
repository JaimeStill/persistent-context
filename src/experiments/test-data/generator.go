package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// TestMemory represents a memory for testing
type TestMemory struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Source    string         `json:"source"`
	Type      string         `json:"type"`
	Size      string         `json:"size"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// Memory templates based on real Claude Code sessions
var templates = struct {
	CodeSnippets []string
	Discussions  []string
	FileContents []string
	Commands     []string
	Errors       []string
}{
	CodeSnippets: []string{
		"Added error handling to the memory consolidation function:\n```go\nif err := p.journal.ConsolidateMemories(ctx, memories); err != nil {\n    return fmt.Errorf(\"failed to consolidate memories: %%w\", err)\n}\n```",
		"Implemented batch processing for large memory groups:\n```go\nfor i := 0; i < len(memories); i += batchSize {\n    end := i + batchSize\n    if end > len(memories) {\n        end = len(memories)\n    }\n    batch := memories[i:end]\n    // Process batch\n}\n```",
		"Created new configuration struct for memory processing:\n```go\ntype MemoryConfig struct {\n    MaxTokens              int     `mapstructure:\"max_tokens\"`\n    SafetyMargin           float64 `mapstructure:\"safety_margin\"`\n    MemoryCountThreshold   uint32  `mapstructure:\"memory_count_threshold\"`\n}\n```",
		"Refactored the VectorJournal to use interface-based design:\n```go\ntype Journal interface {\n    CaptureContext(ctx context.Context, source string, content string, metadata map[string]any) (*models.MemoryEntry, error)\n    GetMemories(ctx context.Context, limit uint32) ([]*models.MemoryEntry, error)\n    ConsolidateMemories(ctx context.Context, memories []*models.MemoryEntry) error\n}\n```",
		"Added retry logic with exponential backoff:\n```go\nfor attempt := 0; attempt <= c.config.MaxRetries; attempt++ {\n    result, err = c.makeRequest(ctx, data)\n    if err == nil {\n        break\n    }\n    time.Sleep(time.Duration(attempt+1) * time.Second)\n}\n```",
	},
	Discussions: []string{
		"User asked about improving consolidation performance. The current approach times out with 7+ memories. We need to investigate batch processing strategies.",
		"Discussing the trade-offs between LLM-based consolidation and graph-based approaches. LLM provides semantic understanding but has performance limitations.",
		"User wants to ensure session continuity - memories from one Claude Code session should be accessible in the next session. This is critical for the project's success.",
		"Exploring whether consolidation should happen at capture time vs batch processing. Real-time consolidation could reduce latency but increase system load.",
		"User concerned about scalability - a typical Claude Code session generates hundreds of memories. Current architecture may not handle this volume efficiently.",
		"Investigating alternative models to phi3:mini. User has system constraints but needs better performance for consolidation tasks.",
		"Discussion about memory scoring algorithms - how to determine which memories are important enough to consolidate vs discard.",
		"User wants to maintain philosophical alignment - the system should mirror human memory consolidation patterns while being computationally efficient.",
	},
	FileContents: []string{
		"File: pkg/memory/processor.go\n" + strings.Repeat("// Memory processor handles consolidation\nfunc (p *Processor) processMemories(ctx context.Context, memories []*models.MemoryEntry) error {\n    // Implementation details...\n}\n", 50),
		"File: pkg/journal/vector.go\n" + strings.Repeat("// VectorJournal implements Journal interface\ntype VectorJournal struct {\n    vectorDB  vectordb.VectorDB\n    llmClient llm.LLM\n    config    *config.JournalConfig\n}\n", 40),
		"File: docker-compose.yml\n" + strings.Repeat("  ollama:\n    image: ollama/ollama:latest\n    volumes:\n      - ./data/ollama:/root/.ollama\n    ports:\n      - \"11434:11434\"\n", 20),
		"File: src/pkg/config/memory.go\n" + strings.Repeat("// Configuration validation\nfunc (c *MemoryConfig) ValidateConfig() error {\n    if c.MaxTokens <= 0 {\n        return fmt.Errorf(\"max_tokens must be positive\")\n    }\n    return nil\n}\n", 30),
	},
	Commands: []string{
		"Running: docker compose up -d --build",
		"Executing: go test ./pkg/memory/... -v",
		"Command: ./bin/persistent-context-cli memory list --limit 10",
		"Testing: curl -X POST http://localhost:8543/api/v1/journal/consolidate",
		"Debugging: docker logs persistent-context-ollama-1 --tail 100",
	},
	Errors: []string{
		"Error: failed to consolidate memories: failed after 4 attempts: Post \"http://ollama:11434/api/generate\": context deadline exceeded",
		"Warning: Cannot safely consolidate during context init - insufficient context window",
		"Error: failed to generate embedding: request failed with status 503",
		"Debug: Memory consolidation taking longer than expected: 7 memories, 2m15s elapsed",
		"Info: Consolidation completed successfully: 5 memories consolidated into 1 semantic memory",
	},
}

func generateMemory(id int, r *rand.Rand) TestMemory {
	types := []string{"code", "discussion", "file", "command", "error"}
	memType := types[r.Intn(len(types))]
	
	var content string
	var source string
	var size string
	
	switch memType {
	case "code":
		content = templates.CodeSnippets[r.Intn(len(templates.CodeSnippets))]
		source = "code_edit"
		size = "medium"
	case "discussion":
		content = templates.Discussions[r.Intn(len(templates.Discussions))]
		source = "user_discussion"
		size = "small"
	case "file":
		content = templates.FileContents[r.Intn(len(templates.FileContents))]
		source = "file_read"
		size = "large"
	case "command":
		content = templates.Commands[r.Intn(len(templates.Commands))]
		source = "command_execution"
		size = "small"
	case "error":
		content = templates.Errors[r.Intn(len(templates.Errors))]
		source = "system_event"
		size = "small"
	}
	
	// Add some randomization to content
	if r.Float32() < 0.3 {
		content = fmt.Sprintf("%s\n\nAdditional context: %s", content, generateRandomContext(r))
	}
	
	return TestMemory{
		ID:      fmt.Sprintf("mem_%d", id),
		Content: content,
		Source:  source,
		Type:    memType,
		Size:    size,
		Metadata: map[string]any{
			"session_id": fmt.Sprintf("session_%d", r.Intn(5)+1),
			"importance": r.Float64(),
			"tags":       generateTags(memType, r),
		},
		CreatedAt: time.Now().Add(-time.Duration(r.Intn(3600)) * time.Second),
	}
}

func generateRandomContext(r *rand.Rand) string {
	contexts := []string{
		"This change was made to address performance issues reported by the user.",
		"Part of the refactoring effort to improve code maintainability.",
		"Implementing user feedback from the previous session.",
		"Critical fix for production deployment.",
		"Experimental approach - may need revision.",
		"Following best practices from the Go community.",
		"Addressing code review comments.",
		"Temporary workaround until proper solution is implemented.",
	}
	return contexts[r.Intn(len(contexts))]
}

func generateTags(memType string, r *rand.Rand) []string {
	allTags := map[string][]string{
		"code":       {"refactor", "bugfix", "feature", "optimization", "cleanup"},
		"discussion": {"architecture", "requirements", "feedback", "planning", "review"},
		"file":       {"config", "implementation", "test", "documentation", "build"},
		"command":    {"build", "test", "deploy", "debug", "monitor"},
		"error":      {"critical", "warning", "info", "performance", "timeout"},
	}
	
	tags := allTags[memType]
	numTags := r.Intn(3) + 1
	
	selected := make([]string, 0, numTags)
	for i := 0; i < numTags && i < len(tags); i++ {
		tag := tags[r.Intn(len(tags))]
		// Avoid duplicates
		exists := false
		for _, s := range selected {
			if s == tag {
				exists = true
				break
			}
		}
		if !exists {
			selected = append(selected, tag)
		}
	}
	
	return selected
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Generate different sized datasets
	datasets := map[string]int{
		"small":  10,
		"medium": 50,
		"large":  200,
		"xlarge": 500,
	}
	
	for name, count := range datasets {
		memories := make([]TestMemory, count)
		for i := 0; i < count; i++ {
			memories[i] = generateMemory(i+1, r)
		}
		
		// Save to file
		filename := fmt.Sprintf("dataset_%s.json", name)
		data, err := json.MarshalIndent(memories, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling %s dataset: %v\n", name, err)
			continue
		}
		
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Printf("Error writing %s dataset: %v\n", name, err)
			continue
		}
		
		fmt.Printf("Generated %s dataset with %d memories\n", name, count)
	}
	
	// Generate a special dataset for stress testing consolidation
	stressMemories := make([]TestMemory, 7)
	for i := 0; i < 7; i++ {
		// Use file contents for larger memories
		stressMemories[i] = TestMemory{
			ID:      fmt.Sprintf("stress_%d", i+1),
			Content: templates.FileContents[i%len(templates.FileContents)],
			Source:  "file_read",
			Type:    "file",
			Size:    "large",
			Metadata: map[string]any{
				"session_id": "stress_test",
				"importance": 0.9,
				"tags":       []string{"stress", "consolidation", "test"},
			},
			CreatedAt: time.Now().Add(-time.Duration(i*60) * time.Second),
		}
	}
	
	data, _ := json.MarshalIndent(stressMemories, "", "  ")
	os.WriteFile("dataset_stress_7_memories.json", data, 0644)
	fmt.Println("Generated stress test dataset with 7 large memories")
}