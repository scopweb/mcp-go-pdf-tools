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

---

## 🚀 Fase 3: Nuevas herramientas PDF (por implementar)

### ⭐ **Prioridad Alta** - Implementar primero

- [ ] **`pdf_merge`** — Combinar múltiples PDFs en uno
  - Caso de uso: unir reportes, combinar documentos
  - Complejidad: baja (pdfcpu lo soporta nativamente)
  - Impacto: muy alto
  - Parámetros: `pdf_paths[]` (array de rutas), `output_path`

- [ ] **`pdf_extract`** — Extraer rango de páginas a nuevo PDF
  - Caso de uso: extraer páginas específicas sin crear tantos archivos como split
  - Complejidad: baja
  - Impacto: alto (complementa split)
  - Parámetros: `pdf_path`, `start_page`, `end_page`, `output_path`

- [ ] **`pdf_rotate`** — Rotar páginas (90°, 180°, 270°)
  - Caso de uso: escaneos al revés, correcciones rápidas
  - Complejidad: muy baja
  - Impacto: medio-alto
  - Parámetros: `pdf_path`, `rotation` (90|180|270), `pages` (opcional, todas si no se especifica), `output_path`

- [ ] **`pdf_remove_pages`** — Eliminar páginas específicas
  - Caso de uso: limpiar PDFs antes de compartir, eliminar portadas
  - Complejidad: baja
  - Impacto: alto
  - Parámetros: `pdf_path`, `pages[]` (array de números de página), `output_path`

### 📊 **Prioridad Media** - Implementar después

- [ ] **`pdf_watermark`** — Agregar marca de agua (texto/imagen)
  - Caso de uso: proteger documentos, marcar como "CONFIDENCIAL", "DRAFT"
  - Complejidad: media (requiere manejo de imágenes/fuentes)
  - Impacto: medio
  - Parámetros: `pdf_path`, `watermark_text`, `opacity`, `output_path`

- [ ] **`pdf_encrypt`** — Cifrar PDF con contraseña
  - Caso de uso: proteger documentos sensibles
  - Complejidad: media
  - Impacto: medio-alto (seguridad)
  - Parámetros: `pdf_path`, `password`, `output_path`, `owner_password` (opcional)

- [ ] **`pdf_decrypt`** — Desencriptar PDF
  - Caso de uso: remover protección de PDFs
  - Complejidad: media
  - Impacto: medio
  - Parámetros: `pdf_path`, `password`, `output_path`

- [ ] **`pdf_bookmark`** — Agregar índice/marcadores
  - Caso de uso: mejorar navegación en PDFs grandes
  - Complejidad: media
  - Impacto: bajo-medio
  - Parámetros: `pdf_path`, `bookmarks[]` (array de {title, page}), `output_path`

### 🔧 **Prioridad Baja** - Nice to have

- [ ] **`pdf_flatten`** — Aplanar formularios (eliminar interactividad)
  - Complejidad: media
  - Impacto: bajo
  - Parámetros: `pdf_path`, `output_path`

- [ ] **`pdf_reorder`** — Reordenar páginas
  - Caso de uso: reorganizar documentos
  - Complejidad: baja
  - Impacto: bajo
  - Parámetros: `pdf_path`, `order[]` (array de números de página en nuevo orden), `output_path`

- [ ] **`pdf_add_text`** — Agregar texto a páginas
  - Caso de uso: anotar documentos, agregar información
  - Complejidad: media
  - Impacto: bajo-medio
  - Parámetros: `pdf_path`, `text`, `x`, `y`, `font_size`, `output_path`

---

## 📈 Estrategia de implementación

**Recomendación: Implementar en este orden**

1. ✅ `pdf_merge` (1-2 horas) — Mayor ROI, muchos casos de uso
2. ✅ `pdf_extract` (1 hora) — Complementa split, muy útil
3. ✅ `pdf_rotate` (30 min) — Rápido, muy solicitado
4. ✅ `pdf_remove_pages` (1 hora) — Uso común
5. ✅ `pdf_watermark` (2 horas) — Media complejidad
6. ✅ `pdf_encrypt/decrypt` (2 horas) — Seguridad
7. Resto según demanda

**Tiempo estimado:**
- Top 4 herramientas: ~4-5 horas
- Top 6 herramientas: ~9-10 horas
- Todas (10 herramientas): ~15-20 horas

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
