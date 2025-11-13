# MCP / Claude Desktop Integration

Este directorio contiene ejemplos y notas para integrar este proyecto con Claude Desktop como un MCP server.

Opciones de integración:

- `stdio` (recomendado para Claude Desktop): el MCP server se ejecuta como proceso local y se comunica por stdin/stdout con Claude Desktop.
- `http` (alternativa): el MCP server expone una API HTTP y se configura en Claude para llamar a los endpoints. Menos integrado; requiere un puente en Claude.

1) Integración rápida — pegar en `claude_desktop_config.json`:

- Abre el archivo de configuración de Claude Desktop (ver `MCP_CLAUDE_DESKTOP.md` en el ejemplo para rutas). Copia el contenido de `claude_desktop_config_sample.json` en la sección `mcpServers`.

2) Uso recomendado (stdio):

- Compila un binario `mcp-server.exe` que implemente el protocolo MCP (stdio). El proyecto de ejemplo `comoejemplo/mcp-go-pdf-to-img-2/mcp` muestra un servidor que expone `pdf_to_images` y `pdf_info` (mira `mcp/server.go`). Puedes usarlo como referencia para implementar `cmd/mcp-server` en este repo.

3) Uso alternativo (HTTP):

- Si prefieres usar el servidor HTTP ya incluido (`cmd/server`), configura Claude Desktop para lanzar el servidor (o ejecutarlo manualmente) y conecta Claude a `http://localhost:8080` mediante un puente. Nota: Claude Desktop espera stdio, por lo que la integración HTTP puede necesitar adaptadores.

4) Próximos pasos que puedo hacer por ti:

- Implementar el `cmd/mcp-server` que ejecuta un loop stdio compatible con Claude y usa las funciones internas (`internal/pdf`, etc.).
- Generar el binario `mcp-server.exe` en `bin/` listo para usar en Windows.
- Crear un descriptor JSON con las herramientas y esquemas ya listos para que Claude las registre automáticamente.

Si quieres que implemente el `stdio` MCP server ahora (recomendado), lo hago: crearé `cmd/mcp-server` con un pequeño protocolo JSON compatible con el ejemplo y lo enlazaré a las funciones ya implementadas.
