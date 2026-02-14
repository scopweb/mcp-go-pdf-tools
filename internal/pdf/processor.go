package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/types"
)

// Processor encapsula la lógica de manipulación de PDF con inyección de dependencias.
type Processor struct {
	config config.PDFConfig
	logger logging.Logger
}

// NewProcessor crea un nuevo procesador PDF.
func NewProcessor(cfg config.PDFConfig, logger logging.Logger) *Processor {
	return &Processor{
		config: cfg,
		logger: logger,
	}
}

// ValidateFile verifica que el archivo PDF exista y sea válido.
func (p *Processor) ValidateFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", path)
	}

	// Validar que sea un PDF legible
	ctx, err := api.ReadContextFile(path)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %w", err)
	}

	if err := api.ValidateContext(ctx); err != nil {
		return fmt.Errorf("failed to validate PDF: %w", err)
	}

	return nil
}

// Split divide un PDF en archivos de una página cada uno.
func (p *Processor) Split(inputPath string) ([]string, error) {
	p.logger.Debug("splitting PDF",
		fmt.Sprintf("path=%s", inputPath))

	if err := p.ValidateFile(inputPath); err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "pdf-split-")
	if err != nil {
		p.logger.Error("failed to create temp directory", err)
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	conf := p.newConfiguration()
	if err := api.SplitFile(inputPath, tmpDir, 1, conf); err != nil {
		_ = os.RemoveAll(tmpDir)
		p.logger.Error("PDF split operation failed", err)
		return nil, fmt.Errorf("pdfcpu split failed: %w", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		p.logger.Error("failed to read temp directory", err)
		return nil, fmt.Errorf("failed to read temp dir: %w", err)
	}

	var outFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".pdf" {
			outFiles = append(outFiles, filepath.Join(tmpDir, e.Name()))
		}
	}

	p.logger.Debug("PDF split complete",
		fmt.Sprintf("output_files=%d", len(outFiles)))

	return outFiles, nil
}

// GetInfo retorna información sobre un PDF.
func (p *Processor) GetInfo(inputPath string) (*types.PDFInfoResult, error) {
	p.logger.Debug("reading PDF info",
		fmt.Sprintf("path=%s", inputPath))

	if err := p.ValidateFile(inputPath); err != nil {
		return nil, err
	}

	fi, err := os.Stat(inputPath)
	if err != nil {
		p.logger.Error("failed to stat file", err)
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		p.logger.Error("failed to read PDF context", err)
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	return &types.PDFInfoResult{
		TotalPages: ctx.PageCount,
		SizeBytes:  fi.Size(),
		Filename:   filepath.Base(inputPath),
	}, nil
}

// Compress comprime un PDF optimizando imágenes y limpiando metadata.
func (p *Processor) Compress(inputPath, outputPath string) (*types.CompressResult, error) {
	p.logger.Debug("compressing PDF",
		fmt.Sprintf("input=%s", inputPath),
		fmt.Sprintf("output=%s", outputPath))

	if err := p.ValidateFile(inputPath); err != nil {
		return nil, err
	}

	// Obtener tamaño original
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		p.logger.Error("failed to stat input file", err)
		return nil, fmt.Errorf("failed to stat input file: %w", err)
	}
	originalSize := originalInfo.Size()

	// Asegurar directorio de salida
	if err := ensureOutputDir(outputPath); err != nil {
		p.logger.Error("failed to create output directory", err)
		return nil, err
	}

	conf := p.newConfiguration()
	if err := api.OptimizeFile(inputPath, outputPath, conf); err != nil {
		p.logger.Error("PDF compression failed", err)
		return nil, fmt.Errorf("failed to optimize PDF: %w", err)
	}

	// Obtener tamaño comprimido
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		p.logger.Error("failed to stat output file", err)
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	// Calcular ratio de compresión
	ratio := 1.0
	if originalSize > 0 {
		ratio = float64(originalSize-compressedSize) / float64(originalSize)
	}

	result := &types.CompressResult{
		OutputPath:       outputPath,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: ratio,
	}

	p.logger.Debug("PDF compression complete",
		fmt.Sprintf("original_size=%d", originalSize),
		fmt.Sprintf("compressed_size=%d", compressedSize))

	return result, nil
}

// RemovePages elimina o conserva páginas específicas de un PDF.
func (p *Processor) RemovePages(inputPath, outputPath, pageSelection string, mode types.PageRemovalMode) (*types.RemovePagesResult, error) {
	p.logger.Debug("removing pages from PDF",
		fmt.Sprintf("input=%s", inputPath),
		fmt.Sprintf("mode=%s", mode))

	if !mode.IsValid() {
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}

	if err := p.ValidateFile(inputPath); err != nil {
		return nil, err
	}

	// Leer contexto para obtener página total
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		p.logger.Error("failed to read PDF context", err)
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	if err := api.ValidateContext(ctx); err != nil {
		p.logger.Error("failed to validate PDF", err)
		return nil, fmt.Errorf("failed to validate PDF: %w", err)
	}

	totalPages := ctx.PageCount

	// Parsear selección de páginas
	selectedPages, err := parsePageSelection(pageSelection, totalPages)
	if err != nil {
		p.logger.Warn("invalid page selection", fmt.Errorf("selection=%s: %w", pageSelection, err))
		return nil, err
	}

	if len(selectedPages) == 0 {
		return nil, fmt.Errorf("page selection resolved to zero pages")
	}

	// Determinar páginas a eliminar
	var pagesToRemove []int
	if mode == types.ModeKeep {
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
		pagesToRemove = selectedPages
	}

	if len(pagesToRemove) == 0 {
		return nil, fmt.Errorf("no pages to remove (selection matches all pages)")
	}

	if len(pagesToRemove) >= totalPages {
		return nil, fmt.Errorf("cannot remove all %d pages from the PDF", totalPages)
	}

	// Asegurar directorio de salida
	if err := ensureOutputDir(outputPath); err != nil {
		p.logger.Error("failed to create output directory", err)
		return nil, err
	}

	// Construir selección de páginas para pdfcpu
	removeEntries := intsToPageSelectionSlice(pagesToRemove)

	conf := p.newConfiguration()
	if err := api.RemovePagesFile(inputPath, outputPath, removeEntries, conf); err != nil {
		p.logger.Error("failed to remove pages", err)
		return nil, fmt.Errorf("failed to remove pages: %w", err)
	}

	result := &types.RemovePagesResult{
		OutputPath:     outputPath,
		OriginalPages:  totalPages,
		RemovedPages:   pagesToRemove,
		RemovedCount:   len(pagesToRemove),
		RemainingPages: totalPages - len(pagesToRemove),
		Mode:           mode,
	}

	p.logger.Debug("page removal complete",
		fmt.Sprintf("removed=%d", len(pagesToRemove)),
		fmt.Sprintf("remaining=%d", result.RemainingPages))

	return result, nil
}

// newConfiguration crea una configuración pdfcpu con valores del config.
func (p *Processor) newConfiguration() *model.Configuration {
	conf := model.NewDefaultConfiguration()

	mode := model.ValidationStrict
	if p.config.ValidationMode == "relaxed" {
		mode = model.ValidationRelaxed
	}
	conf.ValidationMode = mode

	return conf
}

// ensureOutputDir crea el directorio de salida si es necesario.
func ensureOutputDir(outputPath string) error {
	outDir := filepath.Dir(outputPath)
	if outDir == "" || outDir == "." {
		return nil
	}

	// Validar que la ruta no contenga path traversal
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}
