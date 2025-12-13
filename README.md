# mcp-go-pdf-tools

Servidor MCP sencillo para herramientas relacionadas con PDFs, escrito en Go. El objetivo principal de esta primera fase es ofrecer la funcionalidad de "separar un PDF en PDFs independientes (una página por archivo)" y exponerla como API HTTP y como camino para integrarlo con Claude Desktop (MCP).

**Estado:** prototipo funcional — separación por página, compresión y endpoints HTTP completados.

**Estructura importante**
- `cmd/server`: servidor HTTP (endpoints `GET /health`, `POST /api/v1/pdf/split`, `POST /api/v1/pdf/compress`).
- `cmd/mcp-server`: servidor stdio MCP con herramientas `pdf_split`, `pdf_info`, `pdf_compress`.
- `internal/pdf`: lógica para manipular PDFs (usa `pdfcpu`). Incluye funciones para split y compresión.
- `examples`: cliente de ejemplo para subir un PDF y guardar `split.zip`.
- `mcp`: notas y ejemplos para integrar con Claude Desktop.

**Funcionalidades**
- **pdf_split**: divide un PDF en archivos de una página cada uno. Puede crear un ZIP con las páginas.
- **pdf_info**: devuelve información básica del PDF (páginas, tamaño).
- **pdf_compress**: comprime un PDF optimizando imágenes, eliminando metadatos y limpiando la estructura. Reduce el tamaño de 30-70% según el contenido.

**Dependencias clave**
- `github.com/pdfcpu/pdfcpu` — usado para manipulación de PDFs (split, compresión, información).

**Quick Start (desarrollo)**

- Build y run (desarrollo):

```powershell
cd c:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\cmd\server
go run .
```

- Probar el endpoint `split` con `curl`:

```powershell
curl -F "file=@C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\test.pdf" http://localhost:8080/api/v1/pdf/split --output split.zip
```

- Probar el endpoint `compress` con `curl`:

```powershell
curl -F "file=@C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\test.pdf" http://localhost:8080/api/v1/pdf/compress --output compressed.pdf
```

- Ver información de compresión en headers:

```powershell
curl -I -F "file=@C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\test.pdf" http://localhost:8080/api/v1/pdf/compress
# Mostrará X-Original-Size, X-Compressed-Size, X-Reduction-Percent
```

- Cliente de ejemplo (Go):

```powershell
go run examples\upload.go C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\test.pdf
```

**Docker**

- Construir imagen local:

```powershell
docker build -t mcp-pdf-server:local .
```

- Ejecutar contenedor:

```powershell
docker run --rm -p 8080:8080 mcp-pdf-server:local
```

**Tests**

- Ejecutar tests:

```powershell
go test ./...
```

- Nota: `internal/pdf/split_test.go` es una prueba de integración y se salta si no existe `test.pdf` en la raíz del repo. Puedes usar cualquier PDF pequeño para testing local.

**Integración con Claude Desktop (MCP)**

Para integrar con Claude Desktop tienes dos opciones principales:

- **stdio (recomendado para Claude Desktop):** ejecutar un proceso `mcp-server.exe` que hable por `stdin/stdout` con Claude Desktop. Esto es lo que Claude espera para registrar herramientas automáticamente.
- **HTTP (alternativa):** ejecutar el servidor HTTP y usar un adaptador desde Claude (menos directo).

Ejemplo de fragmento para pegar en `claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "mcp-pdf-tools-stdio": {
      "command": "C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\bin\\mcp-server.exe",
      "args": ["--stdio"],
      "env": {}
    }
  }
}
```

Consejos:
- Revisa `mcp/README.md` y `mcp/claude_desktop_config_sample.json` para ejemplos y notas.
- El servidor MCP (`cmd/mcp-server`) ya está implementado y expone `pdf_split`, `pdf_info` y `pdf_compress` como herramientas.
- Compila con `go build ./cmd/mcp-server` y usa el binario resultante en `claude_desktop_config.json`.

**Roadmap y próximos pasos**

- Ver `Roadmap.md` en la raíz del repo para ver lo completado y lo pendiente (implementación stdio, tests adicionales, empaquetado de releases).

**Ejemplos de uso MCP (stdio)**

Ejemplo de llamada a `pdf_compress` con PowerShell:

```powershell
'{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"pdf_compress","arguments":{"pdf_path":"C:\\\\MCPs\\\\clone_PROYECTOS\\\\mcp-go-pdf-tools\\\\test.pdf","output_path":"C:\\\\MCPs\\\\clone_PROYECTOS\\\\mcp-go-pdf-tools\\\\test-compressed.pdf"}}}' | .\mcp-server.exe
```

Respuesta (ejemplo):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "original_size": 1500000,
    "compressed_size": 450000,
    "reduction_bytes": 1050000,
    "reduction_percent": 70.0,
    "output_file": "test-compressed.pdf",
    "output_path": "C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\test-compressed.pdf"
  }
}
```

**Contacto / ayuda**

- Para reportar issues o solicitar nuevas funcionalidades, abre un issue en el repositorio.


