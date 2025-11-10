package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/minibeast/usb-agent/src/core/config"
)

// TestDefault verifies default configuration values
// Mathematical property: Default() always returns valid config
func TestDefault(t *testing.T) {
	cfg := config.Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	// Verify default values match PRD specifications
	if !cfg.PII {
		t.Error("Expected PII to be true by default")
	}
	if !cfg.Collect.WiFiSSIDs {
		t.Error("Expected WiFiSSIDs to be true by default")
	}
	if !cfg.Collect.HardwareIDs {
		t.Error("Expected HardwareIDs to be true by default")
	}
	if cfg.Collect.CategoryTimeoutMs != 500 {
		t.Errorf("Expected CategoryTimeoutMs=500, got %d", cfg.Collect.CategoryTimeoutMs)
	}
	if !cfg.Output.Sign {
		t.Error("Expected Sign to be true by default")
	}
	if cfg.Output.Encrypt {
		t.Error("Expected Encrypt to be false by default")
	}
	if cfg.Performance.MaxGoroutines != 8 {
		t.Errorf("Expected MaxGoroutines=8, got %d", cfg.Performance.MaxGoroutines)
	}
	if cfg.Performance.Phase1TimeoutMs != 2000 {
		t.Errorf("Expected Phase1TimeoutMs=2000, got %d", cfg.Performance.Phase1TimeoutMs)
	}
	if cfg.LLM.MaxTokens != 160 {
		t.Errorf("Expected MaxTokens=160, got %d", cfg.LLM.MaxTokens)
	}
	if cfg.LLM.Temperature != 0.1 {
		t.Errorf("Expected Temperature=0.1, got %f", cfg.LLM.Temperature)
	}
}

// TestValidate_Valid verifies validation passes for valid configs
func TestValidate_Valid(t *testing.T) {
	cfg := config.Default()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validation failed for default config: %v", err)
	}
}

// TestValidate_InvalidTimeouts verifies validation catches invalid timeouts
func TestValidate_InvalidTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		modifier func(*config.Config)
	}{
		{
			name: "zero category timeout",
			modifier: func(c *config.Config) {
				c.Collect.CategoryTimeoutMs = 0
			},
		},
		{
			name: "negative category timeout",
			modifier: func(c *config.Config) {
				c.Collect.CategoryTimeoutMs = -100
			},
		},
		{
			name: "zero phase1 timeout",
			modifier: func(c *config.Config) {
				c.Performance.Phase1TimeoutMs = 0
			},
		},
		{
			name: "negative phase2 timeout",
			modifier: func(c *config.Config) {
				c.Performance.Phase2TimeoutMs = -1000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			tt.modifier(cfg)

			if err := cfg.Validate(); err == nil {
				t.Error("Expected validation error, got nil")
			}
		})
	}
}

// TestValidate_InvalidGoroutines verifies goroutine bounds checking
func TestValidate_InvalidGoroutines(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
	}{
		{"zero goroutines", 0},
		{"negative goroutines", -5},
		{"too many goroutines", 33},
		{"way too many goroutines", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Performance.MaxGoroutines = tt.goroutines

			if err := cfg.Validate(); err == nil {
				t.Errorf("Expected validation error for %d goroutines", tt.goroutines)
			}
		})
	}
}

// TestValidate_InvalidLLMParams verifies LLM parameter validation
func TestValidate_InvalidLLMParams(t *testing.T) {
	tests := []struct {
		name     string
		modifier func(*config.Config)
	}{
		{
			name: "zero max tokens",
			modifier: func(c *config.Config) {
				c.LLM.MaxTokens = 0
			},
		},
		{
			name: "too many tokens",
			modifier: func(c *config.Config) {
				c.LLM.MaxTokens = 3000
			},
		},
		{
			name: "negative temperature",
			modifier: func(c *config.Config) {
				c.LLM.Temperature = -0.5
			},
		},
		{
			name: "temperature too high",
			modifier: func(c *config.Config) {
				c.LLM.Temperature = 2.5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			tt.modifier(cfg)

			if err := cfg.Validate(); err == nil {
				t.Error("Expected validation error, got nil")
			}
		})
	}
}

// TestGetTimeouts verifies timeout accessor methods
func TestGetTimeouts(t *testing.T) {
	cfg := config.Default()

	// Test category timeout
	categoryTimeout := cfg.GetCategoryTimeout()
	expected := 500 * time.Millisecond
	if categoryTimeout != expected {
		t.Errorf("GetCategoryTimeout() = %v, want %v", categoryTimeout, expected)
	}

	// Test phase1 timeout
	phase1Timeout := cfg.GetPhase1Timeout()
	expected = 2000 * time.Millisecond
	if phase1Timeout != expected {
		t.Errorf("GetPhase1Timeout() = %v, want %v", phase1Timeout, expected)
	}

	// Test phase2 timeout
	phase2Timeout := cfg.GetPhase2Timeout()
	expected = 3000 * time.Millisecond
	if phase2Timeout != expected {
		t.Errorf("GetPhase2Timeout() = %v, want %v", phase2Timeout, expected)
	}
}

