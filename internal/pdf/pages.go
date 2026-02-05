package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// RemovePagesFromFile removes or keeps specific pages from a PDF.
//
// Parameters:
//   - inputPath: path to the source PDF
//   - outputPath: path where the result will be written
//   - pageSelection: comma-separated page ranges, e.g. "2,5-8,11"
//   - keepMode: if true, keep only the specified pages (remove everything else);
//     if false, remove the specified pages (keep everything else)
//
// Returns a map with operation details (pages removed, pages kept, total original, output path).
func RemovePagesFromFile(inputPath, outputPath, pageSelection string, keepMode bool) (map[string]interface{}, error) {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", inputPath)
	}

	pageSelection = strings.TrimSpace(pageSelection)
	if pageSelection == "" {
		return nil, fmt.Errorf("page selection cannot be empty")
	}

	// Read context to get total page count
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}
	if err := api.ValidateContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate PDF: %w", err)
	}
	totalPages := ctx.PageCount

	// Parse the page selection into a set of page numbers
	selectedPages, err := parsePageSelection(pageSelection, totalPages)
	if err != nil {
		return nil, fmt.Errorf("invalid page selection: %w", err)
	}

	if len(selectedPages) == 0 {
		return nil, fmt.Errorf("page selection resolved to zero pages")
	}

	// Determine which pages to remove
	var pagesToRemove []int
	if keepMode {
		// Keep mode: remove all pages NOT in the selection
		keepSet := make(map[int]bool, len(selectedPages))
		for _, p := range selectedPages {
			keepSet[p] = true
		}
		for i := 1; i <= totalPages; i++ {
			if !keepSet[i] {
				pagesToRemove = append(pagesToRemove, i)
			}
		}
	} else {
		// Remove mode: remove the selected pages directly
		pagesToRemove = selectedPages
	}

	if len(pagesToRemove) == 0 {
		return nil, fmt.Errorf("no pages to remove (selection matches all pages)")
	}

	if len(pagesToRemove) >= totalPages {
		return nil, fmt.Errorf("cannot remove all %d pages from the PDF", totalPages)
	}

	// Build page selection entries for pdfcpu — each range as a separate slice element
	removeEntries := intsToPageSelectionSlice(pagesToRemove)

	// Ensure output directory exists
	outDir := filepath.Dir(outputPath)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	if err := api.RemovePagesFile(inputPath, outputPath, removeEntries, conf); err != nil {
		return nil, fmt.Errorf("failed to remove pages: %w", err)
	}

	// Calculate kept pages
	sort.Ints(pagesToRemove)
	remainingPages := totalPages - len(pagesToRemove)

	result := map[string]interface{}{
		"original_pages":  totalPages,
		"removed_pages":   pagesToRemove,
		"removed_count":   len(pagesToRemove),
		"remaining_pages": remainingPages,
		"mode":            modeLabel(keepMode),
		"selection":       pageSelection,
		"output_file":     filepath.Base(outputPath),
		"output_path":     outputPath,
	}

	return result, nil
}

// parsePageSelection parses a string like "2,5-8,11" into a sorted, unique
// list of page numbers. Pages outside [1, totalPages] cause an error.
func parsePageSelection(selection string, totalPages int) ([]int, error) {
	seen := make(map[int]bool)
	parts := strings.Split(selection, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range: "5-8"
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range start %q: %w", bounds[0], err)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range end %q: %w", bounds[1], err)
			}
			if start > end {
				return nil, fmt.Errorf("invalid range %d-%d: start > end", start, end)
			}
			if start < 1 || end > totalPages {
				return nil, fmt.Errorf("range %d-%d out of bounds (PDF has %d pages)", start, end, totalPages)
			}
			for i := start; i <= end; i++ {
				seen[i] = true
			}
		} else {
			// Single page: "2"
			p, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q: %w", part, err)
			}
			if p < 1 || p > totalPages {
				return nil, fmt.Errorf("page %d out of bounds (PDF has %d pages)", p, totalPages)
			}
			seen[p] = true
		}
	}

	pages := make([]int, 0, len(seen))
	for p := range seen {
		pages = append(pages, p)
	}
	sort.Ints(pages)
	return pages, nil
}

// intsToPageSelectionSlice converts a sorted slice of ints into a slice of
// compact range strings, collapsing consecutive numbers into ranges.
// e.g. [1,2,3,5,7,8,9] -> ["1-3","5","7-9"]
// Each element is a separate entry for pdfcpu's selectedPages parameter.
func intsToPageSelectionSlice(pages []int) []string {
	if len(pages) == 0 {
		return nil
	}

	sort.Ints(pages)
	var parts []string
	start := pages[0]
	end := pages[0]

	for i := 1; i < len(pages); i++ {
		if pages[i] == end+1 {
			end = pages[i]
		} else {
			parts = append(parts, rangeStr(start, end))
			start = pages[i]
			end = pages[i]
		}
	}
	parts = append(parts, rangeStr(start, end))
	return parts
}

func rangeStr(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}

func modeLabel(keepMode bool) string {
	if keepMode {
		return "keep"
	}
	return "remove"
}
