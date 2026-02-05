# mcp-go-pdf-tools

Servidor MCP sencillo para herramientas relacionadas con PDFs, escrito en Go. Ofrece funcionalidades de split, compresion, informacion y eliminacion/seleccion de paginas, expuestas como API HTTP, CLI y como servidor MCP para Claude Desktop.

**Estado:** prototipo funcional — separacion por pagina, compresion, eliminacion de paginas y endpoints HTTP completados.

**Estructura importante**
- `cmd/server`: servidor HTTP (endpoints `GET /health`, `POST /api/v1/pdf/split`, `POST /api/v1/pdf/compress`, `POST /api/v1/pdf/remove-pages`).
- `cmd/mcp-server`: servidor stdio MCP con herramientas `pdf_split`, `pdf_info`, `pdf_compress`, `pdf_remove_pages`.
- `cmd/cli`: herramienta de linea de comandos con subcomandos `split` y `remove-pages`.
- `internal/pdf`: logica para manipular PDFs (usa `pdfcpu`). Incluye funciones para split, compresion y eliminacion de paginas.
- `examples`: cliente de ejemplo para subir un PDF y guardar `split.zip`.
- `mcp`: notas y ejemplos para integrar con Claude Desktop.

## Funcionalidades

### pdf_split
Divide un PDF en archivos de una pagina cada uno. Puede crear un ZIP con las paginas.

### pdf_info
Devuelve informacion basica del PDF (paginas, tamano).

### pdf_compress
Comprime un PDF optimizando imagenes, eliminando metadatos y limpiando la estructura. Reduce el tamano de 30-70% segun el contenido.

### pdf_remove_pages
Elimina o conserva paginas especificas de un PDF. Soporta rangos de paginas con la sintaxis `2,5-8,11`.

Dos modos de operacion:
- **`remove`** (por defecto): elimina las paginas indicadas y conserva el resto.
- **`keep`**: conserva solo las paginas indicadas y elimina el resto.

Ejemplos de seleccion:
- `"5"` — solo la pagina 5
- `"1-3"` — paginas 1, 2 y 3
- `"2,5-8,11"` — paginas 2, 5, 6, 7, 8 y 11
- `"7-10,61-66,77,80"` — paginas 7-10, 61-66, 77 y 80

**Dependencias clave**
- `github.com/pdfcpu/pdfcpu` — usado para manipulacion de PDFs (split, compresion, informacion, eliminacion de paginas).

## Quick Start (desarrollo)

Build y run (desarrollo):

```powershell
cd c:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\cmd\server
go run .
```

Build de todos los binarios:

```powershell
go build -o bin/server.exe ./cmd/server/
go build -o bin/mcp-server.exe ./cmd/mcp-server/
go build -o bin/cli.exe ./cmd/cli/
```

## Endpoints HTTP

### Split

```powershell
curl -F "file=@test.pdf" http://localhost:8080/api/v1/pdf/split --output split.zip
```

### Compress

```powershell
curl -F "file=@test.pdf" http://localhost:8080/api/v1/pdf/compress --output compressed.pdf
```

Ver informacion de compresion en headers:

```powershell
curl -I -F "file=@test.pdf" http://localhost:8080/api/v1/pdf/compress
# Mostrara X-Original-Size, X-Compressed-Size, X-Reduction-Percent
```

### Remove Pages

Eliminar paginas 2, 5-8 y 11:

```powershell
curl -F "file=@test.pdf" -F "pages=2,5-8,11" -F "mode=remove" http://localhost:8080/api/v1/pdf/remove-pages --output result.pdf
```

Conservar solo paginas 1-3 y 10:

```powershell
curl -F "file=@test.pdf" -F "pages=1-3,10" -F "mode=keep" http://localhost:8080/api/v1/pdf/remove-pages --output result.pdf
```

## CLI

### Split

```powershell
.\bin\cli.exe split -i test.pdf -outdir output
.\bin\cli.exe split -i test.pdf -zip split.zip
```

### Remove Pages

Eliminar paginas 2, 5-8 y 11:

```powershell
.\bin\cli.exe remove-pages -i test.pdf -o result.pdf -pages "2,5-8,11"
```

Conservar solo paginas 1, 3 y 5 (eliminar el resto):

```powershell
.\bin\cli.exe remove-pages -i test.pdf -o result.pdf -pages "1,3,5" -mode keep
```

Ejemplo real — conservar solo paginas especificas de un libro:

```powershell
.\bin\cli.exe remove-pages -i libro.pdf -o seleccion.pdf -pages "7-10,61-66,77,80,119-124" -mode keep
```

## Docker

Construir imagen local:

```powershell
docker build -t mcp-pdf-server:local .
```

Ejecutar contenedor:

```powershell
docker run --rm -p 8080:8080 mcp-pdf-server:local
```

## Tests

Ejecutar tests:

```powershell
go test ./...
```

Nota: `internal/pdf/split_test.go` es una prueba de integracion y se salta si no existe `test.pdf` en la raiz del repo. Puedes usar cualquier PDF para testing local.

## Integracion con Claude Desktop (MCP)

Para integrar con Claude Desktop tienes dos opciones principales:

- **stdio (recomendado para Claude Desktop):** ejecutar un proceso `mcp-server.exe` que hable por `stdin/stdout` con Claude Desktop. Esto es lo que Claude espera para registrar herramientas automaticamente.
- **HTTP (alternativa):** ejecutar el servidor HTTP y usar un adaptador desde Claude (menos directo).

Ejemplo de fragmento para pegar en `claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "pdftools": {
      "command": "C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\bin\\mcp-server.exe",
      "args": [],
      "env": {}
    }
  }
}
```

Herramientas disponibles via MCP: `pdf_split`, `pdf_info`, `pdf_compress`, `pdf_remove_pages`.

### Ejemplos de uso MCP (stdio)

Ejemplo de llamada a `pdf_remove_pages` con PowerShell (modo keep):

```powershell
'{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"pdf_remove_pages","arguments":{"pdf_path":"C:\\\\tmp\\\\libro.pdf","output_path":"C:\\\\tmp\\\\seleccion.pdf","pages":"7-10,61-66,77,80","mode":"keep"}}}' | .\bin\mcp-server.exe
```

Ejemplo de llamada a `pdf_compress`:

```powershell
'{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"pdf_compress","arguments":{"pdf_path":"C:\\\\tmp\\\\test.pdf","output_path":"C:\\\\tmp\\\\test-compressed.pdf"}}}' | .\bin\mcp-server.exe
```

Respuesta MCP (ejemplo):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"original_pages\":357,\"removed_pages\":[1,2,3,...],\"removed_count\":301,\"remaining_pages\":56,\"mode\":\"keep\",\"output_path\":\"C:\\\\tmp\\\\seleccion.pdf\"}"
      }
    ]
  }
}
```

## Roadmap y proximos pasos

Ver `Roadmap.md` en la raiz del repo para ver lo completado y lo pendiente.

## Contacto / ayuda

Para reportar issues o solicitar nuevas funcionalidades, abre un issue en el repositorio.
