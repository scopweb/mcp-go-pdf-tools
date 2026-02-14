package util

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
)

// ResourceCleaner gestiona la limpieza de recursos de manera segura.
type ResourceCleaner struct {
	dirs   []string
	files  []string
	mu     sync.Mutex
	logger logging.Logger
}

// NewResourceCleaner crea un nuevo ResourceCleaner.
func NewResourceCleaner(logger logging.Logger) *ResourceCleaner {
	return &ResourceCleaner{
		logger: logger,
	}
}

// AddDirectory marca un directorio para limpieza.
func (rc *ResourceCleaner) AddDirectory(path string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.dirs = append(rc.dirs, path)
}

// AddFile marca un archivo para limpieza.
func (rc *ResourceCleaner) AddFile(path string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.files = append(rc.files, path)
}

// CleanupSync ejecuta la limpieza de manera sincrónica.
func (rc *ResourceCleaner) CleanupSync() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var lastErr error

	// Eliminar archivos
	for _, f := range rc.files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			rc.logger.Warn("failed to remove file", err)
			lastErr = err
		}
	}

	// Eliminar directorios
	for _, d := range rc.dirs {
		if err := os.RemoveAll(d); err != nil && !os.IsNotExist(err) {
			rc.logger.Warn("failed to remove directory", err)
			lastErr = err
		}
	}

	// Limpiar listas
	rc.dirs = nil
	rc.files = nil

	return lastErr
}

// CleanupAsync ejecuta la limpieza de manera asincrónica con contexto.
// Si el contexto se cancela, la operación se detiene.
func (rc *ResourceCleaner) CleanupAsync(ctx context.Context, delay time.Duration) {
	go func() {
		select {
		case <-time.After(delay):
			if err := rc.CleanupSync(); err != nil {
				rc.logger.Error("cleanup error", err)
			}
		case <-ctx.Done():
			rc.logger.Debug("cleanup cancelled")
		}
	}()
}

// CleanupASAPAfterResponse limpia después de que se haya escrito la respuesta.
// Este patrón es útil para respuestas HTTP donde queremos limpiar después de enviar datos.
func (rc *ResourceCleaner) CleanupASAPAfterResponse(delay time.Duration) {
	go func() {
		time.Sleep(delay)
		if err := rc.CleanupSync(); err != nil {
			rc.logger.Error("async cleanup error", err)
		}
	}()
}
