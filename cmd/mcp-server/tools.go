package main

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
	"github.com/scopweb/mcp-go-pdf-tools/internal/types"
)

// ToolsRegistry gestiona las herramientas disponibles.
type ToolsRegistry struct {
	tools  map[string]*ToolHandler
	logger logging.Logger
}

// ToolHandler es la interfaz para un manejador de herramienta.
type ToolHandler interface {
	Handle(id RequestID, rawArgs json.RawMessage) *Response
	GetDefinition() Tool
}

// NewToolsRegistry crea un nuevo registro de herramientas.
func NewToolsRegistry(processor *pdf.Processor, logger logging.Logger) *ToolsRegistry {
	registry := &ToolsRegistry{
		tools:  make(map[string]*ToolHandler),
		logger: logger,
	}

	// Registrar herramientas
	registry.registerTool(&PDFSplitHandler{processor: processor, logger: logger})
	registry.registerTool(&PDFInfoHandler{processor: processor, logger: logger})
	registry.registerTool(&PDFCompressHandler{processor: processor, logger: logger})
	registry.registerTool(&PDFRemovePagesHandler{processor: processor, logger: logger})

	return registry
}

func (r *ToolsRegistry) registerTool(handler ToolHandler) {
	def := handler.GetDefinition()
	r.tools[def.Name] = &handler
}

// GetToolDefinitions retorna la lista de herramientas disponibles.
func (r *ToolsRegistry) GetToolDefinitions() []Tool {
	var tools []Tool
	for _, handler := range r.tools {
		if handler != nil {
			tools = append(tools, (*handler).GetDefinition())
		}
	}
	return tools
}

// CallTool llama una herramienta por nombre.
func (r *ToolsRegistry) CallTool(id RequestID, name string, rawArgs json.RawMessage) *Response {
	handler, ok := r.tools[name]
	if !ok {
		return NewToolErrorResult(id, fmt.Sprintf("unknown tool: %s", name))
	}
	if handler == nil {
		return NewToolErrorResult(id, fmt.Sprintf("tool handler not initialized: %s", name))
	}
	return (*handler).Handle(id, rawArgs)
}

// PDFSplitHandler maneja pdf_split
type PDFSplitHandler struct {
	processor *pdf.Processor
	logger    logging.Logger
}

type pdfsplitArgs struct {
	PDFPath   string `json:"pdf_path"`
	OutputDir string `json:"output_dir,omitempty"`
	Zip       bool   `json:"zip,omitempty"`
	ZipName   string `json:"zip_name,omitempty"`
	ZipB64    bool   `json:"zip_b64,omitempty"`
}

func (h *PDFSplitHandler) GetDefinition() Tool {
	return Tool{
		Name:        "pdf_split",
		Description: "Split a PDF into single-page PDFs and optionally create a ZIP archive",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pdf_path":   map[string]interface{}{"type": "string", "description": "Absolute path to input PDF"},
				"output_dir": map[string]interface{}{"type": "string", "description": "Optional directory to move page PDFs"},
				"zip":        map[string]interface{}{"type": "boolean", "description": "Create ZIP archive with parts (default false)"},
				"zip_name":   map[string]interface{}{"type": "string", "description": "Optional ZIP filename"},
				"zip_b64":    map[string]interface{}{"type": "boolean", "description": "Return ZIP content as base64 in response"},
			},
			"required":             []string{"pdf_path"},
			"additionalProperties": false,
		},
	}
}

