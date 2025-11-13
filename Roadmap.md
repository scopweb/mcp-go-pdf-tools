# Roadmap — mcp-go-pdf-tools

Este documento resume el estado actual, lo que queda por hacer y mejoras recomendadas para el proyecto `mcp-go-pdf-tools`.

**Estado actual (resumen)**
- Funcionalidad principal completada: `pdf_split` (divide un PDF en PDFs por página).
- Endpoints: HTTP `/api/v1/pdf/split` y servidor stdio MCP (`cmd/mcp-server`) implementados.
- CLI básico (`cmd/cli`) y ejemplos mínimos incluidos.
- Soporte para generar ZIP de las páginas y devolverlo en la respuesta como `zip_b64` implementado.

**Fase 1 (hecho)**
- Implementar `SplitPDF` usando `pdfcpu` y exponerlo por HTTP y MCP. ✅

**Lo que queda / Prioridad Alta (Fase 2)**
- `pdf_to_images` (convertir páginas PDF a imágenes): investigar e integrar un renderizador (p.ej. PDFium vía cgo o bindings, o un servicio externo). Requiere decidir dependencia y plataforma.
- Tests automáticos: añadir unitarios y tests de integración reproducibles (evitar dependencia de archivos locales cuando sea posible). Incluir CI que ejecute los tests.
- Contenerización y releases: completar `Dockerfile` para producción (usuario no-root, tuneo de build), publicar imágenes en registry o generar artefactos desde CI.
- Permisos y seguridad: validar rutas recibidas por MCP/HTTP y aplicar controles (no permitir escape de directorios, validar tamaños, límites de tiempo).

**Prioridad Media**
- Mejorar la CLI: opciones más ricas, progreso, salida a stdout en lugar de archivos temporales si se solicita.
- Soporte de streaming para grandes outputs (ej. enviar ZIP por stdout en vez de base64 cuando el cliente soporte streaming).
- Manejo de configuraciones: fichero `config.yml` o variables de entorno para paths, tiempo de espera, límites.

**Prioridad Baja / Deseable (Fase 3)**
- Métricas y observabilidad: Exponer métricas (Prometheus), logs estructurados y niveles de log.
- Escalado: diseñar arquitectura para procesar PDFs asíncronamente (queue + workers) para cargas grandes.
- UI / Integraciones: ejemplos para Claude Desktop, documentación de cómo añadir la herramienta en otros MCP clients.

**Mejoras en calidad de código**
- Cobertura de tests para `internal/pdf` y handlers MCP/HTTP.
- Revisión y endurecimiento del manejo de errores (propagar errores, mensajes de error amigables).
- Formateo y linting: añadir linters/formatters a CI (golangci-lint, gofmt).

**Aceptación / Criterios**
- `pdf_split` debe pasar tests de integración: dado un PDF de prueba, generar N PDFs (N = número de páginas) y opcionalmente un ZIP válido.
- `tools/list` debe exponer `inputSchema` correcto para que clientes MCP (como Claude Desktop) puedan construir formularios.
- `pdf_to_images` (cuando se implemente) debe producir imágenes con al menos PNG/JPEG y parámetros de DPI.

**Estimaciones (muy aproximadas)**
- `pdf_to_images` (investigar + MVP): 2–5 días (dependiendo de la elección de renderer y problemas de cgo/plataforma).
- Tests y CI completos: 1–2 días.
- Contenerización y mejora de Dockerfile: 0.5–1 día.

**Siguientes pasos recomendados (inmediatos)**
1. Añadir tests de integración para `pdf_split` (automatizar con un fixture `test.pdf`).
2. Decidir estrategia para `pdf_to_images` (usar PDFium, poppler, un servicio externo o una librería Go pura). Si optas por PDFium, preparar builds cross-platform y/o imágenes precompiladas.
3. Añadir validación de entrada (existencia de `pdf_path`, permisos, evitar escapes con `filepath.Clean` y verificación de rutas relativas).
4. Añadir `zip_b64` en la documentación del MCP (ejemplos de payloads para clientes).

Si quieres, puedo:
- Implementar `pdf_to_images` (requiere elegir dependencia). 
- Añadir tests y completar CI. 
- Hacer la auditoría de seguridad y añadir validaciones de ruta ahora.

---
Archivo generado automáticamente por el asistente — si quieres ajustes de formato, inclusión de fechas estimadas concretas o tickets/isssues vinculados, dime y lo actualizo.
# Roadmap

Este documento resume lo ya realizado y los siguientes pasos para el proyecto `mcp-go-pdf-tools`.

**Hecho (Completado)**

- Inicialización del módulo Go y estructura básica del proyecto (`cmd/server`, `internal/pdf`).
- Implementación de la funcionalidad de separación de PDFs (`internal/pdf/split.go`) usando `pdfcpu` (última versión disponible durante el desarrollo).
- Servidor HTTP mínimo en `cmd/server` con endpoints:
  - `GET /health` — health check
  - `POST /api/v1/pdf/split` — recibe `multipart/form-data` con campo `file` y devuelve un ZIP con PDFs por página.
- Cliente de ejemplo en `examples/upload.go` y test de integración `internal/pdf/split_test.go` que se salta si no existe `test.pdf`.
- Contenerización básica (`Dockerfile`, `.dockerignore`) y workflow CI (`.github/workflows/ci.yml`).
- Documentación inicial y notas para integrar con Claude Desktop (`mcp/claude_desktop_config_sample.json` y `mcp/README.md`).

**En progreso**

- Añadir más tests unitarios y casos de borde para `SplitPDFFile`.
- Refinar el ejemplo/cliente y añadir scripts `make` para tareas comunes.

**Pendiente / Próximos pasos**

1. Implementar `cmd/mcp-server` (proceso `stdio`) que implemente el protocolo MCP (Model Context Protocol) para Claude Desktop y exponga las herramientas (`pdf_to_images`, `pdf_info`, `pdf_split`).
2. Compilar y generar `mcp-server.exe` para Windows y colocar en `bin/` (o publicar artefactos en CI).
3. Añadir opciones de configuración para el endpoint `split` (por ejemplo: devolver archivos sueltos, parámetros `pages`, `zip=false`).
4. Tests automatizados que ejecuten `cmd/server` y validen respuestas HTTP y contenido del ZIP.
5. Publicar release + documentación de integración con Claude Desktop (ejemplos de `claude_desktop_config.json`).

**Notas / Riesgos**

- `pdfcpu` se usa como dependencia principal para manipular PDFs; su API puede cambiar entre versiones mayores.
- Integración stdio MCP requiere implementar un pequeño servidor compatible con el protocolo esperado por Claude Desktop (JSON-RPC-like). El ejemplo en `comoejemplo` puede servir de referencia.

**Recomendación inmediata**

- Implementar `cmd/mcp-server` (stdio) para completar la integración con Claude Desktop. Puedo implementarlo a continuación si confirmas.
