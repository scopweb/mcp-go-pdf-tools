package main

import (
	"archive/zip"
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

// Minimal MCP/stdio protocol structures (JSON-RPC style)
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// toolResult builds a successful MCP tools/call response with content array.
func toolResult(id interface{}, data interface{}) *MCPResponse {
	text, _ := json.Marshal(data)
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(text)},
			},
		},
	}
}

// toolError builds an MCP tools/call error response (isError: true).
func toolError(id interface{}, msg string) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": msg},
			},
			"isError": true,
		},
	}
}

// jsonRPCError builds a JSON-RPC 2.0 error response with the required code field.
func jsonRPCError(id interface{}, code int, msg string) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": msg,
		},
	}
}

// supportedVersions lists protocol versions this server supports, newest first.
var supportedVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26"}

func negotiateVersion(clientVersion string) string {
	// If client requests a version we support, echo it back (MUST per spec).
	for _, v := range supportedVersions {
		if v == clientVersion {
			return v
		}
	}
	// Otherwise, respond with our latest supported version (SHOULD per spec).
	return supportedVersions[0]
}

func handleRequest(req MCPRequest) *MCPResponse {
	switch req.Method {
	case "initialize":
		// Extract client's protocolVersion from params for version negotiation.
		clientVersion := ""
		if paramsMap, ok := req.Params.(map[string]interface{}); ok {
			if pv, ok := paramsMap["protocolVersion"].(string); ok {
				clientVersion = pv
			}
		}
		negotiated := negotiateVersion(clientVersion)

		res := map[string]interface{}{
			"protocolVersion": negotiated,
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]string{
				"name":    "mcp-go-pdf-tools",
				"version": "0.2.1",
			},
		}
		return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: res}

	case "tools/list":
		tools := []Tool{
			{
				Name:        "pdf_split",
				Description: "Split a PDF into single-page PDFs and optionally produce a ZIP",
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
			},
			{
				Name:        "pdf_info",
				Description: "Return basic PDF info (pages, size, dimensions)",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pdf_path": map[string]interface{}{"type": "string", "description": "Path to the PDF file"},
					},
					"required":             []string{"pdf_path"},
					"additionalProperties": false,
				},
			},
			{
				Name:        "pdf_compress",
				Description: "Compress a PDF file by optimizing images, removing metadata, and cleaning structure. Reduces file size by 30-70% depending on content.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pdf_path":    map[string]interface{}{"type": "string", "description": "Absolute path to input PDF"},
						"output_path": map[string]interface{}{"type": "string", "description": "Absolute path where compressed PDF will be saved"},
					},
					"required":             []string{"pdf_path", "output_path"},
					"additionalProperties": false,
				},
			},
			{
				Name:        "pdf_remove_pages",
				Description: "Remove or keep specific pages from a PDF. Supports page ranges like '2,5-8,11'. Two modes: 'remove' deletes the listed pages; 'keep' keeps only the listed pages and deletes the rest.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pdf_path":    map[string]interface{}{"type": "string", "description": "Absolute path to input PDF"},
						"output_path": map[string]interface{}{"type": "string", "description": "Absolute path where the resulting PDF will be saved"},
						"pages":       map[string]interface{}{"type": "string", "description": "Comma-separated page numbers or ranges to remove/keep. Examples: '2,5-8,11', '1-3', '5'"},
						"mode":        map[string]interface{}{"type": "string", "enum": []string{"remove", "keep"}, "description": "Operation mode: 'remove' deletes listed pages (default), 'keep' keeps only listed pages and deletes the rest"},
					},
					"required":             []string{"pdf_path", "output_path", "pages"},
					"additionalProperties": false,
				},
			},
		}
		return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: ToolsListResult{Tools: tools}}

	case "tools/call":
		var params CallToolParams
		if paramBytes, err := json.Marshal(req.Params); err == nil {
			_ = json.Unmarshal(paramBytes, &params)
		}

		switch params.Name {
		case "pdf_split":
			return handlePdfSplit(req.ID, params.Arguments)
		case "pdf_info":
			return handlePdfInfo(req.ID, params.Arguments)
		case "pdf_compress":
			return handlePdfCompress(req.ID, params.Arguments)
		case "pdf_remove_pages":
			return handlePdfRemovePages(req.ID, params.Arguments)
		case "pdf_to_images":
			return toolError(req.ID, "pdf_to_images is not implemented in this server. Use the HTTP endpoint or the CLI tool to convert PDFs to images.")
		default:
			return toolError(req.ID, "unknown tool: "+params.Name)
		}

	case "ping":
		return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}

	case "notifications/initialized", "notifications/cancelled":
		return nil

	default:
		if req.ID != nil {
			return jsonRPCError(req.ID, -32601, "method not found")
		}
		return nil
	}
}

