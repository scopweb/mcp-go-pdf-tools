package pdf

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/types"
)

// RemovePagesFromFile is a backward-compatible wrapper for removing/keeping pages.
// Deprecated: Use Processor.RemovePages() instead.
func RemovePagesFromFile(inputPath, outputPath, pageSelection string, keepMode bool) (map[string]interface{}, error) {
	defaultConfig := config.PDFConfig{
		ValidationMode: "relaxed",
	}
	processor := NewProcessor(defaultConfig, logging.New("info"))

	mode := pageRemovalModeFromBool(keepMode)
	result, err := processor.RemovePages(inputPath, outputPath, pageSelection, mode)
	if err != nil {
		return nil, err
	}

	// Convert to map for backward compatibility
	return map[string]interface{}{
		"original_pages":  result.OriginalPages,
		"removed_pages":   result.RemovedPages,
		"removed_count":   result.RemovedCount,
		"remaining_pages": result.RemainingPages,
		"mode":            modeLabel(keepMode),
		"selection":       pageSelection,
		"output_file":     filepath.Base(outputPath),
		"output_path":     outputPath,
	}, nil
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

// pageRemovalModeFromBool convierte un bool a PageRemovalMode para compatibilidad hacia atrás.
func pageRemovalModeFromBool(keepMode bool) types.PageRemovalMode {
	if keepMode {
		return types.ModeKeep
	}
	return types.ModeRemove
}
