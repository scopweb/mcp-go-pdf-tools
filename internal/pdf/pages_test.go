package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParsePageSelection(t *testing.T) {
	tests := []struct {
		name       string
		selection  string
		totalPages int
		want       []int
		wantErr    bool
	}{
		{
			name:       "single page",
			selection:  "5",
			totalPages: 10,
			want:       []int{5},
		},
		{
			name:       "multiple single pages",
			selection:  "1,3,5",
			totalPages: 10,
			want:       []int{1, 3, 5},
		},
		{
			name:       "range",
			selection:  "2-5",
			totalPages: 10,
			want:       []int{2, 3, 4, 5},
		},
		{
			name:       "mixed ranges and singles",
			selection:  "2,5-8,11",
			totalPages: 15,
			want:       []int{2, 5, 6, 7, 8, 11},
		},
		{
			name:       "complex real-world selection",
			selection:  "7-10,61-66,77,80",
			totalPages: 100,
			want:       []int{7, 8, 9, 10, 61, 62, 63, 64, 65, 66, 77, 80},
		},
		{
			name:       "with spaces",
			selection:  " 2 , 5 - 8 , 11 ",
			totalPages: 15,
			want:       []int{2, 5, 6, 7, 8, 11},
		},
		{
			name:       "first page",
			selection:  "1",
			totalPages: 5,
			want:       []int{1},
		},
		{
			name:       "last page",
			selection:  "10",
			totalPages: 10,
			want:       []int{10},
		},
		{
			name:       "all pages range",
			selection:  "1-5",
			totalPages: 5,
			want:       []int{1, 2, 3, 4, 5},
		},
		{
			name:       "duplicate pages deduplicated",
			selection:  "1,1,2,2-3",
			totalPages: 5,
			want:       []int{1, 2, 3},
		},
		// Error cases
		{
			name:       "empty selection",
			selection:  "",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "page zero",
			selection:  "0",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "negative page",
			selection:  "-1",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "page exceeds total",
			selection:  "11",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "range exceeds total",
			selection:  "5-15",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "reversed range",
			selection:  "8-5",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "non-numeric",
			selection:  "abc",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "range with non-numeric start",
			selection:  "a-5",
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "range with non-numeric end",
			selection:  "1-b",
			totalPages: 10,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePageSelection(tt.selection, tt.totalPages)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePageSelection(%q, %d) expected error, got %v", tt.selection, tt.totalPages, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePageSelection(%q, %d) unexpected error: %v", tt.selection, tt.totalPages, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePageSelection(%q, %d) = %v, want %v", tt.selection, tt.totalPages, got, tt.want)
			}
		})
	}
}

func TestIntsToPageSelectionSlice(t *testing.T) {
	tests := []struct {
		name  string
		pages []int
		want  []string
	}{
		{
			name:  "single page",
			pages: []int{5},
			want:  []string{"5"},
		},
		{
			name:  "consecutive range",
			pages: []int{1, 2, 3},
			want:  []string{"1-3"},
		},
		{
			name:  "non-consecutive",
			pages: []int{1, 3, 5},
			want:  []string{"1", "3", "5"},
		},
		{
			name:  "mixed ranges",
			pages: []int{1, 2, 3, 5, 7, 8, 9},
			want:  []string{"1-3", "5", "7-9"},
		},
		{
			name:  "two adjacent ranges",
			pages: []int{1, 2, 3, 5, 6, 7},
			want:  []string{"1-3", "5-7"},
		},
		{
			name:  "empty",
			pages: []int{},
			want:  nil,
		},
		{
			name:  "nil",
			pages: nil,
			want:  nil,
		},
		{
			name:  "unsorted input",
			pages: []int{9, 1, 3, 2, 8, 7},
			want:  []string{"1-3", "7-9"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intsToPageSelectionSlice(tt.pages)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("intsToPageSelectionSlice(%v) = %v, want %v", tt.pages, got, tt.want)
			}
		})
	}
}

func TestRangeStr(t *testing.T) {
	tests := []struct {
		start, end int
		want       string
	}{
		{1, 1, "1"},
		{1, 3, "1-3"},
		{10, 20, "10-20"},
		{100, 100, "100"},
	}

	for _, tt := range tests {
		got := rangeStr(tt.start, tt.end)
		if got != tt.want {
			t.Errorf("rangeStr(%d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestModeLabel(t *testing.T) {
	if modeLabel(true) != "keep" {
		t.Error("modeLabel(true) should return 'keep'")
	}
	if modeLabel(false) != "remove" {
		t.Error("modeLabel(false) should return 'remove'")
	}
}

// TestRemovePagesFromFileValidation tests error handling without needing a real PDF.
func TestRemovePagesFromFileValidation(t *testing.T) {
	// Non-existent file
	_, err := RemovePagesFromFile("non_existent.pdf", "out.pdf", "1", false)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Empty page selection
	tmpFile, err := os.CreateTemp("", "test_*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = RemovePagesFromFile(tmpFile.Name(), "out.pdf", "", false)
	if err == nil {
		t.Error("expected error for empty page selection")
	}
}

// TestRemovePagesFromFileIntegration tests actual page removal with a real PDF.
func TestRemovePagesFromFileIntegration(t *testing.T) {
	testPDF := filepath.Join("..", "..", "examples", "test2.pdf")
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		testPDF = filepath.Join("..", "..", "test.pdf")
		if _, err := os.Stat(testPDF); os.IsNotExist(err) {
			t.Skip("no test PDF found; skipping integration test")
		}
	}

	// Get page count first
	info, err := GetPDFInfo(testPDF)
	if err != nil {
		t.Fatalf("GetPDFInfo failed: %v", err)
	}
	totalPages := info["pages"].(int)
	if totalPages < 3 {
		t.Skip("test PDF needs at least 3 pages")
	}

	t.Run("remove mode", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "removed.pdf")
		result, err := RemovePagesFromFile(testPDF, outFile, "1", false)
		if err != nil {
			t.Fatalf("RemovePagesFromFile failed: %v", err)
		}

		if result["mode"] != "remove" {
			t.Errorf("mode = %v, want 'remove'", result["mode"])
		}
		if result["removed_count"] != 1 {
			t.Errorf("removed_count = %v, want 1", result["removed_count"])
		}
		if result["remaining_pages"] != totalPages-1 {
			t.Errorf("remaining_pages = %v, want %d", result["remaining_pages"], totalPages-1)
		}

		// Verify output file exists
		if _, err := os.Stat(outFile); os.IsNotExist(err) {
			t.Error("output file was not created")
		}
	})

	t.Run("keep mode", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "kept.pdf")
		result, err := RemovePagesFromFile(testPDF, outFile, "1-2", true)
		if err != nil {
			t.Fatalf("RemovePagesFromFile failed: %v", err)
		}

		if result["mode"] != "keep" {
			t.Errorf("mode = %v, want 'keep'", result["mode"])
		}
		if result["remaining_pages"] != 2 {
			t.Errorf("remaining_pages = %v, want 2", result["remaining_pages"])
		}
		if result["removed_count"] != totalPages-2 {
			t.Errorf("removed_count = %v, want %d", result["removed_count"], totalPages-2)
		}
	})

	t.Run("cannot remove all pages", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "all.pdf")
		sel := "1-" + fmt.Sprintf("%d", totalPages)
		_, err := RemovePagesFromFile(testPDF, outFile, sel, false)
		if err == nil {
			t.Error("expected error when removing all pages")
		}
	})
}

// fmt is needed for the Sprintf in the integration test
func fmt_Sprintf(format string, a ...interface{}) string {
	// This is a workaround — we import fmt via the test file's own import
	return ""
}
