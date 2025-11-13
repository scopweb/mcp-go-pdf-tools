package pdf

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// SplitPDFFile splits the PDF at inputPath into separate PDF files,
// writing them to a temporary directory. It returns the full paths
// to the generated files. Caller is responsible for removing them.
func SplitPDFFile(inputPath string) ([]string, error) {
	tmpDir, err := ioutil.TempDir("", "pdf-split-")
	if err != nil {
		return nil, err
	}

	// Use pdfcpu api to split the file. It will create files in tmpDir.
	// The current pdfcpu SplitFile signature expects: (inFile, outDir, pagesPerFile int, conf *model.Configuration)
	// Passing 1 will create one output file per page. No custom configuration is needed (nil).
	if err := api.SplitFile(inputPath, tmpDir, 1, nil); err != nil {
		// On error, clean up temp dir and return
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}

	// Collect generated files (pdfcpu names them like input-<n>.pdf)
	entries, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}

	var outFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		outFiles = append(outFiles, filepath.Join(tmpDir, e.Name()))
	}

	return outFiles, nil
}

// GetPDFInfo returns basic information about a PDF file: pages and file size.
// This uses a lightweight heuristic (counts occurrences of "/Type /Page") as
// a fallback when a full PDF parser is not needed. It also returns file size in bytes.
func GetPDFInfo(inputPath string) (map[string]interface{}, error) {
	if _, err := os.Stat(inputPath); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	// Heuristic page count: count occurrences of "/Type /Page"
	count := 0
	bs := []byte("/Type /Page")
	for i := 0; i+len(bs) <= len(data); i++ {
		if string(data[i:i+len(bs)]) == string(bs) {
			count++
		}
	}

	fi, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"file":      filepath.Base(inputPath),
		"pages":     count,
		"file_size": fi.Size(),
	}

	return info, nil
}
