package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/minibeast/usb-agent/src/core/io"
)

// TestWriteAtomic verifies atomic file writes
func TestWriteAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	writer := io.NewWriter()
	testData := []byte("Hello, MiniBeast!")

	if err := writer.WriteAtomic(testFile, testData, 0644); err != nil {
		t.Fatalf("WriteAtomic() failed: %v", err)
	}

	// Verify file exists
	if !io.FileExists(testFile) {
		t.Error("File was not created")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Content mismatch: got %q, want %q", content, testData)
	}

	// Verify no temp file left behind
	tempFile := testFile + ".tmp"
	if io.FileExists(tempFile) {
		t.Error("Temporary file was not cleaned up")
	}
}

// TestWriteAtomic_CreateDirectory verifies directory creation
func TestWriteAtomic_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "test.txt")

	writer := io.NewWriter()
	testData := []byte("test")

	if err := writer.WriteAtomic(nestedPath, testData, 0644); err != nil {
		t.Fatalf("WriteAtomic() failed: %v", err)
	}

	if !io.FileExists(nestedPath) {
		t.Error("File not created in nested directory")
	}
}

// TestWriteJSON verifies JSON file writing
func TestWriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	writer := io.NewWriter()
	jsonData := []byte(`{"key": "value"}`)

	if err := writer.WriteJSON(testFile, jsonData); err != nil {
		t.Fatalf("WriteJSON() failed: %v", err)
	}

	if !io.FileExists(testFile) {
		t.Error("JSON file was not created")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	if string(content) != string(jsonData) {
		t.Errorf("JSON content mismatch")
	}
}

// TestWriteBinary verifies binary file writing
func TestWriteBinary(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")

	writer := io.NewWriter()
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF}

	if err := writer.WriteBinary(testFile, binaryData); err != nil {
		t.Fatalf("WriteBinary() failed: %v", err)
	}

	if !io.FileExists(testFile) {
		t.Error("Binary file was not created")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read binary file: %v", err)
	}

	if len(content) != len(binaryData) {
		t.Errorf("Binary size mismatch: got %d, want %d", len(content), len(binaryData))
	}

	for i := range binaryData {
		if content[i] != binaryData[i] {
			t.Errorf("Binary content mismatch at byte %d", i)
			break
		}
	}
}

// TestReadFile verifies file reading
func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test data
	testData := []byte("Test content")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read with Reader
	reader := io.NewReader()
	content, err := reader.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Content mismatch: got %q, want %q", content, testData)
	}
}

// TestReadFile_NonExistent verifies error handling
func TestReadFile_NonExistent(t *testing.T) {
	reader := io.NewReader()
	_, err := reader.ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFile() should fail for nonexistent file")
	}
}

// TestReadJSON verifies JSON reading and unmarshaling
func TestReadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Write JSON test data
	jsonData := []byte(`{"name": "MiniBeast", "version": 1}`)
	if err := os.WriteFile(testFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test JSON: %v", err)
	}

	// Read and unmarshal
	reader := io.NewReader()
	var result map[string]interface{}
	if err := reader.ReadJSON(testFile, &result); err != nil {
		t.Fatalf("ReadJSON() failed: %v", err)
	}

	if result["name"] != "MiniBeast" {
		t.Errorf("JSON unmarshal failed: got %v", result)
	}
}

// TestReadJSON_InvalidJSON verifies error handling for malformed JSON
func TestReadJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(testFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	reader := io.NewReader()
	var result map[string]interface{}
	err := reader.ReadJSON(testFile, &result)
	if err == nil {
		t.Error("ReadJSON() should fail for invalid JSON")
	}
}

// TestGetFileSize verifies file size retrieval
func TestGetFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	testData := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	reader := io.NewReader()
	size, err := reader.GetFileSize(testFile)
	if err != nil {
		t.Fatalf("GetFileSize() failed: %v", err)
	}

	if size != int64(len(testData)) {
		t.Errorf("Size mismatch: got %d, want %d", size, len(testData))
	}
}

// TestGetFileSize_NonExistent verifies error handling
func TestGetFileSize_NonExistent(t *testing.T) {
	reader := io.NewReader()
	_, err := reader.GetFileSize("/nonexistent/file.txt")
	if err == nil {
		t.Error("GetFileSize() should fail for nonexistent file")
	}
}

// TestEnsureDirectory verifies directory creation
func TestEnsureDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "nested", "dir")

	if err := io.EnsureDirectory(testDir); err != nil {
		t.Fatalf("EnsureDirectory() failed: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path is not a directory")
	}
}

// TestEnsureDirectory_AlreadyExists verifies idempotency
func TestEnsureDirectory_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Call twice - should not fail
	if err := io.EnsureDirectory(tmpDir); err != nil {
		t.Fatalf("First EnsureDirectory() failed: %v", err)
	}

	if err := io.EnsureDirectory(tmpDir); err != nil {
		t.Fatalf("Second EnsureDirectory() failed: %v", err)
	}
}

// TestFileExists verifies file existence checking
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	nonExistentFile := filepath.Join(tmpDir, "not_exists.txt")

	// Create file
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !io.FileExists(existingFile) {
		t.Error("FileExists() returned false for existing file")
	}

	// Test non-existent file
	if io.FileExists(nonExistentFile) {
		t.Error("FileExists() returned true for non-existent file")
	}
}

// TestWriteAtomic_Overwrite verifies atomic overwriting
func TestWriteAtomic_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	writer := io.NewWriter()

	// Write initial content
	if err := writer.WriteAtomic(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("First WriteAtomic() failed: %v", err)
	}

	// Overwrite
	newData := []byte("overwritten")
	if err := writer.WriteAtomic(testFile, newData, 0644); err != nil {
		t.Fatalf("Second WriteAtomic() failed: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(newData) {
		t.Errorf("Overwrite failed: got %q, want %q", content, newData)
	}
}
