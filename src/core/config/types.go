package config

import "time"

// Config represents the complete MiniBeast configuration
// Mathematical invariant: All fields have valid defaults
type Config struct {
	// PII collection toggle
	PII bool `yaml:"pii"`

	// Collection settings
	Collect CollectConfig `yaml:"collect"`

	// Output settings
	Output OutputConfig `yaml:"output"`

	// LLM settings (Phase 2 stub)
	LLM LLMConfig `yaml:"llm"`

	// Performance settings
	Performance PerformanceConfig `yaml:"performance"`
}

// CollectConfig defines data collection parameters
type CollectConfig struct {
	// Extended collection (future enhancement)
	Extended bool `yaml:"extended"`

	// WiFi SSID collection
	WiFiSSIDs bool `yaml:"wifi_ssids"`

	// Hardware ID collection
	HardwareIDs bool `yaml:"hardware_ids"`

	// Per-category timeout (milliseconds)
	CategoryTimeoutMs int `yaml:"category_timeout_ms"`
}

// OutputConfig defines output file settings
type OutputConfig struct {
	// Enable encryption (Phase 2 feature)
	Encrypt bool `yaml:"encrypt"`

	// Enable Ed25519 signing
	Sign bool `yaml:"sign"`

	// Fields to redact from output
	Redact []string `yaml:"redact"`

	// Output directory (relative to USB root)
	Directory string `yaml:"directory"`
}

// LLMConfig defines LLM inference settings (Phase 2)
type LLMConfig struct {
	// Enable LLM summarization
	Enabled bool `yaml:"enabled"`

	// Maximum tokens to generate
	MaxTokens int `yaml:"max_tokens"`

	// Temperature for sampling
	Temperature float64 `yaml:"temp"`

	// Model path (relative to USB root)
	ModelPath string `yaml:"model_path"`
}

// PerformanceConfig defines performance constraints
type PerformanceConfig struct {
	// Maximum goroutines for parallel collection
	MaxGoroutines int `yaml:"max_goroutines"`

	// Total Phase 1 timeout (milliseconds)
	Phase1TimeoutMs int `yaml:"phase1_timeout_ms"`

	// Total Phase 2 timeout (milliseconds)
	Phase2TimeoutMs int `yaml:"phase2_timeout_ms"`
}

// Default returns a Config with mathematical default values
// Complexity: O(1)
func Default() *Config {
	return &Config{
		PII: true,
		Collect: CollectConfig{
			Extended:          false,
			WiFiSSIDs:         true,
			HardwareIDs:       true,
			CategoryTimeoutMs: 500, // 500ms per category
		},
		Output: OutputConfig{
			Encrypt:   false,
			Sign:      true,
			Redact:    []string{},
			Directory: "out",
		},
		LLM: LLMConfig{
			Enabled:     true,
			MaxTokens:   160,
			Temperature: 0.1,
			ModelPath:   "models/tinyllama-1.1b-q4.gguf",
		},
		Performance: PerformanceConfig{
			MaxGoroutines:   8,
			Phase1TimeoutMs: 2000, // 2 seconds
			Phase2TimeoutMs: 3000, // 3 seconds
		},
	}
}

// Validate checks configuration mathematical invariants
// Returns error if invariants violated
// Complexity: O(1)
func (c *Config) Validate() error {
	// Validate positive timeouts
	if c.Collect.CategoryTimeoutMs <= 0 {
		return &ValidationError{Field: "collect.category_timeout_ms", Reason: "must be positive"}
	}
	if c.Performance.Phase1TimeoutMs <= 0 {
		return &ValidationError{Field: "performance.phase1_timeout_ms", Reason: "must be positive"}
	}
	if c.Performance.Phase2TimeoutMs <= 0 {
		return &ValidationError{Field: "performance.phase2_timeout_ms", Reason: "must be positive"}
	}

	// Validate goroutine bounds (prevent resource exhaustion)
	if c.Performance.MaxGoroutines < 1 || c.Performance.MaxGoroutines > 32 {
		return &ValidationError{Field: "performance.max_goroutines", Reason: "must be between 1 and 32"}
	}

	// Validate LLM parameters
	if c.LLM.MaxTokens < 1 || c.LLM.MaxTokens > 2048 {
		return &ValidationError{Field: "llm.max_tokens", Reason: "must be between 1 and 2048"}
	}
	if c.LLM.Temperature < 0.0 || c.LLM.Temperature > 2.0 {
		return &ValidationError{Field: "llm.temperature", Reason: "must be between 0.0 and 2.0"}
	}

	return nil
}

// GetCategoryTimeout returns the timeout duration for category collection
// Complexity: O(1)
func (c *Config) GetCategoryTimeout() time.Duration {
	return time.Duration(c.Collect.CategoryTimeoutMs) * time.Millisecond
}

// GetPhase1Timeout returns the total timeout for Phase 1
// Complexity: O(1)
func (c *Config) GetPhase1Timeout() time.Duration {
	return time.Duration(c.Performance.Phase1TimeoutMs) * time.Millisecond
}

// GetPhase2Timeout returns the total timeout for Phase 2
// Complexity: O(1)
func (c *Config) GetPhase2Timeout() time.Duration {
	return time.Duration(c.Performance.Phase2TimeoutMs) * time.Millisecond
}

// ValidationError represents a configuration validation failure
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return "config validation failed: " + e.Field + " - " + e.Reason
}
