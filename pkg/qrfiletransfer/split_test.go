package qrfiletransfer

import (
	"os"
	"path/filepath"
	"testing"

	"awesomeProjectQrFileTransfer/pkg/split"
)

func TestSplitPackage(t *testing.T) {
	// Create a temporary directory for testing
	testDir, err := os.MkdirTemp(os.TempDir(), "split_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Convert to an absolute path
	testDir, err = filepath.Abs(testDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Using temporary directory: %s", testDir)

	// Create a test file
	testFilePath := filepath.Join(testDir, "test.txt")
	testContent := "This is a test file for the split package. It contains some text that will be split into chunks."
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create an output directory
	outDir := filepath.Join(testDir, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Open the test file
	file, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Create a Split instance
	s := split.NewSplit()

	// Split the file into chunks
	if err := s.SplitFile(file, outDir, 3); err != nil {
		t.Fatalf("SplitFile failed: %v", err)
	}

	// Check if chunks were created
	chunks, err := filepath.Glob(filepath.Join(outDir, "*.part"))
	if err != nil {
		t.Fatalf("Failed to list chunk files: %v", err)
	}

	firstChunk, err := filepath.Glob(filepath.Join(outDir, "*.tmp"))
	if err != nil {
		t.Fatalf("Failed to find first chunk: %v", err)
	}

	if len(chunks) == 0 && len(firstChunk) == 0 {
		t.Fatal("No chunk files were created")
	}

	// Merge the chunks
	if err := s.MergeFile(outDir); err != nil {
		t.Fatalf("MergeFile failed: %v", err)
	}

	// Check if the original file was reconstructed
	reconstructedFilePath := filepath.Join(outDir, "test.txt")
	if _, err := os.Stat(reconstructedFilePath); os.IsNotExist(err) {
		t.Fatal("Reconstructed file does not exist")
	}

	// Read the reconstructed file
	reconstructedContent, err := os.ReadFile(reconstructedFilePath)
	if err != nil {
		t.Fatalf("Failed to read reconstructed file: %v", err)
	}

	// Check if the reconstructed content matches the original content
	if string(reconstructedContent) != testContent {
		t.Fatalf("Reconstructed content does not match original content.\nOriginal: %s\nReconstructed: %s", testContent, string(reconstructedContent))
	}
}
