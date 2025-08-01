package qrfiletransfer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQRFileTransfer(t *testing.T) {
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

	// Create a test file
	testFilePath := filepath.Join(testDir, "test.txt")
	testContent := "This is a test file for QR file transfer. " +
		"It contains some text that will be split into chunks and converted to QR codes."

	if err := os.WriteFile(testFilePath, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create an output directory
	outDir := filepath.Join("testdata", "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create a QRFileTransfer instance
	qrft := NewQRFileTransfer()

	// Set a small chunk size to ensure multiple chunks are created
	qrft.maxChunkSize = 10

	// Convert the file to QR codes
	if err := qrft.FileToQRCodes(testFilePath, outDir); err != nil {
		t.Fatalf("FileToQRCodes failed: %v", err)
	}

	// Check if QR codes and data files were created
	qrDir := filepath.Join(outDir, "qrcodes")
	dataDir := filepath.Join(outDir, "data")

	qrFiles, err := filepath.Glob(filepath.Join(qrDir, "*.png"))
	if err != nil {
		t.Fatalf("Failed to list QR code files: %v", err)
	}

	if len(qrFiles) == 0 {
		t.Fatal("No QR code files were created")
	}

	dataFiles, err := filepath.Glob(filepath.Join(dataDir, "*.dat"))
	if err != nil {
		t.Fatalf("Failed to list data files: %v", err)
	}

	if len(dataFiles) == 0 {
		t.Fatal("No data files were created")
	}

	// Check if the number of QR codes matches the number of data files
	if len(qrFiles) != len(dataFiles) {
		t.Fatalf("Number of QR codes (%d) does not match number of data files (%d)", len(qrFiles), len(dataFiles))
	}

	// Create a directory for the reconstructed file
	reconstructDir := filepath.Join(testDir, "reconstruct")
	if err := os.MkdirAll(reconstructDir, 0755); err != nil {
		t.Fatalf("Failed to create directory for reconstructed file: %v", err)
	}

	// Reconstruct the file from QR codes
	reconstructedFilePath := filepath.Join(reconstructDir, "reconstructed.txt")
	if err := qrft.QRCodesToFile(outDir, reconstructedFilePath); err != nil {
		t.Fatalf("QRCodesToFile failed: %v", err)
	}

	// Check if the reconstructed file exists
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
		t.Fatalf("Reconstructed content does not match original content.\nOriginal: %s\nReconstructed: %s",
			testContent, string(reconstructedContent))
	}
}
