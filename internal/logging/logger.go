package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger es la interfaz centralizada para logging.
type Logger interface {
	Debug(msg string, attrs ...slog.Attr)
	Info(msg string, attrs ...slog.Attr)
	Warn(msg string, attrs ...slog.Attr)
	Error(msg string, err error, attrs ...slog.Attr)
}

// SlogLogger es una implementación de Logger usando slog.
type SlogLogger struct {
	logger *slog.Logger
}

// New crea un nuevo logger con el nivel especificado.
// level puede ser "debug", "info", "warn", "error".
func New(level string) *SlogLogger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	return &SlogLogger{logger: logger}
}

// NewJSON crea un logger con formato JSON.
func NewJSON(level string) *SlogLogger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)

	return &SlogLogger{logger: logger}
}

// Debug registra un mensaje de nivel DEBUG.
func (l *SlogLogger) Debug(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

// Info registra un mensaje de nivel INFO.
func (l *SlogLogger) Info(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// Warn registra un mensaje de nivel WARN.
func (l *SlogLogger) Warn(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

// Error registra un mensaje de nivel ERROR con error adjunto.
func (l *SlogLogger) Error(msg string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

// WithAttrs retorna un logger con atributos adicionales (para contexto).
func (l *SlogLogger) WithAttrs(attrs ...slog.Attr) *SlogLogger {
	return &SlogLogger{logger: l.logger.With(attrs...)}
}

// contextKey es el tipo para las claves del contexto.
type contextKey string

const loggerKey contextKey = "logger"

// WithContext adjunta el logger al contexto.
func WithContext(ctx context.Context, logger *SlogLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext obtiene el logger del contexto, o retorna un logger por defecto.
func FromContext(ctx context.Context) *SlogLogger {
	if logger, ok := ctx.Value(loggerKey).(*SlogLogger); ok {
		return logger
	}
	return New("info")
}
