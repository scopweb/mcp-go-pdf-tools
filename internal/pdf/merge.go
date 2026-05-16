package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// MergePDFFiles merges multiple PDF files into a single output PDF.
// It takes a list of absolute paths for the input PDFs and the absolute path for the output PDF.
func MergePDFFiles(inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("no input files provided")
	}

	// Verify all input files exist
	for _, p := range inputPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", p)
		}
	}

	// Ensure output directory exists
	outDir := filepath.Dir(outputPath)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	conf := model.NewDefaultConfiguration()

	// false specifies that we don't want to insert a blank page (divider) between the original PDFs
	err := api.MergeCreateFile(inputPaths, outputPath, false, conf)
	if err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	return nil
}