package pdf

import (
	"path/filepath"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
)

// CompressOptions holds configuration for PDF compression
// Deprecated: Use config.PDFConfig instead.
type CompressOptions struct {
	RemoveMetadata bool
	ImageQuality   int // 1-100, where lower = higher compression
	OptimizeImages bool
}

// DefaultCompressOptions returns recommended compression settings
// Deprecated: Use config.PDFConfig defaults instead.
func DefaultCompressOptions() CompressOptions {
	return CompressOptions{
		RemoveMetadata: true,
		ImageQuality:   75, // Balanced quality/compression
		OptimizeImages: true,
	}
}

// CompressPDFFile is a backward-compatible wrapper for compressing PDFs.
// Deprecated: Use Processor.Compress() instead.
func CompressPDFFile(inputPath string, outputPath string, opts CompressOptions) (map[string]interface{}, error) {
	cfg := config.PDFConfig{
		ImageQuality:   opts.ImageQuality,
		RemoveMetadata: opts.RemoveMetadata,
		ValidationMode: "relaxed",
	}
	processor := NewProcessor(cfg, logging.New("info"))

	result, err := processor.Compress(inputPath, outputPath)
	if err != nil {
		return nil, err
	}

	// Convert to map for backward compatibility
	reductionBytes := result.OriginalSize - result.CompressedSize
	reductionPercent := result.CompressionRatio * 100

	return map[string]interface{}{
		"original_size":      result.OriginalSize,
		"compressed_size":    result.CompressedSize,
		"reduction_bytes":    reductionBytes,
		"reduction_percent":  reductionPercent,
		"output_file":        filepath.Base(outputPath),
		"output_path":        outputPath,
	}, nil
}

// CompressPDFWithDefaults is a convenience function using default compression options
// Deprecated: Use Processor.Compress() instead.
func CompressPDFWithDefaults(inputPath string, outputPath string) (map[string]interface{}, error) {
	return CompressPDFFile(inputPath, outputPath, DefaultCompressOptions())
}
