package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/scopweb/mcp-go-pdf-tools/internal/config"
	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

func main() {
	// Load configuration from environment
	cfg := config.NewServerConfig()

	// Initialize logger
	var logger logging.Logger
	if cfg.LogFormat == "json" {
		logger = logging.NewJSON(cfg.LogLevel)
	} else {
		logger = logging.New(cfg.LogLevel)
	}

	// Initialize PDF processor
	processor := pdf.NewProcessor(cfg.PDF, logger)

	// Initialize handlers
	handlers := NewHandlers(processor, logger, cfg)

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/v1/pdf/split", handlers.Split)
	mux.HandleFunc("/api/v1/pdf/compress", handlers.Compress)
	mux.HandleFunc("/api/v1/pdf/remove-pages", handlers.RemovePages)

	// Create HTTP server with configuration
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server
	logger.Info("starting HTTP server",
		fmt.Sprintf("addr=%s", addr),
		fmt.Sprintf("log_level=%s", cfg.LogLevel))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", err)
	}
}
