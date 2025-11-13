# mcp-go-pdf-tools

Servidor MCP sencillo para herramientas relacionadas con PDFs, escrito en Go. El objetivo principal de esta primera fase es ofrecer la funcionalidad de "separar un PDF en PDFs independientes (una página por archivo)" y exponerla como API HTTP y como camino para integrarlo con Claude Desktop (MCP).

**Estado:** prototipo funcional — separación por página y endpoint HTTP completados.

**Estructura importante**
- `cmd/server`: servidor HTTP (endpoints `GET /health`, `POST /api/v1/pdf/split`).
- `internal/pdf`: lógica para manipular PDFs (usa `pdfcpu`).
- `examples`: cliente de ejemplo para subir un PDF y guardar `split.zip`.
- `mcp`: notas y ejemplos para integrar con Claude Desktop.

**Dependencias clave**
- `github.com/pdfcpu/pdfcpu` — usado para manipulación de PDFs (split).

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

- Nota: `internal/pdf/split_test.go` es una prueba de integración y se salta si no existe `test.pdf` en la raíz del repo.

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
- Si quieres que implemente el `stdio` MCP server, puedo crear `cmd/mcp-server` y compilar `mcp-server.exe` para Windows.

**Roadmap y próximos pasos**

- Ver `Roadmap.md` en la raíz del repo para ver lo completado y lo pendiente (implementación stdio, tests adicionales, empaquetado de releases).

**Contacto / ayuda**

- Si quieres que siga con la implementación del `cmd/mcp-server` (stdio) o que genere el binario `mcp-server.exe`, dime y lo implemento.


