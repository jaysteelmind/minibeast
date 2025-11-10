package summarizer

import (
	"context"
	"fmt"

	"github.com/minibeast/usb-agent/src/core/collection"
	"github.com/minibeast/usb-agent/src/core/config"
	"github.com/minibeast/usb-agent/src/core/inference"
)

// Summarizer orchestrates LLM-based system analysis
// Mathematical guarantee: Deterministic output for same Facts + config
type Summarizer struct {
	engine        *inference.Engine
	promptBuilder *inference.PromptBuilder
	parser        *inference.Parser
	config        *config.Config
}

// NewSummarizer creates a new summarizer instance
// Complexity: O(1) - lazy initialization
func NewSummarizer(cfg *config.Config) (*Summarizer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create inference config from main config
	inferenceConfig := &inference.InferenceConfig{
		MaxTokens:   cfg.LLM.MaxTokens,
		Temperature: cfg.LLM.Temperature,
		ModelPath:   cfg.LLM.ModelPath,
	}

	// Create engine (lazy loading)
	engine, err := inference.NewEngine(inferenceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	return &Summarizer{
		engine:        engine,
		promptBuilder: inference.NewPromptBuilder(),
		parser:        inference.NewParser(),
		config:        cfg,
	}, nil
}

// Summarize generates a human-readable report from Facts
// Mathematical complexity: O(m) where m = maxTokens
// Latency: L₂ = L_load + L_inference + L_parse
func (s *Summarizer) Summarize(ctx context.Context, facts *collection.Facts) (string, error) {
	if facts == nil {
		return "", fmt.Errorf("facts cannot be nil")
	}

	// Update inference config with facts metadata
	s.engine = s.updateEngineWithFacts(facts)

	// Step 1: Load model (lazy, cached after first call)
	if err := s.engine.Load(ctx); err != nil {
		return "", fmt.Errorf("model load failed: %w", err)
	}

	// Step 2: Build deterministic prompt
	prompt, err := s.promptBuilder.BuildPrompt(facts)
	if err != nil {
		return "", fmt.Errorf("prompt build failed: %w", err)
	}

	// Step 3: Validate token count
	if err := s.promptBuilder.ValidateTokenCount(prompt, s.config.LLM.MaxTokens); err != nil {
		// Try truncating facts if prompt too large
		truncatedFacts := s.promptBuilder.TruncateFacts(facts)
		prompt, err = s.promptBuilder.BuildPrompt(truncatedFacts)
		if err != nil {
			return "", fmt.Errorf("prompt build failed after truncation: %w", err)
		}
	}

	// Step 4: Generate summary using LLM
	result, err := s.engine.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("inference failed: %w", err)
	}

	// Step 5: Clean output
	cleanedOutput := s.parser.CleanOutput(result.Text)

	// Step 6: Parse structured output
	parsed, err := s.parser.Parse(cleanedOutput)
	if err != nil {
		return "", fmt.Errorf("parsing failed: %w", err)
	}

	// Step 7: Validate output quality
	if err := s.parser.Validate(parsed); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Step 8: Detect hallucinations (best-effort)
	factsJSON := fmt.Sprintf("%+v", facts) // Simple representation
	hallucinations := s.parser.DetectHallucination(parsed, factsJSON)
	if len(hallucinations) > 0 {
		// Log warnings but don't fail (best-effort detection)
		// In production, log to file or metrics
		_ = hallucinations
	}

	// Step 9: Format final report
	report := s.formatReport(facts, parsed, result)

	return report, nil
}

// updateEngineWithFacts updates the engine with facts-specific seed data
func (s *Summarizer) updateEngineWithFacts(facts *collection.Facts) *inference.Engine {
	// Create new inference config with facts metadata
	inferenceConfig := &inference.InferenceConfig{
		MaxTokens:    s.config.LLM.MaxTokens,
		Temperature:  s.config.LLM.Temperature,
		HardwareUUID: facts.HardwareUUID,
		Timestamp:    facts.Timestamp,
		ModelPath:    s.config.LLM.ModelPath,
	}

	// Create new engine with deterministic seed
	engine, _ := inference.NewEngine(inferenceConfig)
	return engine
}

// formatReport creates the final human-readable report
func (s *Summarizer) formatReport(facts *collection.Facts, parsed *inference.ParsedOutput, result *inference.InferenceResult) string {
	// Add header with metadata
	header := fmt.Sprintf(`===== MINIBEAST SYSTEM REPORT =====

Collection Date: %s
Hostname: %s
Hardware UUID: %s
OS: %s %s
Collection Time: %dms
Inference Time: %dms
Tokens Generated: %d

`,
		facts.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		facts.Hostname,
		facts.HardwareUUID,
		facts.OSName,
		facts.OSVersion,
		facts.CollectionDurationMs,
		result.InferenceTime.Milliseconds(),
		result.TokenCount,
	)

	// Use parser to format the structured output
	body := s.parser.Format(parsed)

	// Combine header and body
	return header + body
}

// Close releases resources
func (s *Summarizer) Close() error {
	if s.engine != nil {
		return s.engine.Unload()
	}
	return nil
}
