package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// TestRootCommand tests the root command execution
func TestRootCommand(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the root command
	rootCmd.SetArgs([]string{"--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Restore stdout
	_ = w.Close()

	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check if the output contains expected text
	if !bytes.Contains(buf.Bytes(), []byte("QR File Transfer is a tool")) {
		t.Errorf("Expected help output to contain 'QR File Transfer is a tool', got: %s", output)
	}
}

// TestSplitCommandHelp tests the split command help
func TestSplitCommandHelp(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the split command with help flag
	rootCmd.SetArgs([]string{"split", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check if the output contains expected text
	if !bytes.Contains(buf.Bytes(), []byte("Split a file into QR code images")) {
		t.Errorf("Expected help output to contain 'Split a file into QR code images', got: %s", output)
	}
}

// TestJoinCommandHelp tests the join command help
func TestJoinCommandHelp(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the join command with help flag
	rootCmd.SetArgs([]string{"join", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Restore stdout
	_ = w.Close()

	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check if the output contains expected text
	if !bytes.Contains(buf.Bytes(), []byte("Join QR code images back into a file")) {
		t.Errorf("Expected help output to contain 'Join QR code images back into a file', got: %s", output)
	}
}
