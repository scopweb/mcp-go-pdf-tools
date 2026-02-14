package types

// PageRemovalMode define el modo de eliminación de páginas.
type PageRemovalMode string

const (
	ModeRemove PageRemovalMode = "remove" // Elimina las páginas especificadas
	ModeKeep   PageRemovalMode = "keep"   // Conserva solo las páginas especificadas
)

// String retorna la representación en string del modo.
func (m PageRemovalMode) String() string {
	return string(m)
}

// IsValid verifica si el modo es válido.
func (m PageRemovalMode) IsValid() bool {
	return m == ModeRemove || m == ModeKeep
}

// ParseMode convierte un string a PageRemovalMode.
func ParseMode(s string) (PageRemovalMode, bool) {
	switch s {
	case "remove":
		return ModeRemove, true
	case "keep":
		return ModeKeep, true
	default:
		return "", false
	}
}

// SplitResult contiene el resultado de una operación de división.
type SplitResult struct {
	Files      []string `json:"files"`
	TotalPages int      `json:"total_pages"`
	OutputDir  string   `json:"output_dir"`
	ZipPath    string   `json:"zip_path,omitempty"`
	ZipB64     string   `json:"zip_b64,omitempty"`
}

// PDFInfoResult contiene información sobre un PDF.
type PDFInfoResult struct {
	TotalPages int   `json:"total_pages"`
	SizeBytes  int64 `json:"size_bytes"`
	Filename   string `json:"filename"`
}

// CompressResult contiene el resultado de una compresión.
type CompressResult struct {
	OutputPath       string `json:"output_path"`
	OriginalSize     int64  `json:"original_size"`
	CompressedSize   int64  `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// RemovePagesResult contiene el resultado de eliminación de páginas.
type RemovePagesResult struct {
	OutputPath     string         `json:"output_path"`
	OriginalPages  int            `json:"original_pages"`
	RemovedPages   []int          `json:"removed_pages"`
	RemovedCount   int            `json:"removed_count"`
	RemainingPages int            `json:"remaining_pages"`
	Mode           PageRemovalMode `json:"mode"`
}

// ToolResult es el resultado genérico de una herramienta MCP.
type ToolResult struct {
	Content string      `json:"content"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