func handlePdfSplit(id interface{}, args map[string]interface{}) *MCPResponse {
	pdfPathI, ok := args["pdf_path"]
	if !ok {
		return toolError(id, "missing pdf_path")
	}
	pdfPath, ok := pdfPathI.(string)
	if !ok || strings.TrimSpace(pdfPath) == "" {
		return toolError(id, "invalid pdf_path")
	}

	parts, err := pdf.SplitPDFFile(pdfPath)
	if err != nil {
		return toolError(id, err.Error())
	}

	// Optional move to output_dir
	if outDirI, ok := args["output_dir"]; ok {
		if outDir, ok2 := outDirI.(string); ok2 && strings.TrimSpace(outDir) != "" {
			if err := os.MkdirAll(outDir, 0755); err == nil {
				for _, p := range parts {
					_, name := filepath.Split(p)
					dst := filepath.Join(outDir, name)
					_ = os.Rename(p, dst)
				}
				var moved []string
				for _, p := range parts {
					_, name := filepath.Split(p)
					moved = append(moved, filepath.Join(outDir, name))
				}
				parts = moved
			}
		}
	}

	// Optionally create a zip archive
	if zipI, ok := args["zip"]; ok {
		if zipFlag, ok2 := zipI.(bool); ok2 && zipFlag {
			zipName := "split.zip"
			if zn, ok := args["zip_name"].(string); ok && strings.TrimSpace(zn) != "" {
				zipName = zn
			} else {
				base := filepath.Base(pdfPath)
				zipName = base + "-split.zip"
			}

			tmpDir := filepath.Dir(parts[0])
			zipPath := filepath.Join(tmpDir, zipName)

			if len(parts) == 0 {
				return toolError(id, "no output files to zip")
			}

			zf, err := os.Create(zipPath)
			if err != nil {
				return toolError(id, fmt.Sprintf("failed to create zip file: %v", err))
			}

			zw := zip.NewWriter(zf)
			var zipErr error
			for _, p := range parts {
				fr, err := os.Open(p)
				if err != nil {
					zipErr = fmt.Errorf("failed to open part %s: %w", p, err)
					break
				}
				_, name := filepath.Split(p)
				fw, err := zw.Create(name)
				if err != nil {
					fr.Close()
					zipErr = fmt.Errorf("failed to create entry for %s: %w", name, err)
					break
				}
				if _, err := io.Copy(fw, fr); err != nil {
					fr.Close()
					zipErr = fmt.Errorf("failed to write entry %s: %w", name, err)
					break
				}
				fr.Close()
			}

			if cerr := zw.Close(); cerr != nil && zipErr == nil {
				zipErr = fmt.Errorf("failed to close zip writer: %w", cerr)
			}
			if cerr := zf.Close(); cerr != nil && zipErr == nil {
				zipErr = fmt.Errorf("failed to close zip file: %w", cerr)
			}

			if zipErr != nil {
				_ = os.Remove(zipPath)
				return toolError(id, zipErr.Error())
			}

			result := map[string]interface{}{"files": parts, "zip": zipPath}

			if b64I, ok := args["zip_b64"]; ok {
				if b64Flag, ok2 := b64I.(bool); ok2 && b64Flag {
					data, err := os.ReadFile(zipPath)
					if err != nil {
						return toolError(id, fmt.Sprintf("failed to read zip for b64: %v", err))
					}
					result["zip_b64"] = base64.StdEncoding.EncodeToString(data)
				}
			}

			return toolResult(id, result)
		}
	}

	return toolResult(id, map[string]interface{}{"files": parts})
}

