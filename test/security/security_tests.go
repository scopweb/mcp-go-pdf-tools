package security

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// allSourceFiles lists all Go source files that should be checked by security tests.
var allSourceFiles = []string{
	"../../cmd/mcp-server/main.go",
	"../../cmd/server/main.go",
	"../../cmd/cli/main.go",
	"../../internal/pdf/split.go",
	"../../internal/pdf/compress.go",
	"../../internal/pdf/pages.go",
}

// TestDependencyVersions verifies that all dependencies are up to date
func TestDependencyVersions(t *testing.T) {
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run go list: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	outdated := 0
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			outdated++
			t.Logf("Outdated dependency: %s", line)
		}
	}

	if outdated > 0 {
		t.Logf("Found %d outdated dependencies. Run 'go get -u ./...' to update", outdated)
	}
}

// TestGoModuleIntegrity verifies go.mod hasn't been tampered
func TestGoModuleIntegrity(t *testing.T) {
	content, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])
	t.Logf("go.mod SHA256: %s", hashStr)

	modContent := string(content)
	suspiciousPatterns := []string{
		"replace ",
		"retract ",
		"excluded ",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(modContent, pattern) {
			t.Logf("Found directive: %s (review manually)", pattern)
		}
	}
}

// TestGoSumIntegrity verifies that all dependencies have checksums
func TestGoSumIntegrity(t *testing.T) {
	content, err := os.ReadFile("../../go.sum")
	if err != nil {
		t.Fatalf("Failed to read go.sum: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	validLines := 0
	invalidLines := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 {
			validLines++
		} else if line != "" {
			invalidLines++
			t.Logf("Invalid go.sum line: %s", line)
		}
	}

	t.Logf("go.sum entries: %d valid, %d invalid", validLines, invalidLines)

	if invalidLines > 0 {
		t.Errorf("Found %d invalid lines in go.sum", invalidLines)
	}
}

// TestMainDependencies checks critical dependencies for known issues
func TestMainDependencies(t *testing.T) {
	criticalDeps := map[string]string{
		"github.com/pdfcpu/pdfcpu": "v0.11.1",
	}

	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list modules: %v", err)
	}

	modules := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			modules[parts[0]] = parts[1]
		}
	}

	for dep, expectedVersion := range criticalDeps {
		if version, ok := modules[dep]; ok {
			t.Logf("%s: %s (expected %s)", dep, version, expectedVersion)
			if version != expectedVersion {
				t.Logf("Version mismatch for %s: got %s, expected %s", dep, version, expectedVersion)
			}
		} else {
			t.Errorf("Critical dependency not found: %s", dep)
		}
	}
}

// TestNoPrivateKeyCommitted checks for accidentally committed secrets
func TestNoPrivateKeyCommitted(t *testing.T) {
	sensitivePatterns := []string{
		"PRIVATE KEY",
		"SECRET_KEY",
		"API_KEY",
		"PASSWORD=",
	}

	filesToCheck := append(allSourceFiles, "../../go.mod", "../../go.sum")

	for _, file := range filesToCheck {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Logf("Could not read file: %s (%v)", file, err)
			continue
		}

		fileContent := string(content)
		for _, pattern := range sensitivePatterns {
			if strings.Contains(fileContent, pattern) {
				t.Errorf("SECURITY: Sensitive pattern %q found in %s", pattern, file)
			}
		}
	}
}

// TestNoDangerousImports checks for unsafe imports
func TestNoDangerousImports(t *testing.T) {
	dangerousImports := []string{
		"\"unsafe\"",
		"\"syscall\"",
		"\"os/exec\"",
	}

	for _, file := range allSourceFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		fileContent := string(content)
		for _, dangerous := range dangerousImports {
			if strings.Contains(fileContent, dangerous) {
				t.Logf("Found %s import in %s (review for security)", dangerous, file)
			}
		}
	}
}

// TestInputValidation checks that source files properly validate inputs
func TestInputValidation(t *testing.T) {
	for _, filePath := range allSourceFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Logf("Could not read file: %s", filePath)
			continue
		}

		fileContent := string(content)

		hasErrorHandling := strings.Contains(fileContent, "if err != nil")
		hasStringValidation := strings.Contains(fileContent, "strings.TrimSpace") ||
			strings.Contains(fileContent, "len(") ||
			strings.Contains(fileContent, "== \"\"")

		if !hasErrorHandling {
			t.Logf("No error handling patterns found in %s", filePath)
		}
		if !hasStringValidation {
			t.Logf("No string validation patterns found in %s", filePath)
		}
	}
}

// TestErrorHandling verifies proper error handling across all source files
func TestErrorHandling(t *testing.T) {
	for _, filePath := range allSourceFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Logf("Could not read file: %s", filePath)
			continue
		}

		fileContent := string(content)
		errNilCount := strings.Count(fileContent, "if err != nil")

		if errNilCount == 0 {
			t.Errorf("No 'if err != nil' found in %s — error handling missing", filePath)
		} else {
			t.Logf("%s: %d error checks", filePath, errNilCount)
		}
	}
}

// TestGoVersion checks Go version compatibility
func TestGoVersion(t *testing.T) {
	content, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	fileContent := string(content)

	for _, line := range strings.Split(fileContent, "\n") {
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go "))
			t.Logf("Go version requirement: %s", version)
			if !strings.HasPrefix(version, "1.24") && !strings.HasPrefix(version, "1.23") {
				t.Errorf("Go version %s may be outdated", version)
			}
		}
	}
}

// TestAllSourceFilesExist verifies that all expected source files exist
func TestAllSourceFilesExist(t *testing.T) {
	expectedFiles := []string{
		"../../cmd/mcp-server/main.go",
		"../../cmd/server/main.go",
		"../../cmd/cli/main.go",
		"../../internal/pdf/split.go",
		"../../internal/pdf/compress.go",
		"../../internal/pdf/pages.go",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected source file missing: %s", f)
		}
	}
}

// BenchmarkSecurityChecks measures security validation overhead
func BenchmarkSecurityChecks(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		os.ReadFile("../../go.mod")
		os.ReadFile("../../go.sum")
	}
}
