package config

import (
	"os"
	"strconv"
	"time"
)

// PDFConfig contiene configuración para operaciones PDF.
type PDFConfig struct {
	// Compression
	ImageQuality int
	RemoveMetadata bool

	// Validation
	ValidationMode string

	// Temp directory for split operations
	TempDir string
}

// ServerConfig contiene configuración para el servidor HTTP.
type ServerConfig struct {
	// HTTP server settings
	Host string
	Port string

	// Timeouts
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Upload limits
	MaxUploadSize int64

	// Logging
	LogLevel  string
	LogFormat string // "text" o "json"

	// PDF operations
	PDF PDFConfig
}

// MCPServerConfig contiene configuración para el servidor MCP.
type MCPServerConfig struct {
	// Logging
	LogLevel  string
	LogFormat string

	// Buffer size for large payloads
	BufferSize int

	// Protocol versions to support (comma-separated)
	ProtocolVersions string

	// PDF operations
	PDF PDFConfig
}

// CLIConfig contiene configuración para la CLI.
type CLIConfig struct {
	// Logging
	LogLevel string

	// PDF operations
	PDF PDFConfig
}

// NewServerConfig crea una configuración del servidor desde variables de entorno.
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Host: getEnv("HTTP_HOST", "0.0.0.0"),
		Port: getEnv("HTTP_PORT", "8080"),
		ReadTimeout: getDuration("HTTP_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDuration("HTTP_WRITE_TIMEOUT", 120*time.Second),
		IdleTimeout: getDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
		MaxUploadSize: getInt64("HTTP_MAX_UPLOAD_SIZE", 200<<20), // 200MB
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),
		PDF: PDFConfig{
			ImageQuality: getInt("PDF_IMAGE_QUALITY", 75),
			RemoveMetadata: getBool("PDF_REMOVE_METADATA", true),
			ValidationMode: getEnv("PDF_VALIDATION_MODE", "relaxed"),
			TempDir: getEnv("PDF_TEMP_DIR", ""),
		},
	}
}

// NewMCPServerConfig crea una configuración del servidor MCP desde variables de entorno.
func NewMCPServerConfig() *MCPServerConfig {
	return &MCPServerConfig{
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),
		BufferSize: getInt("MCP_BUFFER_SIZE", 10<<20), // 10MB
		ProtocolVersions: getEnv("MCP_PROTOCOL_VERSIONS", "2025-11-25,2025-06-18,2025-03-26"),
		PDF: PDFConfig{
			ImageQuality: getInt("PDF_IMAGE_QUALITY", 75),
			RemoveMetadata: getBool("PDF_REMOVE_METADATA", true),
			ValidationMode: getEnv("PDF_VALIDATION_MODE", "relaxed"),
			TempDir: getEnv("PDF_TEMP_DIR", ""),
		},
	}
}

// NewCLIConfig crea una configuración de CLI desde variables de entorno.
func NewCLIConfig() *CLIConfig {
	return &CLIConfig{
		LogLevel: getEnv("LOG_LEVEL", "info"),
		PDF: PDFConfig{
			ImageQuality: getInt("PDF_IMAGE_QUALITY", 75),
			RemoveMetadata: getBool("PDF_REMOVE_METADATA", true),
			ValidationMode: getEnv("PDF_VALIDATION_MODE", "relaxed"),
			TempDir: getEnv("PDF_TEMP_DIR", ""),
		},
	}
}

// Utility functions
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultVal
}
