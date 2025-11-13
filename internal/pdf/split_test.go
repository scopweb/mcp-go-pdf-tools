package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

// This test will run only if a test PDF exists at the repository root named `test.pdf`.
// It verifies that SplitPDFFile returns at least one part and cleans up generated files.
func TestSplitPDFFileIntegration(t *testing.T) {
	testPDF := filepath.Join("..", "..", "test.pdf")
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		t.Skip("test.pdf not found in repository root; skipping integration test")
	}

	parts, err := SplitPDFFile(testPDF)
	if err != nil {
		t.Fatalf("SplitPDFFile returned error: %v", err)
	}

	if len(parts) == 0 {
		t.Fatalf("expected at least one part, got 0")
	}

	// Ensure files exist then remove them
	partsDir := filepath.Dir(parts[0])
	for _, p := range parts {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected part file exists, but got error: %v", err)
		}
	}

	if err := os.RemoveAll(partsDir); err != nil {
		t.Logf("warning: failed to clean up parts dir: %v", err)
	}
}
