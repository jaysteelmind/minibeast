package inference

import "time"

// InferenceConfig contains configuration for GGUF inference
type InferenceConfig struct {
	MaxTokens    int       // Maximum tokens to generate (160)
	Temperature  float64   // Sampling temperature (0.1)
	HardwareUUID string    // For deterministic seed generation
	Timestamp    time.Time // For deterministic seed generation
	ModelPath    string    // Path to GGUF model file
}

// InferenceResult contains the output from LLM inference
type InferenceResult struct {
	Text          string        // Generated text
	TokenCount    int           // Number of tokens generated
	InferenceTime time.Duration // Time taken for inference
	Seed          int64         // Seed used for generation
}

// ParsedOutput contains structured LLM output
type ParsedOutput struct {
	Summary []string // 3-line summary (max)
	Risks   []string // Risk bullets (0-3)
	Actions []string // Action items (0-2)
}
