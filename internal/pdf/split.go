package pdf

import (
	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/types"
)

// SplitPDFFile is a backward-compatible wrapper for splitting PDFs.
// Deprecated: Use Processor.Split() instead.
func SplitPDFFile(inputPath string) ([]string, error) {
	defaultConfig := config.PDFConfig{
		ValidationMode: "relaxed",
	}
	processor := NewProcessor(defaultConfig, logging.New("info"))
	return processor.Split(inputPath)
}

// GetPDFInfo is a backward-compatible wrapper for getting PDF information.
// Deprecated: Use Processor.GetInfo() instead, which returns *types.PDFInfoResult.
func GetPDFInfo(inputPath string) (map[string]interface{}, error) {
	defaultConfig := config.PDFConfig{
		ValidationMode: "relaxed",
	}
	processor := NewProcessor(defaultConfig, logging.New("info"))

	info, err := processor.GetInfo(inputPath)
	if err != nil {
		return nil, err
	}

	// Convert to map for backward compatibility
	return map[string]interface{}{
		"file":      info.Filename,
		"pages":     info.TotalPages,
		"file_size": info.SizeBytes,
	}, nil
}