func handlePdfInfo(id interface{}, args map[string]interface{}) *MCPResponse {
	pdfPathI, ok := args["pdf_path"]
	if !ok {
		return toolError(id, "missing pdf_path")
	}
	pdfPath, ok := pdfPathI.(string)
	if !ok || strings.TrimSpace(pdfPath) == "" {
		return toolError(id, "invalid pdf_path")
	}

	info, err := pdf.GetPDFInfo(pdfPath)
	if err != nil {
		return toolError(id, err.Error())
	}
	return toolResult(id, info)
}

func handlePdfCompress(id interface{}, args map[string]interface{}) *MCPResponse {
	pdfPathI, ok := args["pdf_path"]
	if !ok {
		return toolError(id, "missing pdf_path")
	}
	pdfPath, ok := pdfPathI.(string)
	if !ok || strings.TrimSpace(pdfPath) == "" {
		return toolError(id, "invalid pdf_path")
	}

	outputPathI, ok := args["output_path"]
	if !ok {
		return toolError(id, "missing output_path")
	}
	outputPath, ok := outputPathI.(string)
	if !ok || strings.TrimSpace(outputPath) == "" {
		return toolError(id, "invalid output_path")
	}

	result, err := pdf.CompressPDFWithDefaults(pdfPath, outputPath)
	if err != nil {
		return toolError(id, err.Error())
	}
	return toolResult(id, result)
}

func handlePdfRemovePages(id interface{}, args map[string]interface{}) *MCPResponse {
	pdfPathI, ok := args["pdf_path"]
	if !ok {
		return toolError(id, "missing pdf_path")
	}
	pdfPath, ok := pdfPathI.(string)
	if !ok || strings.TrimSpace(pdfPath) == "" {
		return toolError(id, "invalid pdf_path")
	}

	outputPathI, ok := args["output_path"]
	if !ok {
		return toolError(id, "missing output_path")
	}
	outputPath, ok := outputPathI.(string)
	if !ok || strings.TrimSpace(outputPath) == "" {
		return toolError(id, "invalid output_path")
	}

	pagesI, ok := args["pages"]
	if !ok {
		return toolError(id, "missing pages")
	}
	pages, ok := pagesI.(string)
	if !ok || strings.TrimSpace(pages) == "" {
		return toolError(id, "invalid pages")
	}

	keepMode := false
	if modeI, ok := args["mode"]; ok {
		if modeStr, ok2 := modeI.(string); ok2 && strings.TrimSpace(modeStr) == "keep" {
			keepMode = true
		}
	}

	result, err := pdf.RemovePagesFromFile(pdfPath, outputPath, pages, keepMode)
	if err != nil {
		return toolError(id, err.Error())
	}
	return toolResult(id, result)
}

func main() {
	logger := log.New(os.Stderr, "[MCP] ", log.LstdFlags)
	logger.Printf("Starting mcp stdio server")

	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer to 10MB to handle large JSON-RPC messages (e.g. base64 payloads).
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			logger.Printf("invalid json: %v", err)
			continue
		}

		logger.Printf("request: %s", line)

		// Validate: requests (with method expecting response) MUST have non-null id.
		// Notifications (method starts with "notifications/") are allowed without id.
		if req.ID == nil && !strings.HasPrefix(req.Method, "notifications/") && req.Method != "" {
			// Per JSON-RPC 2.0, a request without id is a notification — but MCP
			// methods like initialize, tools/list, tools/call require an id.
			logger.Printf("request missing id for method: %s (treating as notification, skipping)", req.Method)
			continue
		}

		resp := handleRequest(req)
		if resp != nil {
			b, err := json.Marshal(resp)
			if err != nil {
				logger.Printf("failed to marshal response: %v", err)
				continue
			}
			fmt.Println(string(b))
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		logger.Printf("scanner error: %v", err)
	} else {
		logger.Printf("stdin closed (EOF). Shutting down.")
	}
}
