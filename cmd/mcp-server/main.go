package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

func main() {
	// Cargar configuración
	cfg := config.NewMCPServerConfig()

	// Inicializar logger
	var logger logging.Logger
	if cfg.LogFormat == "json" {
		logger = logging.NewJSON(cfg.LogLevel)
	} else {
		logger = logging.New(cfg.LogLevel)
	}

	// Inicializar procesador PDF
	processor := pdf.NewProcessor(cfg.PDF, logger)

	// Crear servidor MCP
	server := NewMCPServer(processor, logger)

	logger.Info("starting MCP stdio server",
		fmt.Sprintf("log_level=%s", cfg.LogLevel),
		fmt.Sprintf("supported_versions=2025-11-25,2025-06-18,2025-03-26"))

	// Escáner con buffer grande para payloads grandes
	scanner := bufio.NewScanner(os.Stdin)
	// Buffer de 10MB para manejar grandes payloads JSON-RPC (ej: base64)
	scanner.Buffer(make([]byte, 0, 64*1024), cfg.BufferSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parsear solicitud
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			logger.Error("failed to parse JSON", err)
			continue
		}

		logger.Debug("received request",
			fmt.Sprintf("method=%s", req.Method),
			fmt.Sprintf("id=%v", req.ID))

		// Validar ID requerido
		if req.ID == nil && RequiresID(req.Method) {
			logger.Warn("request missing required id",
				fmt.Sprintf("method=%s", req.Method))
			continue
		}

		// Procesar solicitud
		resp := server.HandleRequest(&req)

		// Enviar respuesta si es necesario
		if resp != nil {
			respBytes, err := json.Marshal(resp)
			if err != nil {
				logger.Error("failed to marshal response", err)
				continue
			}
			fmt.Println(string(respBytes))
		}
	}

	// Manejar errores del escáner
	if err := scanner.Err(); err != nil && err != io.EOF {
		logger.Error("scanner error", err)
	} else {
		logger.Info("stdin closed (EOF). Shutting down.")
	}
}
