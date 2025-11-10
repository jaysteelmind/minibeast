package io

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer provides atomic file write operations
// Mathematical guarantee: Either complete valid file OR no file (never partial)
type Writer struct{}

// NewWriter creates a new atomic writer
// Complexity: O(1)
func NewWriter() *Writer {
	return &Writer{}
}

// WriteAtomic writes data to a file atomically using write-then-rename pattern
// Mathematical specification:
//  1. Write to temp file (path.tmp)
//  2. Fsync temp file (durability)
//  3. Rename temp to final (atomic operation)
//  4. Fsync parent directory (metadata persistence)
//
// POSIX guarantee: Rename is atomic - observers see either old or new file, never partial
// Complexity: O(n) where n = len(data)
func (w *Writer) WriteAtomic(path string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Step 1: Write to temporary file
	tempPath := path + ".tmp"
	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write data
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Step 2: Fsync for durability (flush to disk)
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Step 3: Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Cleanup on failure
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Step 4: Fsync parent directory for metadata persistence
	if err := syncDirectory(dir); err != nil {
		// Non-fatal: file is written, but metadata might not be durable
		// Log warning in production
		return fmt.Errorf("warning: failed to sync directory: %w", err)
	}

	return nil
}

// WriteJSON writes JSON data atomically
// Complexity: O(n) where n = len(jsonData)
func (w *Writer) WriteJSON(path string, jsonData []byte) error {
	return w.WriteAtomic(path, jsonData, 0644)
}

// WriteBinary writes binary data atomically
// Complexity: O(n) where n = len(data)
func (w *Writer) WriteBinary(path string, data []byte) error {
	return w.WriteAtomic(path, data, 0644)
}

// syncDirectory fsyncs a directory to ensure metadata changes are durable
// This is critical for atomic rename durability
// Complexity: O(1)
func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Fsync the directory
	return dir.Sync()
}

// EnsureDirectory creates directory if it doesn't exist
// Complexity: O(1)
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
// Complexity: O(1)
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
