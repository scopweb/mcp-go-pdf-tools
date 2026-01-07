package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"log"

	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/pdf/split", splitHandler)
	mux.HandleFunc("/api/v1/pdf/compress", compressHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	log.Printf("Starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// splitHandler accepts multipart/form-data with a file field `file`.
// It splits the PDF into separate files and returns a ZIP archive.
func splitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 200MB)
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		http.Error(w, "invalid multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		http.Error(w, "failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Save uploaded file to disk
	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		http.Error(w, "failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	// Split using internal/pdf
	parts, err := pdf.SplitPDFFile(tmpPath)
	if err != nil {
		http.Error(w, "failed to split PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Clean up parts dir after response
	if len(parts) == 0 {
		http.Error(w, "no pages produced", http.StatusInternalServerError)
		return
	}

	// Prepare zip response
	w.Header().Set("Content-Type", "application/zip")
	zipName := fmt.Sprintf("%s-split.zip", header.Filename)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipName+"\"")

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, p := range parts {
		f, err := os.Open(p)
		if err != nil {
			// log and continue
			log.Printf("warning: cannot open part %s: %v", p, err)
			continue
		}

		_, name := filepath.Split(p)
		fw, err := zw.Create(name)
		if err != nil {
			f.Close()
			log.Printf("warning: cannot create zip entry: %v", err)
			continue
		}

		if _, err := io.Copy(fw, f); err != nil {
			log.Printf("warning: cannot write zip entry: %v", err)
		}
		f.Close()
	}

	// cleanup temp parts directory
	if len(parts) > 0 {
		partsDir := filepath.Dir(parts[0])
		go func(dir string) {
			// small delay to let response finish streaming
			time.Sleep(2 * time.Second)
			_ = os.RemoveAll(dir)
		}(partsDir)
	}
}

// compressHandler accepts multipart/form-data with a file field `file`.
// It compresses the PDF and returns the compressed file as download.
func compressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 200MB)
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		http.Error(w, "invalid multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temp input file
	tmpInputFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		http.Error(w, "failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpInputPath := tmpInputFile.Name()
	defer os.Remove(tmpInputPath)

	// Save uploaded file to disk
	if _, err := io.Copy(tmpInputFile, file); err != nil {
		tmpInputFile.Close()
		http.Error(w, "failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpInputFile.Close()

	// Create temp output file
	tmpOutputFile, err := os.CreateTemp("", "compressed-*.pdf")
	if err != nil {
		http.Error(w, "failed to create output temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpOutputPath := tmpOutputFile.Name()
	tmpOutputFile.Close()
	defer os.Remove(tmpOutputPath)

	// Compress the PDF
	result, err := pdf.CompressPDFWithDefaults(tmpInputPath, tmpOutputPath)
	if err != nil {
		http.Error(w, "failed to compress PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read the compressed file
	compressedFile, err := os.Open(tmpOutputPath)
	if err != nil {
		http.Error(w, "failed to open compressed file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer compressedFile.Close()

	// Send compressed PDF as download
	w.Header().Set("Content-Type", "application/pdf")
	compressedName := fmt.Sprintf("%s-compressed.pdf", header.Filename[:len(header.Filename)-4])
	w.Header().Set("Content-Disposition", "attachment; filename=\""+compressedName+"\"")

	// Send compression info as headers
	w.Header().Set("X-Original-Size", fmt.Sprintf("%d", result["original_size"]))
	w.Header().Set("X-Compressed-Size", fmt.Sprintf("%d", result["compressed_size"]))
	w.Header().Set("X-Reduction-Percent", fmt.Sprintf("%.1f%%", result["reduction_percent"]))

	if _, err := io.Copy(w, compressedFile); err != nil {
		log.Printf("error writing compressed PDF response: %v", err)
	}
}