func (h *PDFSplitHandler) Handle(id RequestID, rawArgs json.RawMessage) *Response {
	var args pdfsplitArgs
	if err := UnmarshalParams(rawArgs, &args); err != nil {
		h.logger.Error("failed to unmarshal pdf_split args", err)
		return NewToolErrorResult(id, fmt.Sprintf("invalid arguments: %v", err))
	}

	if strings.TrimSpace(args.PDFPath) == "" {
		return NewToolErrorResult(id, "missing or invalid pdf_path")
	}

	h.logger.Debug("executing pdf_split",
		fmt.Sprintf("pdf_path=%s", args.PDFPath),
		fmt.Sprintf("zip=%v", args.Zip))

	// Split PDF
	parts, err := h.processor.Split(args.PDFPath)
	if err != nil {
		h.logger.Error("pdf_split failed", err)
		return NewToolErrorResult(id, err.Error())
	}

	// Opcionalmente mover a output_dir
	if strings.TrimSpace(args.OutputDir) != "" {
		if err := os.MkdirAll(args.OutputDir, 0755); err == nil {
			var movedParts []string
			for _, p := range parts {
				_, name := filepath.Split(p)
				dst := filepath.Join(args.OutputDir, name)
				if err := os.Rename(p, dst); err != nil {
					h.logger.Warn("failed to move file", err)
				} else {
					movedParts = append(movedParts, dst)
				}
			}
			if len(movedParts) > 0 {
				parts = movedParts
			}
		}
	}

	result := map[string]interface{}{"files": parts}

	// Opcionalmente crear ZIP
	if args.Zip {
		zipName := args.ZipName
		if strings.TrimSpace(zipName) == "" {
			base := filepath.Base(args.PDFPath)
			zipName = base + "-split.zip"
		}

		zipPath := filepath.Join(filepath.Dir(parts[0]), zipName)

		if err := h.createZipArchive(zipPath, parts); err != nil {
			h.logger.Error("failed to create zip archive", err)
			return NewToolErrorResult(id, fmt.Sprintf("failed to create ZIP: %v", err))
		}

		result["zip"] = zipPath

		// Opcionalmente codificar como base64
		if args.ZipB64 {
			data, err := os.ReadFile(zipPath)
			if err != nil {
				h.logger.Error("failed to read zip file", err)
				return NewToolErrorResult(id, fmt.Sprintf("failed to read ZIP: %v", err))
			}
			result["zip_b64"] = base64.StdEncoding.EncodeToString(data)
		}
	}

	resultJSON, _ := json.Marshal(result)
	return NewToolResult(id, string(resultJSON))
}

func (h *PDFSplitHandler) createZipArchive(zipPath string, parts []string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	for _, p := range parts {
		fr, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open part %s: %w", p, err)
		}

		_, name := filepath.Split(p)
		fw, err := zw.Create(name)
		if err != nil {
			fr.Close()
			return fmt.Errorf("failed to create zip entry %s: %w", name, err)
		}

		if _, err := io.Copy(fw, fr); err != nil {
			fr.Close()
			return fmt.Errorf("failed to write zip entry %s: %w", name, err)
		}
		fr.Close()
	}

	return nil
}

// PDFInfoHandler maneja pdf_info
type PDFInfoHandler struct {
	processor *pdf.Processor
	logger    logging.Logger
}

type pdfInfoArgs struct {
	PDFPath string `json:"pdf_path"`
}

func (h *PDFInfoHandler) GetDefinition() Tool {
	return Tool{
		Name:        "pdf_info",
		Description: "Return basic PDF information (page count, file size)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pdf_path": map[string]interface{}{"type": "string", "description": "Path to the PDF file"},
			},
			"required":             []string{"pdf_path"},
			"additionalProperties": false,
		},
	}
}

func (h *PDFInfoHandler) Handle(id RequestID, rawArgs json.RawMessage) *Response {
	var args pdfInfoArgs
	if err := UnmarshalParams(rawArgs, &args); err != nil {
		h.logger.Error("failed to unmarshal pdf_info args", err)
		return NewToolErrorResult(id, fmt.Sprintf("invalid arguments: %v", err))
	}

	if strings.TrimSpace(args.PDFPath) == "" {
		return NewToolErrorResult(id, "missing or invalid pdf_path")
	}

	h.logger.Debug("executing pdf_info",
		fmt.Sprintf("pdf_path=%s", args.PDFPath))

	info, err := h.processor.GetInfo(args.PDFPath)
	if err != nil {
		h.logger.Error("pdf_info failed", err)
		return NewToolErrorResult(id, err.Error())
	}

	resultJSON, _ := json.Marshal(info)
	return NewToolResult(id, string(resultJSON))
}

// PDFCompressHandler maneja pdf_compress
type PDFCompressHandler struct {
	processor *pdf.Processor
	logger    logging.Logger
}

type pdfCompressArgs struct {
	PDFPath    string `json:"pdf_path"`
	OutputPath string `json:"output_path"`
}

func (h *PDFCompressHandler) GetDefinition() Tool {
	return Tool{
		Name:        "pdf_compress",
		Description: "Compress a PDF by optimizing images, removing metadata, and cleaning structure (30-70% reduction)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pdf_path":    map[string]interface{}{"type": "string", "description": "Absolute path to input PDF"},
				"output_path": map[string]interface{}{"type": "string", "description": "Absolute path where compressed PDF will be saved"},
			},
			"required":             []string{"pdf_path", "output_path"},
			"additionalProperties": false,
		},
	}
}

