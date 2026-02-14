package errors

import (
	"errors"
	"fmt"
)

// PdfError representa un error específico de operaciones PDF.
type PdfError struct {
	Op  string // Operación que falló (split, compress, etc.)
	Path string // Ruta del archivo afectado
	Err error  // Error subyacente
}

// Error implementa la interfaz error.
func (e *PdfError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap retorna el error subyacente para compatibilidad con errors.Is/As.
func (e *PdfError) Unwrap() error {
	return e.Err
}

// New crea un nuevo PdfError con operación y error.
func New(op string, err error) *PdfError {
	return &PdfError{Op: op, Err: err}
}

// NewWithPath crea un nuevo PdfError con operación, ruta y error.
func NewWithPath(op, path string, err error) *PdfError {
	return &PdfError{Op: op, Path: path, Err: err}
}

// Is permite comparar errores con errors.Is.
func (e *PdfError) Is(target error) bool {
	t, ok := target.(*PdfError)
	if !ok {
		return false
	}
	return e.Op == t.Op && e.Path == t.Path
}

// IsValidationError verifica si el error es de validación.
func IsValidationError(err error) bool {
	var target error
	if errors.As(err, &target) {
		return target.Error() == "validation error"
	}
	return false
}

// ValidationError representa un error de validación de entrada.
type ValidationError struct {
	Field   string
	Message string
}

// Error implementa la interfaz error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError crea un nuevo ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}
