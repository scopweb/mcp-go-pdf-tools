# Roadmap — mcp-go-pdf-tools

Este documento resume el estado actual, lo que queda por hacer y mejoras recomendadas para el proyecto `mcp-go-pdf-tools`.

**Estado actual (resumen - 2025-12-13)**
- ✅ Funcionalidad principal completada: `pdf_split`, `pdf_info`, `pdf_compress`
- ✅ Endpoints HTTP: `/api/v1/pdf/split` y `/api/v1/pdf/compress`
- ✅ Servidor stdio MCP (`cmd/mcp-server`) con 3 herramientas implementadas
- ✅ Soporte para ZIP de páginas con `zip_b64`
- ✅ Compilados binarios Windows (`mcp-server.exe`, `server.exe`)
- ✅ Integración con Claude Desktop completada
- ✅ Documentación: README, SETUP_CLAUDE_DESKTOP.md, mcp/README.md

**Fase 1 ✅ (completado)**
- Implementar `SplitPDF` usando `pdfcpu` y exponerlo por HTTP y MCP. ✅
- Implementar `pdf_info` para obtener metadatos del PDF. ✅
- Implementar `pdf_compress` para optimización de PDFs. ✅

**Fase 2: Lo que queda / Prioridad Alta**
- `pdf_to_images` (convertir páginas PDF a imágenes): **YA IMPLEMENTADO EN REPOSITORIO SEPARADO**
  - ✅ Proyecto: https://github.com/scopweb/mcp-go-pdf-to-img
  - ✅ Usa PDFium (WebAssembly) para conversión
  - ✅ Incluye manejo de errores para PDFs grandes
  - ⚠️ No incluida en este repo porque se enfoca en herramientas de manipulación, no conversión

- Tests automáticos: añadir unitarios para `pdf_compress` y `pdf_info`
- Contenerización: completar `Dockerfile` para producción
- Permisos y seguridad: validar rutas y aplicar controles de escape

**Prioridad Media**
- Mejorar la CLI: opciones más ricas, progreso
- Soporte de streaming para grandes outputs
- Manejo de configuraciones: fichero `config.yml` o variables de entorno

**Prioridad Baja / Deseable**
- Métricas y observabilidad: Prometheus, logs estructurados
- Escalado: procesamiento asincrónico (queue + workers)
- Integración con más MCP clients (no solo Claude Desktop)

**Mejoras en calidad de código**
- ✅ Tests de seguridad (path traversal, etc.)
- Cobertura de tests para `internal/pdf`
- Revisión y endurecimiento del manejo de errores
- Linting: golangci-lint, gofmt

**Aceptación / Criterios**
- ✅ `pdf_split` pasa tests de integración
- ✅ `pdf_compress` reduce tamaño 30-70%
- ✅ `pdf_info` retorna metadatos correctos
- ✅ MCP schema correcto para Claude Desktop
- ✅ Documentación de integración completa

**Siguientes pasos recomendados (inmediatos)**
1. Probar con Claude Desktop (guía en SETUP_CLAUDE_DESKTOP.md)
2. Agregar tests unitarios para `pdf_compress` y `pdf_info`
3. Optimizar el binario MCP (actualmente 20MB)
4. Considerar agregar más herramientas (merge, rotate, watermark, etc.)

---

**Notas de arquitectura**

Este proyecto se divide en dos repositorios:

1. **mcp-go-pdf-tools** (este) — Herramientas de MANIPULACIÓN
   - `pdf_split` — divide en páginas
   - `pdf_info` — obtiene metadatos
   - `pdf_compress` — optimiza tamaño
   - Usa `pdfcpu` para operaciones genéricas

2. **mcp-go-pdf-to-img** — Herramientas de CONVERSIÓN
   - `pdf_to_images` — convierte a imágenes PNG/JPG
   - Usa PDFium (WebAssembly) para renderizado
   - Manejo especializado de PDFs grandes

Esta separación permite:
- ✅ Responsabilidades claras
- ✅ Dependencias ligeras en cada proyecto
- ✅ Facilita mantenimiento independiente

**Última actualización**: 2025-12-13
**Versión del proyecto**: prototipo v0.2.0 (con compression feature)
