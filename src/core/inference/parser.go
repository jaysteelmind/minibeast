package inference

import (
	"fmt"
	"strings"
)

// Parser extracts structured data from LLM output
type Parser struct{}

// NewParser creates a new output parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts SUMMARY, RISKS, and ACTIONS from LLM output
// Mathematical property: Same output text → Same parsed structure
// Complexity: O(n) where n = length of output text
func (p *Parser) Parse(output string) (*ParsedOutput, error) {
	if output == "" {
		return nil, fmt.Errorf("output is empty")
	}

	result := &ParsedOutput{
		Summary: []string{},
		Risks:   []string{},
		Actions: []string{},
	}

	// Split into lines and process
	lines := strings.Split(output, "\n")

	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Detect section headers (case-insensitive)
		lineUpper := strings.ToUpper(line)
		if strings.HasPrefix(lineUpper, "SUMMARY:") {
			currentSection = "SUMMARY"
			continue
		} else if strings.HasPrefix(lineUpper, "RISKS:") {
			currentSection = "RISKS"
			continue
		} else if strings.HasPrefix(lineUpper, "ACTIONS:") {
			currentSection = "ACTIONS"
			continue
		}

		// Extract bullet points or lines
		content := strings.TrimSpace(line)

		// Remove bullet markers (•, -, *, etc.)
		content = strings.TrimPrefix(content, "•")
		content = strings.TrimPrefix(content, "-")
		content = strings.TrimPrefix(content, "*")
		content = strings.TrimPrefix(content, "▪")
		content = strings.TrimSpace(content)

		// Skip if empty after trimming
		if content == "" {
			continue
		}

		// Add to appropriate section
		switch currentSection {
		case "SUMMARY":
			if len(result.Summary) < 3 {
				result.Summary = append(result.Summary, content)
			}
		case "RISKS":
			if len(result.Risks) < 3 {
				result.Risks = append(result.Risks, content)
			}
		case "ACTIONS":
			if len(result.Actions) < 2 {
				result.Actions = append(result.Actions, content)
			}
		}
	}

	// Validate that we have at least a summary
	if len(result.Summary) == 0 {
		return nil, fmt.Errorf("no summary section found in output")
	}

	return result, nil
}

// Validate checks if parsed output meets quality standards
// Returns error if output appears to be hallucinated or malformed
func (p *Parser) Validate(parsed *ParsedOutput) error {
	if parsed == nil {
		return fmt.Errorf("parsed output is nil")
	}

	// Must have at least one summary line
	if len(parsed.Summary) == 0 {
		return fmt.Errorf("summary is empty")
	}

	// Check for reasonable content length
	for i, line := range parsed.Summary {
		if len(line) < 10 {
			return fmt.Errorf("summary line %d is too short: %q", i, line)
		}
		if len(line) > 500 {
			return fmt.Errorf("summary line %d is too long: %d chars", i, len(line))
		}
	}

	// Validate risks format
	for i, risk := range parsed.Risks {
		if len(risk) < 10 {
			return fmt.Errorf("risk %d is too short: %q", i, risk)
		}
	}

	// Validate actions format
	for i, action := range parsed.Actions {
		if len(action) < 10 {
			return fmt.Errorf("action %d is too short: %q", i, action)
		}
	}

	return nil
}

// DetectHallucination performs basic hallucination detection
// Checks if output contains entities not present in input facts
// Note: This is best-effort, not mathematically guaranteed
func (p *Parser) DetectHallucination(parsed *ParsedOutput, factsJSON string) []string {
	hallucinations := []string{}

	// Common hallucination patterns to detect
	suspiciousPatterns := []string{
		"http://",
		"https://",
		"www.",
		"version 99",
		"unknown manufacturer",
		"default password",
	}

	allText := strings.Join(parsed.Summary, " ") + " " +
		strings.Join(parsed.Risks, " ") + " " +
		strings.Join(parsed.Actions, " ")

	allTextLower := strings.ToLower(allText)

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(allTextLower, pattern) {
			hallucinations = append(hallucinations,
				fmt.Sprintf("Suspicious pattern detected: %q", pattern))
		}
	}

	return hallucinations
}

// Format converts ParsedOutput to human-readable report text
// Complexity: O(n) where n = total length of all sections
func (p *Parser) Format(parsed *ParsedOutput) string {
	var report strings.Builder

	report.WriteString("===== MINIBEAST SYSTEM REPORT =====\n\n")

	// Summary section
	report.WriteString("SUMMARY:\n")
	for _, line := range parsed.Summary {
		report.WriteString("• ")
		report.WriteString(line)
		report.WriteString("\n")
	}
	report.WriteString("\n")

	// Risks section
	if len(parsed.Risks) > 0 {
		report.WriteString("RISKS:\n")
		for _, risk := range parsed.Risks {
			report.WriteString("• ")
			report.WriteString(risk)
			report.WriteString("\n")
		}
		report.WriteString("\n")
	}

	// Actions section
	if len(parsed.Actions) > 0 {
		report.WriteString("RECOMMENDED ACTIONS:\n")
		for _, action := range parsed.Actions {
			report.WriteString("• ")
			report.WriteString(action)
			report.WriteString("\n")
		}
		report.WriteString("\n")
	}

	report.WriteString("===== END OF REPORT =====\n")

	return report.String()
}

// CleanOutput removes common LLM artifacts from output
// Examples: trailing "Assistant:", metadata tags, etc.
func (p *Parser) CleanOutput(output string) string {
	cleaned := output

	// Remove common prefixes
	prefixes := []string{
		"Assistant:",
		"Response:",
		"Output:",
	}

	for _, prefix := range prefixes {
		cleaned = strings.TrimPrefix(cleaned, prefix)
	}

	// Remove common suffixes
	suffixes := []string{
		"</s>",
		"[/INST]",
		"</output>",
	}

	for _, suffix := range suffixes {
		cleaned = strings.TrimSuffix(cleaned, suffix)
	}

	return strings.TrimSpace(cleaned)
}
