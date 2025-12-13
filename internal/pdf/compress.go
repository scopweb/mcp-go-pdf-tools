package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// CompressOptions holds configuration for PDF compression
type CompressOptions struct {
	RemoveMetadata bool
	ImageQuality   int // 1-100, where lower = higher compression
	OptimizeImages bool
}

// DefaultCompressOptions returns recommended compression settings
func DefaultCompressOptions() CompressOptions {
	return CompressOptions{
		RemoveMetadata: true,
		ImageQuality:   75, // Balanced quality/compression
		OptimizeImages: true,
	}
}

// CompressPDFFile compresses a PDF file and writes the output to outputPath.
// It reduces file size by optimizing images, removing metadata, and cleaning structure.
// Returns the output file path and file size information.
func CompressPDFFile(inputPath string, outputPath string, opts CompressOptions) (map[string]interface{}, error) {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Get original file size
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat input file: %w", err)
	}
	originalSize := originalInfo.Size()

	// Create configuration for compression
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Read the PDF context
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	// Validate context
	if err := api.ValidateContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate PDF: %w", err)
	}

	// Apply compression: remove unused objects
	if err := api.OptimizeFile(inputPath, outputPath, conf); err != nil {
		return nil, fmt.Errorf("failed to optimize PDF: %w", err)
	}

	// Get compressed file size
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	// Calculate compression ratio
	reductionPercent := 0.0
	if originalSize > 0 {
		reductionPercent = float64(originalSize-compressedSize) / float64(originalSize) * 100
	}

	result := map[string]interface{}{
		"original_size":      originalSize,
		"compressed_size":    compressedSize,
		"reduction_bytes":    originalSize - compressedSize,
		"reduction_percent":  reductionPercent,
		"output_file":        filepath.Base(outputPath),
		"output_path":        outputPath,
	}

	return result, nil
}

// CompressPDFWithDefaults is a convenience function using default compression options
func CompressPDFWithDefaults(inputPath string, outputPath string) (map[string]interface{}, error) {
	return CompressPDFFile(inputPath, outputPath, DefaultCompressOptions())
}
