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
- `pdf_compress`: comprime un PDF optimizando imágenes, eliminando metadata y limpiando estructura.
- `pdf_remove_pages`: elimina o conserva páginas específicas de un PDF. Soporta rangos como `2,5-8,11`. Dos modos: `remove` (borra las páginas indicadas) y `keep` (conserva solo las indicadas y borra el resto).

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

Ejemplo de llamada a `pdf_remove_pages` — modo remove (borra páginas 2, 5-8 y 11):
`'{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"pdf_remove_pages","arguments":{"pdf_path":"C:\\\\MCPs\\\\...\\\\test.pdf","output_path":"C:\\\\MCPs\\\\...\\\\result.pdf","pages":"2,5-8,11","mode":"remove"}}}' | .\bin\mcp-server.exe`

Ejemplo de llamada a `pdf_remove_pages` — modo keep (conserva solo páginas 1, 3 y 10):
`'{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"pdf_remove_pages","arguments":{"pdf_path":"C:\\\\MCPs\\\\...\\\\test.pdf","output_path":"C:\\\\MCPs\\\\...\\\\result.pdf","pages":"1,3,10","mode":"keep"}}}' | .\bin\mcp-server.exe`

Respuesta `pdf_remove_pages` (resumen):
- `result.original_pages`: total de páginas del PDF original.
- `result.removed_pages`: lista de páginas eliminadas.
- `result.removed_count`: cantidad de páginas eliminadas.
- `result.remaining_pages`: páginas que quedan en el resultado.
- `result.mode`: modo usado (`remove` o `keep`).
- `result.output_path`: ruta del PDF resultante.

Importante: cuando construyas JSON en clientes, escapa las barras invertidas (`\`) o usa rutas con `/` para evitar secuencias de escape (p.ej. `\t` → tab).

Cómo interactuar (HTTP)
-----------------------
POST `/api/v1/pdf/split` — multipart/form-data campo `file` con el PDF. Devuelve `application/zip` con las páginas.

Ejemplo curl (PowerShell):
```powershell
curl -X POST "http://localhost:8080/api/v1/pdf/split" -F "file=@C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\examples\\test.pdf" --output split.zip
```

`POST /api/v1/pdf/remove-pages` — multipart/form-data con campo `file` (PDF), campo `pages` (selección, ej: `"2,5-8,11"`) y campo opcional `mode` (`"remove"` o `"keep"`). Devuelve el PDF resultante.

```powershell
curl -X POST "http://localhost:8080/api/v1/pdf/remove-pages" -F "file=@test.pdf" -F "pages=2,5-8,11" -F "mode=remove" --output result.pdf
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