func (h *PDFCompressHandler) Handle(id RequestID, rawArgs json.RawMessage) *Response {
	var args pdfCompressArgs
	if err := UnmarshalParams(rawArgs, &args); err != nil {
		h.logger.Error("failed to unmarshal pdf_compress args", err)
		return NewToolErrorResult(id, fmt.Sprintf("invalid arguments: %v", err))
	}

	if strings.TrimSpace(args.PDFPath) == "" {
		return NewToolErrorResult(id, "missing or invalid pdf_path")
	}
	if strings.TrimSpace(args.OutputPath) == "" {
		return NewToolErrorResult(id, "missing or invalid output_path")
	}

	h.logger.Debug("executing pdf_compress",
		fmt.Sprintf("pdf_path=%s", args.PDFPath),
		fmt.Sprintf("output_path=%s", args.OutputPath))

	result, err := h.processor.Compress(args.PDFPath, args.OutputPath)
	if err != nil {
		h.logger.Error("pdf_compress failed", err)
		return NewToolErrorResult(id, err.Error())
	}

	resultJSON, _ := json.Marshal(result)
	return NewToolResult(id, string(resultJSON))
}

// PDFRemovePagesHandler maneja pdf_remove_pages
type PDFRemovePagesHandler struct {
	processor *pdf.Processor
	logger    logging.Logger
}

type pdfRemovePagesArgs struct {
	PDFPath    string `json:"pdf_path"`
	OutputPath string `json:"output_path"`
	Pages      string `json:"pages"`
	Mode       string `json:"mode,omitempty"`
}

func (h *PDFRemovePagesHandler) GetDefinition() Tool {
	return Tool{
		Name:        "pdf_remove_pages",
		Description: "Remove or keep specific pages from a PDF. Modes: 'remove' deletes listed pages, 'keep' keeps only listed pages",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pdf_path":    map[string]interface{}{"type": "string", "description": "Absolute path to input PDF"},
				"output_path": map[string]interface{}{"type": "string", "description": "Absolute path where the result will be saved"},
				"pages":       map[string]interface{}{"type": "string", "description": "Comma-separated pages or ranges: '2', '5-8', '2,5-8,11'"},
				"mode":        map[string]interface{}{"type": "string", "enum": []string{"remove", "keep"}, "description": "Operation mode (default: 'remove')"},
			},
			"required":             []string{"pdf_path", "output_path", "pages"},
			"additionalProperties": false,
		},
	}
}

func (h *PDFRemovePagesHandler) Handle(id RequestID, rawArgs json.RawMessage) *Response {
	var args pdfRemovePagesArgs
	if err := UnmarshalParams(rawArgs, &args); err != nil {
		h.logger.Error("failed to unmarshal pdf_remove_pages args", err)
		return NewToolErrorResult(id, fmt.Sprintf("invalid arguments: %v", err))
	}

	if strings.TrimSpace(args.PDFPath) == "" {
		return NewToolErrorResult(id, "missing or invalid pdf_path")
	}
	if strings.TrimSpace(args.OutputPath) == "" {
		return NewToolErrorResult(id, "missing or invalid output_path")
	}
	if strings.TrimSpace(args.Pages) == "" {
		return NewToolErrorResult(id, "missing or invalid pages")
	}

	// Parsear modo
	modeStr := strings.TrimSpace(args.Mode)
	if modeStr == "" {
		modeStr = "remove"
	}
	mode, ok := types.ParseMode(modeStr)
	if !ok {
		return NewToolErrorResult(id, "invalid mode: must be 'remove' or 'keep'")
	}

	h.logger.Debug("executing pdf_remove_pages",
		fmt.Sprintf("pdf_path=%s", args.PDFPath),
		fmt.Sprintf("mode=%s", mode),
		fmt.Sprintf("pages=%s", args.Pages))

	result, err := h.processor.RemovePages(args.PDFPath, args.OutputPath, args.Pages, mode)
	if err != nil {
		h.logger.Error("pdf_remove_pages failed", err)
		return NewToolErrorResult(id, err.Error())
	}

	resultJSON, _ := json.Marshal(result)
	return NewToolResult(id, string(resultJSON))
}
