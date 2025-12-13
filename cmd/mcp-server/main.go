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
	"time"

	"github.com/scopweb/mcp-go-pdf-to-img/internal/pdf"
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

func handleRequest(req MCPRequest) *MCPResponse {
	switch req.Method {
	case "initialize":
		// Return a structured initialize result including capabilities.tools
		res := map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]string{
				"name":    "mcp-go-pdf-tools",
				"version": "0.1.0",
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
					"required": []string{"pdf_path"},
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
					"required": []string{"pdf_path"},
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
					"required": []string{"pdf_path", "output_path"},
				},
			},
		}
		return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: ToolsListResult{Tools: tools}}

	case "tools/call":
		// Unmarshal params into CallToolParams
		var params CallToolParams
		if paramBytes, err := json.Marshal(req.Params); err == nil {
			_ = json.Unmarshal(paramBytes, &params)
		}

		switch params.Name {
		case "pdf_split":
			pdfPathI, ok := params.Arguments["pdf_path"]
			if !ok {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "missing pdf_path"}}
			}
			pdfPath, ok := pdfPathI.(string)
			if !ok || strings.TrimSpace(pdfPath) == "" {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "invalid pdf_path"}}
			}

			parts, err := pdf.SplitPDFFile(pdfPath)
			if err != nil {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": err.Error()}}
			}

			// Optional move to output_dir
			if outDirI, ok := params.Arguments["output_dir"]; ok {
				if outDir, ok2 := outDirI.(string); ok2 && strings.TrimSpace(outDir) != "" {
					if err := os.MkdirAll(outDir, 0755); err == nil {
						for _, p := range parts {
							_, name := filepath.Split(p)
							dst := filepath.Join(outDir, name)
							_ = os.Rename(p, dst)
						}
						// refresh parts list
						var moved []string
						for _, p := range parts {
							_, name := filepath.Split(p)
							moved = append(moved, filepath.Join(outDir, name))
						}
						parts = moved
					}
				}
			}

			// Optionally create a zip archive of the parts
			// Parameters:
			//   zip: bool (create zip)
			//   zip_name: string (optional filename)
			//   zip_b64: bool (return base64 content)
			if zipI, ok := params.Arguments["zip"]; ok {
				if zipFlag, ok2 := zipI.(bool); ok2 && zipFlag {
					// determine zip name
					zipName := "split.zip"
					if zn, ok := params.Arguments["zip_name"].(string); ok && strings.TrimSpace(zn) != "" {
						zipName = zn
					} else {
						// default: <input>-split.zip
						base := filepath.Base(pdfPath)
						zipName = base + "-split.zip"
					}

					tmpDir := filepath.Dir(parts[0])
					zipPath := filepath.Join(tmpDir, zipName)

					// create zip with proper error handling
					if len(parts) == 0 {
						return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "no output files to zip"}}
					}

					zf, err := os.Create(zipPath)
					if err != nil {
						return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": fmt.Sprintf("failed to create zip file: %v", err)}}
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

					// close writers and check for errors
					if cerr := zw.Close(); cerr != nil && zipErr == nil {
						zipErr = fmt.Errorf("failed to close zip writer: %w", cerr)
					}
					if cerr := zf.Close(); cerr != nil && zipErr == nil {
						zipErr = fmt.Errorf("failed to close zip file: %w", cerr)
					}

					if zipErr != nil {
						// cleanup incomplete zip
						_ = os.Remove(zipPath)
						return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": zipErr.Error()}}
					}

					result := map[string]interface{}{"files": parts, "zip": zipPath}

					// optionally return base64 content
					if b64I, ok := params.Arguments["zip_b64"]; ok {
						if b64Flag, ok2 := b64I.(bool); ok2 && b64Flag {
							data, err := os.ReadFile(zipPath)
							if err != nil {
								// if reading fails, return error
								return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": fmt.Sprintf("failed to read zip for b64: %v", err)}}
							}
							result["zip_b64"] = base64.StdEncoding.EncodeToString(data)
						}
					}

					return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
				}
			}

			// Default: Return JSON list of file paths
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"files": parts}}

		case "pdf_info":
			pdfPathI, ok := params.Arguments["pdf_path"]
			if !ok {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "missing pdf_path"}}
			}
			pdfPath, ok := pdfPathI.(string)
			if !ok || strings.TrimSpace(pdfPath) == "" {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "invalid pdf_path"}}
			}

			info, err := pdf.GetPDFInfo(pdfPath)
			if err != nil {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": err.Error()}}
			}
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: info}

		case "pdf_compress":
			pdfPathI, ok := params.Arguments["pdf_path"]
			if !ok {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "missing pdf_path"}}
			}
			pdfPath, ok := pdfPathI.(string)
			if !ok || strings.TrimSpace(pdfPath) == "" {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "invalid pdf_path"}}
			}

			outputPathI, ok := params.Arguments["output_path"]
			if !ok {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "missing output_path"}}
			}
			outputPath, ok := outputPathI.(string)
			if !ok || strings.TrimSpace(outputPath) == "" {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "invalid output_path"}}
			}

			result, err := pdf.CompressPDFWithDefaults(pdfPath, outputPath)
			if err != nil {
				return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": err.Error()}}
			}
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: result}

		case "pdf_to_images":
			// Not implemented: inform the caller how to proceed.
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "pdf_to_images is not implemented in this server. Use the HTTP endpoint or the CLI tool to convert PDFs to images."}}

		default:
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "unknown tool: " + params.Name}}
		}

	case "notifications/initialized":
		// No response needed
		return nil

	default:
		if req.ID != nil {
			return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]interface{}{"message": "method not found"}}
		}
		return nil
	}
}

func main() {
	logger := log.New(os.Stderr, "[MCP] ", log.LstdFlags)
	logger.Printf("Starting mcp stdio server")

	scanner := bufio.NewScanner(os.Stdin)
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

		resp := handleRequest(req)
		if resp != nil {
			b, err := json.Marshal(resp)
			if err != nil {
				logger.Printf("failed to marshal response: %v", err)
				continue
			}
			// Write response to stdout followed by newline
			fmt.Println(string(b))
		}
	}

	// If we reached EOF (client closed its write end) keep the process alive
	// for a short period so Claude has time to read any final output. This
	// helps debugging when the client closes unexpectedly.
	if err := scanner.Err(); err != nil && err != io.EOF {
		logger.Printf("scanner error: %v", err)
	} else {
		logger.Printf("stdin closed (EOF). Keeping process alive for 30s for debugging.")
		// Wait 30s then exit; this avoids immediate termination so Claude can
		// read any pending stdout/stderr messages.
		time.Sleep(30 * time.Second)
	}
}
