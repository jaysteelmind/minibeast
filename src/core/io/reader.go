package io

import (
	"encoding/json"
	"fmt"
	"os"
)

// Reader provides file read operations
type Reader struct{}

// NewReader creates a new reader
// Complexity: O(1)
func NewReader() *Reader {
	return &Reader{}
}

// ReadFile reads entire file contents
// Complexity: O(n) where n = file size
func (r *Reader) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// ReadJSON reads and unmarshals JSON file
// Complexity: O(n) where n = file size
func (r *Reader) ReadJSON(path string, v interface{}) error {
	data, err := r.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// GetFileSize returns file size in bytes
// Complexity: O(1)
func (r *Reader) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return info.Size(), nil
}
