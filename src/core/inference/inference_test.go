package inference

import (
	"testing"
	"time"

	"github.com/minibeast/usb-agent/src/core/collection"
	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// TestNewEngine verifies engine creation
func TestNewEngine(t *testing.T) {
	config := &InferenceConfig{
		MaxTokens:    160,
		Temperature:  0.1,
		HardwareUUID: "test-uuid-123",
		Timestamp:    time.Now(),
		ModelPath:    "test.gguf",
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("NewEngine() failed: %v", err)
	}

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}

	if engine.maxTokens != 160 {
		t.Errorf("maxTokens = %d, want 160", engine.maxTokens)
	}

	if engine.temperature != 0.1 {
		t.Errorf("temperature = %f, want 0.1", engine.temperature)
	}
}

// TestGenerateDeterministicSeed verifies seed generation
func TestGenerateDeterministicSeed(t *testing.T) {
	uuid1 := "uuid-123"
	uuid2 := "uuid-456"
	ts := time.Now()

	// Same inputs should produce same seed
	seed1a := generateDeterministicSeed(uuid1, ts)
	seed1b := generateDeterministicSeed(uuid1, ts)

	if seed1a != seed1b {
		t.Error("Same inputs produced different seeds")
	}

	// Different UUIDs should produce different seeds
	seed2 := generateDeterministicSeed(uuid2, ts)
	if seed1a == seed2 {
		t.Error("Different UUIDs produced same seed")
	}
}

// TestPromptBuilder verifies prompt construction
func TestPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()

	facts := &collection.Facts{
		Timestamp:    time.Now().UTC(),
		Hostname:     "test-host",
		HardwareUUID: "test-uuid",
		OSName:       "Linux",
		OSVersion:    "24.04",
		Users:        []types.User{{Username: "testuser"}},
	}

	prompt, err := pb.BuildPrompt(facts)
	if err != nil {
		t.Fatalf("BuildPrompt() failed: %v", err)
	}

	if prompt == "" {
		t.Error("BuildPrompt() returned empty prompt")
	}

	// Check for key components
	if !contains(prompt, "SYSTEM FACTS") {
		t.Error("Prompt missing SYSTEM FACTS section")
	}

	if !contains(prompt, "OUTPUT FORMAT") {
		t.Error("Prompt missing OUTPUT FORMAT section")
	}

	if !contains(prompt, "test-host") {
		t.Error("Prompt missing hostname from facts")
	}
}

// TestEstimateTokenCount verifies token estimation
func TestEstimateTokenCount(t *testing.T) {
	pb := NewPromptBuilder()

	tests := []struct {
		text   string
		expect int
	}{
		{"", 0},
		{"test", 1},
		{"test text here", 3},
		{"a b c d", 1}, // 7 chars / 4 = 1.75 ≈ 1
	}

	for _, tt := range tests {
		got := pb.EstimateTokenCount(tt.text)
		if got != tt.expect {
			t.Errorf("EstimateTokenCount(%q) = %d, want %d", tt.text, got, tt.expect)
		}
	}
}

// TestParser verifies output parsing
func TestParser(t *testing.T) {
	parser := NewParser()

	validOutput := `SUMMARY:
- Line 1 of summary
- Line 2 of summary
- Line 3 of summary

RISKS:
- Risk item 1
- Risk item 2

ACTIONS:
- Action item 1`

	parsed, err := parser.Parse(validOutput)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(parsed.Summary) != 3 {
		t.Errorf("Summary length = %d, want 3", len(parsed.Summary))
	}

	if len(parsed.Risks) != 2 {
		t.Errorf("Risks length = %d, want 2", len(parsed.Risks))
	}

	if len(parsed.Actions) != 1 {
		t.Errorf("Actions length = %d, want 1", len(parsed.Actions))
	}
}

// TestParserEmptyOutput verifies error handling
func TestParserEmptyOutput(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse("")
	if err == nil {
		t.Error("Parse() should fail for empty output")
	}
}

// TestParserNoSummary verifies summary requirement
func TestParserNoSummary(t *testing.T) {
	parser := NewParser()

	invalidOutput := `RISKS:
- Some risk

ACTIONS:
- Some action`

	_, err := parser.Parse(invalidOutput)
	if err == nil {
		t.Error("Parse() should fail without summary section")
	}
}

// TestValidate verifies output validation
func TestValidate(t *testing.T) {
	parser := NewParser()

	validParsed := &ParsedOutput{
		Summary: []string{
			"This is a valid summary line",
			"Second line of summary",
		},
		Risks:   []string{"Valid risk description"},
		Actions: []string{"Valid action item"},
	}

	if err := parser.Validate(validParsed); err != nil {
		t.Errorf("Validate() failed for valid output: %v", err)
	}
}

// TestValidateShortSummary verifies length checks
func TestValidateShortSummary(t *testing.T) {
	parser := NewParser()

	invalidParsed := &ParsedOutput{
		Summary: []string{"short"}, // Too short
	}

	err := parser.Validate(invalidParsed)
	if err == nil {
		t.Error("Validate() should fail for too-short summary")
	}
}

// TestFormat verifies report formatting
func TestFormat(t *testing.T) {
	parser := NewParser()

	parsed := &ParsedOutput{
		Summary: []string{"Summary line 1", "Summary line 2"},
		Risks:   []string{"Risk 1"},
		Actions: []string{"Action 1"},
	}

	report := parser.Format(parsed)

	if report == "" {
		t.Error("Format() returned empty report")
	}

	if !contains(report, "SUMMARY") {
		t.Error("Report missing SUMMARY header")
	}

	if !contains(report, "RISKS") {
		t.Error("Report missing RISKS header")
	}

	if !contains(report, "RECOMMENDED ACTIONS") {
		t.Error("Report missing ACTIONS header")
	}

	if !contains(report, "Summary line 1") {
		t.Error("Report missing summary content")
	}
}

// TestCleanOutput verifies output cleaning
func TestCleanOutput(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input  string
		expect string
	}{
		{"Assistant: Hello", "Hello"},
		{"Response: Test", "Test"},
		{"Normal text", "Normal text"},
		{"Text</s>", "Text"},
		{"  spaces  ", "spaces"},
	}

	for _, tt := range tests {
		got := parser.CleanOutput(tt.input)
		if got != tt.expect {
			t.Errorf("CleanOutput(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
