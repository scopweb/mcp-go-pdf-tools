Descripción breve
-----------------
Este repositorio contiene una pequeña suite en Go para operaciones con PDF, pensada para usarse como:

- Servicio HTTP (`cmd/server`) con endpoint `/api/v1/pdf/split`.
- Servidor stdio MCP (Model Context Protocol) para integrarse con clientes como Claude Desktop (`cmd/mcp-server`).
- Herramientas CLI básicas (`cmd/cli`).

Funcionalidad implementada
--------------------------
- `pdf_split`: divide un PDF en archivos de una página cada uno. Puede crear un ZIP con las páginas y devolverlo en la respuesta como `zip_b64`.
- `pdf_info`: devuelve información básica del PDF (páginas, tamaño).

Nota: `pdf_to_images` todavía no está implementado en este repo.

Estructura clave
-----------------
- `cmd/mcp-server`: servidor stdio (MCP) — devuelve `tools/list` y acepta `tools/call`.
- `cmd/server`: servidor HTTP con `POST /api/v1/pdf/split`.
- `internal/pdf`: lógica de manipulación de PDF (usa `pdfcpu`).
- `bin\mcp-server.exe`: binario compilado (Windows) para uso local con Claude Desktop.

Cómo interactuar (MCP / stdio)
-----------------------------
1) Obtener la lista de herramientas (`tools/list`): el servidor responde con `tools` que incluyen `pdf_split` y `pdf_info` y su `inputSchema`.

Ejemplo de `tools/list` (request):
`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`

Ejemplo de llamada a `pdf_split` (`tools/call`) — PowerShell (nota los `\\` para backslashes):
`'{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pdf_split","arguments":{"zip":true,"zip_b64":true,"pdf_path":"C:\\\\MCPs\\\\...\\\\examples\\\\test.pdf","zip_name":"test_pages.zip","output_dir":"C:\\\\MCPs\\\\...\\\\examples\\\\split_zip_test"}}}' | .\bin\mcp-server.exe`

Respuesta (resumen):
- `result.files`: lista de archivos creados.
- `result.zip`: ruta al ZIP en disco (si `zip=true`).
- `result.zip_b64`: contenido del ZIP codificado en base64 (si `zip_b64=true`).

Importante: cuando construyas JSON en clientes, escapa las barras invertidas (`\`) o usa rutas con `/` para evitar secuencias de escape (p.ej. `\t` → tab).

Cómo interactuar (HTTP)
-----------------------
POST `/api/v1/pdf/split` — multipart/form-data campo `file` con el PDF. Devuelve `application/zip` con las páginas.

Ejemplo curl (PowerShell):
```powershell
curl -X POST "http://localhost:8080/api/v1/pdf/split" -F "file=@C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\examples\\test.pdf" --output split.zip
```

Notas operativas / debugging
---------------------------
- Si el cliente (Claude) no puede acceder al sistema de ficheros donde corre el servidor, pide `zip_b64:true` para recibir el ZIP inline.
- El servidor MCP ahora reporta errores claros si falla la creación del ZIP.

Siguientes pasos recomendados
-----------------------------
- Añadir tests automáticos e integración.
- Implementar `pdf_to_images` si necesitas imágenes de páginas (recomendado investigar PDFium o poppler).

Contacto/uso
------------
Si necesitas que prepare ejemplos concretos para Claude (payloads, flujos de trabajo), dímelo y los añado a `mcp/README.md`.
