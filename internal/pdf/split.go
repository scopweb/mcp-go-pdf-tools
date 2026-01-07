package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// SplitPDFFile splits the PDF at inputPath into separate PDF files,
// writing them to a temporary directory. It returns the full paths
// to the generated files. Caller is responsible for removing them.
func SplitPDFFile(inputPath string) ([]string, error) {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", inputPath)
	}

	tmpDir, err := os.MkdirTemp("", "pdf-split-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Use pdfcpu api to split the file. It will create files in tmpDir.
	// The current pdfcpu SplitFile signature expects: (inFile, outDir, pagesPerFile int, conf *model.Configuration)
	// Passing 1 will create one output file per page. No custom configuration is needed (nil).
	if err := api.SplitFile(inputPath, tmpDir, 1, nil); err != nil {
		// On error, clean up temp dir and return
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("pdfcpu split failed: %w", err)
	}

	// Collect generated files (pdfcpu names them like input-<n>.pdf)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to read temp dir: %w", err)
	}

	var outFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Basic check to ensure we only pick up PDFs (pdfcpu should only generate PDFs but good to be safe)
		if filepath.Ext(e.Name()) == ".pdf" {
			outFiles = append(outFiles, filepath.Join(tmpDir, e.Name()))
		}
	}

	return outFiles, nil
}

// GetPDFInfo returns basic information about a PDF file: pages and file size.
// It uses pdfcpu to parse the context and get the accurate page count.
func GetPDFInfo(inputPath string) (map[string]interface{}, error) {
	fi, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read context/validate to get page count accurately
	// We use RelaxedConf because we just want info, even if the PDF has minor issues
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	// Validate to populate context details like page count if needed,
	// though ReadContextFile usually populates XRefTable/PageCount.
	// For simple page count, ReadContextFile is often enough, but let's be safe.
	if err := api.ValidateContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate PDF: %w", err)
	}

	info := map[string]interface{}{
		"file":      filepath.Base(inputPath),
		"pages":     ctx.PageCount,
		"file_size": fi.Size(),
	}

	return info, nil
}
