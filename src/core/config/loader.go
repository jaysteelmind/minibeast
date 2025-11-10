package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a YAML configuration file
// Mathematical guarantee: Returns valid Config or error (never invalid Config)
// Complexity: O(n) where n = file size
func Load(path string) (*Config, error) {
	// Read file with atomic operation
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	cfg := Default() // Start with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Validate mathematical invariants
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadOrDefault attempts to load config, returns default on failure
// Mathematical guarantee: Always returns valid Config (never nil)
// Complexity: O(n) where n = file size
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		// Graceful degradation: use defaults
		return Default()
	}
	return cfg
}

// Save writes configuration to YAML file
// Mathematical guarantee: Atomic write (complete or nothing)
// Complexity: O(n) where n = config size
func Save(cfg *Config, path string) error {
	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Atomic write pattern: write to temp, then rename
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Cleanup on failure
		return fmt.Errorf("failed to rename config: %w", err)
	}

	return nil
}
