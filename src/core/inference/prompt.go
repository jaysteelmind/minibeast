package inference

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minibeast/usb-agent/src/core/collection"
)

// PromptBuilder constructs deterministic prompts from Facts
type PromptBuilder struct {
	systemPrompt string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		systemPrompt: buildSystemPrompt(),
	}
}

// BuildPrompt creates a complete prompt from Facts
// Mathematical property: Same Facts → Same Prompt (deterministic)
// Complexity: O(|Facts|) for JSON serialization
func (pb *PromptBuilder) BuildPrompt(facts *collection.Facts) (string, error) {
	if facts == nil {
		return "", fmt.Errorf("facts cannot be nil")
	}

	// Convert Facts to JSON for structured input
	factsJSON, err := json.MarshalIndent(facts, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal facts: %w", err)
	}

	// Build complete prompt with system instructions + facts + format
	var prompt strings.Builder

	// System prompt with instructions
	prompt.WriteString(pb.systemPrompt)
	prompt.WriteString("\n\n")

	// Facts in JSON format
	prompt.WriteString("SYSTEM FACTS:\n")
	prompt.WriteString(string(factsJSON))
	prompt.WriteString("\n\n")

	// Output format instructions
	prompt.WriteString(buildOutputFormatInstructions())

	return prompt.String(), nil
}

// buildSystemPrompt creates the system-level instructions
// These are fixed and deterministic
func buildSystemPrompt() string {
	return `You are a system analysis assistant. Your task is to analyze system information and provide a concise, factual summary.

CRITICAL RULES:
1. Use ONLY the facts provided in the SYSTEM FACTS section
2. Do NOT invent, assume, or hallucinate any information
3. If a fact is "unknown", acknowledge it as unknown
4. Be concise: summary must be exactly 3 lines maximum
5. Only identify risks if clear thresholds are exceeded
6. Provide actionable recommendations, not generic advice

ANALYSIS GUIDELINES:
- Focus on hardware, network, and user configuration
- Identify potential security concerns (multiple admin accounts, unusual network configs)
- Note any deprecated OS versions or missing updates
- Highlight unusual user activity patterns
- Keep technical language clear but not overly simplified`
}

// buildOutputFormatInstructions provides the structured output format
func buildOutputFormatInstructions() string {
	return `OUTPUT FORMAT (follow exactly):

SUMMARY:
- [Line 1: System identification - OS, hostname, hardware]
- [Line 2: Key characteristics - users, network, notable configurations]
- [Line 3: Overall status assessment]

RISKS:
- [Only include if clear risk detected, max 3 bullets]
- [Each risk must reference specific facts from input]
- [Format: "Risk description (Evidence: specific fact)"]

ACTIONS:
- [Only include if actionable recommendation exists, max 2 items]
- [Must be specific and directly related to detected risks]
- [Format: "Action to take based on specific finding"]

Generate your analysis now:`
}

// EstimateTokenCount estimates the number of tokens in the prompt
// Used for context window management (2048 token limit)
// Heuristic: ~4 characters per token
func (pb *PromptBuilder) EstimateTokenCount(prompt string) int {
	return len(prompt) / 4
}

// ValidateTokenCount checks if prompt fits within context window
func (pb *PromptBuilder) ValidateTokenCount(prompt string, maxTokens int) error {
	estimatedTokens := pb.EstimateTokenCount(prompt)
	contextWindow := 2048 // TinyLlama context window

	// Reserve space for output (maxTokens) and prompt
	requiredTokens := estimatedTokens + maxTokens

	if requiredTokens > contextWindow {
		return fmt.Errorf("prompt too large: %d tokens required, %d available",
			requiredTokens, contextWindow)
	}

	return nil
}

// TruncateFacts reduces Facts size if prompt is too large
// Strategy: Remove less critical fields while preserving key info
func (pb *PromptBuilder) TruncateFacts(facts *collection.Facts) *collection.Facts {
	truncated := *facts // Copy

	// Keep only essential fields, truncate arrays
	if len(truncated.Users) > 10 {
		truncated.Users = truncated.Users[:10]
	}
	if len(truncated.WiFiSSIDs) > 10 {
		truncated.WiFiSSIDs = truncated.WiFiSSIDs[:10]
	}
	if len(truncated.HomeDirs) > 10 {
		truncated.HomeDirs = truncated.HomeDirs[:10]
	}

	return &truncated
}
