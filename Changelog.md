# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2025-12-13

### Added
- **PDF Compression Tool** (`pdf_compress`)
  - New function `CompressPDFFile()` in `internal/pdf/compress.go`
  - Optimizes PDFs by removing metadata, optimizing images, and cleaning structure
  - Reduces file size by 30-70% depending on content
  - Returns compression statistics (original/compressed size, reduction percentage)

- **HTTP Endpoint for Compression**
  - `POST /api/v1/pdf/compress` endpoint
  - Accepts multipart/form-data with PDF file
  - Returns compressed PDF with statistics in response headers
  - Headers: `X-Original-Size`, `X-Compressed-Size`, `X-Reduction-Percent`

- **MCP Tool for Compression**
  - New `pdf_compress` tool in MCP server
  - Full parameter validation and error handling
  - Integrated with Claude Desktop MCP protocol

### Changed
- Updated all Go dependencies to their latest versions using `go get -u ./...`.
- Updated `go.mod` and `go.sum` via `go mod tidy`.
- Refactored `internal/pdf` to replace deprecated `ioutil` with `os` and `io`.
- Improved `GetPDFInfo` to use `pdfcpu` for accurate page counting and validation.
- Added stronger input validation to PDF utility functions.
- Updated README with compression feature documentation and examples

### Fixed
- Fixed compilation error in `test/security` package where files were incorrectly declared as `package main`. Changed to `package security`.
