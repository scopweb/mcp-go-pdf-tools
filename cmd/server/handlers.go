package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
	"github.com/scopweb/mcp-go-pdf-tools/internal/types"
	"github.com/scopweb/mcp-go-pdf-tools/internal/util"
)

// Handlers encapsula todos los handlers HTTP con sus dependencias.
type Handlers struct {
	processor *pdf.Processor
	logger    logging.Logger
	config    *config.ServerConfig
}

// NewHandlers crea un nuevo set de handlers.
func NewHandlers(processor *pdf.Processor, logger logging.Logger, config *config.ServerConfig) *Handlers {
	return &Handlers{
		processor: processor,
		logger:    logger,
		config:    config,
	}
}

// Health responde con status OK.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Split divide un PDF en múltiples archivos y devuelve un ZIP.
func (h *Handlers) Split(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(h.config.MaxUploadSize); err != nil {
		h.logger.Error("failed to parse multipart form", err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("missing file field", err)
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Crear archivo temporal para upload
	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		h.logger.Error("failed to create temp file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Guardar archivo
	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		h.logger.Error("failed to save uploaded file", err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	// Dividir PDF
	parts, err := h.processor.Split(tmpPath)
	if err != nil {
		h.logger.Error("PDF split failed", err)
		http.Error(w, "failed to split PDF", http.StatusBadRequest)
		return
	}

	if len(parts) == 0 {
		h.logger.Warn("no pages produced")
		http.Error(w, "no pages produced", http.StatusInternalServerError)
		return
	}

	// Preparar respuesta ZIP
	w.Header().Set("Content-Type", "application/zip")
	zipName := sanitizeFilename(filepath.Base(header.Filename)) + "-split.zip"
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))

	cleaner := util.NewResourceCleaner(h.logger)
	zw := zip.NewWriter(w)

	for _, p := range parts {
		f, err := os.Open(p)
		if err != nil {
			h.logger.Warn("cannot open part file", err)
			continue
		}

		_, name := filepath.Split(p)
		fw, err := zw.Create(name)
		if err != nil {
			f.Close()
			h.logger.Warn("cannot create zip entry", err)
			continue
		}

		if _, err := io.Copy(fw, f); err != nil {
			h.logger.Warn("cannot write zip entry", err)
		}
		f.Close()
	}
	zw.Close()

	// Cleanup temporal después de respuesta
	if len(parts) > 0 {
		partsDir := filepath.Dir(parts[0])
		cleaner.AddDirectory(partsDir)
		cleaner.CleanupASAPAfterResponse(500 * time.Millisecond)
	}
}

// RemovePages elimina o conserva páginas específicas de un PDF.
func (h *Handlers) RemovePages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(h.config.MaxUploadSize); err != nil {
		h.logger.Error("failed to parse multipart form", err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("missing file field", err)
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validar páginas
	pageSelection := r.FormValue("pages")
	if pageSelection == "" {
		http.Error(w, "missing pages field", http.StatusBadRequest)
		return
	}

	// Parse modo
	modeStr := r.FormValue("mode")
	if modeStr == "" {
		modeStr = "remove"
	}

	mode, ok := types.ParseMode(modeStr)
	if !ok {
		http.Error(w, "invalid mode: must be 'remove' or 'keep'", http.StatusBadRequest)
		return
	}

	// Crear archivo temporal de entrada
	tmpInputFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		h.logger.Error("failed to create temp file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpInputPath := tmpInputFile.Name()
	defer os.Remove(tmpInputPath)

	if _, err := io.Copy(tmpInputFile, file); err != nil {
		tmpInputFile.Close()
		h.logger.Error("failed to save uploaded file", err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	tmpInputFile.Close()

	// Crear archivo temporal de salida
	tmpOutputFile, err := os.CreateTemp("", "removed-pages-*.pdf")
	if err != nil {
		h.logger.Error("failed to create output temp file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpOutputPath := tmpOutputFile.Name()
	tmpOutputFile.Close()
	defer os.Remove(tmpOutputPath)

	// Remover páginas
	result, err := h.processor.RemovePages(tmpInputPath, tmpOutputPath, pageSelection, mode)
	if err != nil {
		h.logger.Error("page removal failed", err)
		http.Error(w, "failed to remove pages: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Abrir archivo resultante
	resultFile, err := os.Open(tmpOutputPath)
	if err != nil {
		h.logger.Error("failed to open result file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer resultFile.Close()

	// Enviar resultado
	w.Header().Set("Content-Type", "application/pdf")
	resultName := sanitizeFilename(filepath.Base(header.Filename)) + "-pages-removed.pdf"
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", resultName))

	// Enviar metadata como headers
	w.Header().Set("X-Removed-Pages", fmt.Sprintf("%d", result.RemovedCount))
	w.Header().Set("X-Remaining-Pages", fmt.Sprintf("%d", result.RemainingPages))
	w.Header().Set("X-Mode", string(result.Mode))

	if _, err := io.Copy(w, resultFile); err != nil {
		h.logger.Error("error writing response", err)
	}
}

// Compress comprime un PDF y lo devuelve como descarga.
func (h *Handlers) Compress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(h.config.MaxUploadSize); err != nil {
		h.logger.Error("failed to parse multipart form", err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("missing file field", err)
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Crear archivo temporal de entrada
	tmpInputFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		h.logger.Error("failed to create temp file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpInputPath := tmpInputFile.Name()
	defer os.Remove(tmpInputPath)

	if _, err := io.Copy(tmpInputFile, file); err != nil {
		tmpInputFile.Close()
		h.logger.Error("failed to save uploaded file", err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	tmpInputFile.Close()

	// Crear archivo temporal de salida
	tmpOutputFile, err := os.CreateTemp("", "compressed-*.pdf")
	if err != nil {
		h.logger.Error("failed to create output temp file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpOutputPath := tmpOutputFile.Name()
	tmpOutputFile.Close()
	defer os.Remove(tmpOutputPath)

	// Comprimir PDF
	result, err := h.processor.Compress(tmpInputPath, tmpOutputPath)
	if err != nil {
		h.logger.Error("compression failed", err)
		http.Error(w, "failed to compress PDF", http.StatusInternalServerError)
		return
	}

	// Abrir archivo comprimido
	compressedFile, err := os.Open(tmpOutputPath)
	if err != nil {
		h.logger.Error("failed to open compressed file", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer compressedFile.Close()

	// Enviar resultado
	w.Header().Set("Content-Type", "application/pdf")
	compressedName := sanitizeFilename(filepath.Base(header.Filename)) + "-compressed.pdf"
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", compressedName))

	// Enviar metadata como headers
	w.Header().Set("X-Original-Size", fmt.Sprintf("%d", result.OriginalSize))
	w.Header().Set("X-Compressed-Size", fmt.Sprintf("%d", result.CompressedSize))
	w.Header().Set("X-Compression-Ratio", fmt.Sprintf("%.2f%%", result.CompressionRatio*100))

	if _, err := io.Copy(w, compressedFile); err != nil {
		h.logger.Error("error writing response", err)
	}
}

// sanitizeFilename limpia un nombre de archivo para evitar caracteres problemáticos.
func sanitizeFilename(filename string) string {
	// Remover extensión si existe
	if ext := filepath.Ext(filename); ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	// Remover caracteres problemáticos
	return filepath.Base(filename)
}
