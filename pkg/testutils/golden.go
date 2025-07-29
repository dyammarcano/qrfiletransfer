package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// normalizeSystemdUnitContent normalizes systemd unit file content for comparison.
// It parses the content into sections and properties, sorts them, and returns a normalized string.
func normalizeSystemdUnitContent(content string) string {
	lines := strings.Split(content, "\n")
	sections := make(map[string][]string)

	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			if _, exists := sections[currentSection]; !exists {
				sections[currentSection] = []string{}
			}
		} else if currentSection != "" {
			// Add the property to the current section
			sections[currentSection] = append(sections[currentSection], line)
		}
	}

	// Sort the sections and properties
	sectionNames := make([]string, 0, len(sections))
	for section := range sections {
		sectionNames = append(sectionNames, section)
	}
	sort.Strings(sectionNames)

	// Build the normalized content
	var normalized strings.Builder
	for _, section := range sectionNames {
		normalized.WriteString(section + "\n")

		properties := sections[section]
		sort.Strings(properties)

		for _, property := range properties {
			normalized.WriteString(property + "\n")
		}
		normalized.WriteString("\n")
	}

	return normalized.String()
}

// CompareWithGoldenFile compares the actual content with the content of a golden file.
// If the update flag is true, it updates the golden file with the actual content.
// Returns true if the content matches or the golden file was updated, false otherwise.
func CompareWithGoldenFile(t *testing.T, goldenFilePath string, actualContent string, update bool) bool {
	t.Helper()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(goldenFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory for golden file: %v", err)
		return false
	}

	// If update flag is set, update the golden file
	if update {
		err := os.WriteFile(goldenFilePath, []byte(actualContent), 0644)
		if err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
			return false
		}
		t.Logf("Updated golden file: %s", goldenFilePath)
		return true
	}

	// Read the golden file
	expectedContent, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
		return false
	}

	// Check if this is a systemd unit file
	if strings.HasSuffix(goldenFilePath, ".service.golden") ||
		strings.HasSuffix(goldenFilePath, ".timer.golden") ||
		strings.HasSuffix(goldenFilePath, ".target.golden") {
		// Normalize the content for systemd unit files
		normalizedExpected := normalizeSystemdUnitContent(string(expectedContent))
		normalizedActual := normalizeSystemdUnitContent(actualContent)

		require.Equal(t, normalizedExpected, normalizedActual,
			fmt.Sprintf("Content does not match golden file (after normalization): %s", goldenFilePath))

		return normalizedExpected == normalizedActual
	}

	// For other file types, compare the content directly
	require.Equal(t, string(expectedContent), actualContent,
		fmt.Sprintf("Content does not match golden file: %s", goldenFilePath))

	return string(expectedContent) == actualContent
}

// ReadGoldenFile reads the content of a golden file.
// Returns the content as a string and an error if the file cannot be read.
func ReadGoldenFile(t *testing.T, goldenFilePath string) (string, error) {
	t.Helper()

	content, err := os.ReadFile(goldenFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read golden file: %w", err)
	}

	return string(content), nil
}

// UpdateGoldenFile updates the content of a golden file.
// Returns an error if the file cannot be updated.
func UpdateGoldenFile(t *testing.T, goldenFilePath string, content string) error {
	t.Helper()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(goldenFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for golden file: %w", err)
	}

	// Write the content to the file
	err := os.WriteFile(goldenFilePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to update golden file: %w", err)
	}

	t.Logf("Updated golden file: %s", goldenFilePath)
	return nil
}

func AssertFileMatchesGolden(t *testing.T, afs afero.Fs, actualPath string, goldenPath string) {
	t.Helper()

	// Read the actual content from the afero filesystem
	actualContent, err := afero.ReadFile(afs, actualPath)
	if err != nil {
		t.Fatalf("Failed to read actual file: %v", err)
		return
	}

	// Read the golden content from the afero filesystem
	goldenContent, err := afero.ReadFile(afs, goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
		return
	}

	// Compare the contents
	require.Equal(t, string(goldenContent), string(actualContent),
		fmt.Sprintf("Content does not match golden file: %s", goldenPath))
}