// TestLoad_ValidYAML verifies loading from valid YAML file
func TestLoad_ValidYAML(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	// Write test YAML
	yamlContent := `pii: true
collect:
  extended: false
  wifi_ssids: true
  hardware_ids: true
  category_timeout_ms: 600
output:
  encrypt: false
  sign: true
  redact: []
  directory: "output"
llm:
  enabled: true
  max_tokens: 200
  temp: 0.2
  model_path: "models/test.gguf"
performance:
  max_goroutines: 10
  phase1_timeout_ms: 2500
  phase2_timeout_ms: 3500
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load and verify
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify parsed values
	if cfg.Collect.CategoryTimeoutMs != 600 {
		t.Errorf("CategoryTimeoutMs = %d, want 600", cfg.Collect.CategoryTimeoutMs)
	}
	if cfg.LLM.MaxTokens != 200 {
		t.Errorf("MaxTokens = %d, want 200", cfg.LLM.MaxTokens)
	}
	if cfg.Performance.MaxGoroutines != 10 {
		t.Errorf("MaxGoroutines = %d, want 10", cfg.Performance.MaxGoroutines)
	}
}

// TestLoad_NonExistentFile verifies error handling for missing files
func TestLoad_NonExistentFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// TestLoad_InvalidYAML verifies error handling for malformed YAML
func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidYAML := `pii: true
collect:
  extended: false
  invalid_indent
    broken: yaml
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	_, err := config.Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestLoad_InvalidConfig verifies validation during load
func TestLoad_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_values.yaml")

	// Write config with invalid values
	invalidConfig := `pii: true
collect:
  category_timeout_ms: -100
performance:
  max_goroutines: 8
  phase1_timeout_ms: 2000
  phase2_timeout_ms: 3000
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err := config.Load(configPath)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}

// TestLoadOrDefault_Success verifies successful load
func TestLoadOrDefault_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid.yaml")

	yamlContent := `pii: false
performance:
  max_goroutines: 4
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg := config.LoadOrDefault(configPath)
	if cfg == nil {
		t.Fatal("LoadOrDefault() returned nil")
	}
	if cfg.PII {
		t.Error("Expected PII to be false")
	}
	if cfg.Performance.MaxGoroutines != 4 {
		t.Errorf("MaxGoroutines = %d, want 4", cfg.Performance.MaxGoroutines)
	}
}

// TestLoadOrDefault_Fallback verifies fallback to default on error
func TestLoadOrDefault_Fallback(t *testing.T) {
	cfg := config.LoadOrDefault("/nonexistent/config.yaml")
	if cfg == nil {
		t.Fatal("LoadOrDefault() returned nil")
	}

	// Should return default config
	defaultCfg := config.Default()
	if cfg.Performance.MaxGoroutines != defaultCfg.Performance.MaxGoroutines {
		t.Error("LoadOrDefault() did not return default config on error")
	}
}

// TestSave verifies configuration persistence
func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved.yaml")

	// Create and modify config
	cfg := config.Default()
	cfg.Performance.MaxGoroutines = 12
	cfg.LLM.MaxTokens = 180

	// Save
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load back and verify
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.Performance.MaxGoroutines != 12 {
		t.Errorf("MaxGoroutines = %d, want 12", loaded.Performance.MaxGoroutines)
	}
	if loaded.LLM.MaxTokens != 180 {
		t.Errorf("MaxTokens = %d, want 180", loaded.LLM.MaxTokens)
	}
}

// TestSave_InvalidConfig verifies Save rejects invalid configs
func TestSave_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	cfg := config.Default()
	cfg.Performance.MaxGoroutines = -5 // Invalid

	err := config.Save(cfg, configPath)
	if err == nil {
		t.Error("Expected Save() to reject invalid config")
	}

	// Verify file was not created
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Invalid config file should not have been created")
	}
}

// TestSave_AtomicWrite verifies atomic write behavior
func TestSave_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "atomic.yaml")

	cfg := config.Default()

	// Save should be atomic (no .tmp file left behind)
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify no temp file exists
	tmpPath := configPath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Temporary file was not cleaned up")
	}

	// Verify final file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Final config file does not exist")
	}
}
