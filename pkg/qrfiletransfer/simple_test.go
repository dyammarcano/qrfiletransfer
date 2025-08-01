package qrfiletransfer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSimpleFileCreation(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()

	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Errorf("Failed to remove test directory: %v", err)
		}
	}()

	// Convert to an absolute path
	testDir, err := filepath.Abs(testDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Using temporary directory: %s", testDir)

	// Create a subdirectory
	subDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file in the subdirectory
	filePath := filepath.Join(subDir, "test.txt")
	content := "This is a test file."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Read the file to verify it was created correctly
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != content {
		t.Fatalf("File content does not match. Expected: %s, Got: %s", content, string(readContent))
	}
}
