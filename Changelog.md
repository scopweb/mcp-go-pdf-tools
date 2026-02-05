# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-03

### Added
- **PDF Remove Pages Tool** (`pdf_remove_pages`)
  - New function `RemovePagesFromFile()` in `internal/pdf/pages.go`
  - Supports two modes: `remove` (delete listed pages) and `keep` (keep only listed pages, delete the rest)
  - Page selection with ranges: `2,5-8,11` syntax
  - Full validation: out-of-bounds pages, invalid ranges, cannot remove all pages
  - Returns operation details: original pages, removed pages, remaining count, mode used

- **HTTP Endpoint for Remove Pages**
  - `POST /api/v1/pdf/remove-pages` endpoint
  - Accepts multipart/form-data with `file` (PDF), `pages` (selection string), `mode` (optional: `remove`/`keep`)
  - Returns the resulting PDF as download

- **MCP Tool for Remove Pages**
  - New `pdf_remove_pages` tool in MCP server
  - Parameters: `pdf_path`, `output_path`, `pages`, `mode` (optional)
  - Integrated with Claude Desktop MCP protocol

- **CLI Command for Remove Pages**
  - `cli remove-pages -i <input.pdf> -o <output.pdf> -pages <selection> [-mode remove|keep]`

### Changed
- Updated Go dependencies:
  - `clipperhouse/uax29/v2` v2.3.0 → v2.5.0
  - `golang.org/x/crypto` v0.46.0 → v0.47.0
  - `golang.org/x/image` v0.34.0 → v0.35.0
  - `golang.org/x/text` v0.32.0 → v0.33.0
- Bumped server version to `0.2.0`

### Fixed
- **MCP Protocol Compliance**: Rewritten MCP server response format for all tools
  - Tool errors now return `result` with `isError: true` and `content[]` array (MCP spec) instead of raw JSON-RPC `error` without `code` field
  - Tool results now wrapped in `content[]` array as required by MCP protocol
  - JSON-RPC protocol errors (method not found) now include required `code` field
  - This fixes ZodError validation failures in Claude Desktop for all tools (pdf_split, pdf_info, pdf_compress, pdf_remove_pages)
- Refactored MCP server into separate handler functions per tool for better maintainability

## [0.1.0] - 2025-12-13

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
