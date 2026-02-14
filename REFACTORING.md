# Go Best Practices Refactoring

Este documento detalla las mejoras y refactorizaciones realizadas en el codebase siguiendo las buenas prácticas de Go.

## Resumen de Cambios

### 1. **Creación de Paquetes de Infraestructura**

#### `internal/errors/`
- **Archivo**: `errors.go`
- **Cambios**: Creación de tipos de error personalizados para mejor manejo de contexto
- **Ventajas**:
  - Errores descriptivos con información de operación y ruta
  - Soporte para `errors.Is()` y `errors.As()`
  - Diferenciación entre errores de validación y otros tipos

#### `internal/logging/`
- **Archivo**: `logger.go`
- **Cambios**: Implementación de abstracción de logging basada en `slog` (stdlib desde Go 1.21)
- **Ventajas**:
  - Interfaz centralizada para logging (no acoplamiento a `log` estándar)
  - Soporte para niveles de log (DEBUG, INFO, WARN, ERROR)
  - Múltiples formatos de salida (texto, JSON)
  - Contexto a través de atributos y goroutines
  - Sin dependencias externas (usa slog stdlib)

#### `internal/config/`
- **Archivo**: `config.go`
- **Cambios**: Gestión centralizada de configuración desde variables de entorno
- **Ventajas**:
  - Configuración desde env vars con defaults sensatos
  - Separación de responsabilidades (ServerConfig, MCPServerConfig, CLIConfig)
  - Timeouts, límites de upload y opciones de PDF configurables
  - Fácil de extender para YAML/JSON en el futuro

#### `internal/types/`
- **Archivo**: `types.go`
- **Cambios**: Tipos seguros en tiempo de compilación para operaciones PDF
- **Ventajas**:
  - `PageRemovalMode` enum para evitar strings error-prone
  - Structs de resultado (SplitResult, PDFInfoResult, etc.) reemplazando maps genéricos
  - Type safety en todo el código
  - API contracts documentadas y verificables

#### `internal/util/`
- **Archivo**: `cleanup.go`
- **Cambios**: Utilities para gestión segura de recursos
- **Ventajas**:
  - `ResourceCleaner` para rastrear y limpiar archivos/directorios
  - Cleanup sincrónico y asincrónico
  - Manejo de contexto para cancelación
  - Mejor que goroutinas con `time.Sleep` hardcodeado

### 2. **Refactorización de `internal/pdf/`**

#### Nuevo Procesador con Inyección de Dependencias
- **Archivo**: `processor.go` (nuevo)
- **Cambios principales**:
  - Clase `Processor` que encapsula la lógica PDF con DI
  - Inyección de `config.PDFConfig` y `logging.Logger`
  - Métodos: `Split()`, `GetInfo()`, `Compress()`, `RemovePages()`
  - Validación de archivos centralizada en `ValidateFile()`
  - Mejor manejo de errores con contexto

#### Archivos existentes refactorizados
- **split.go**: Funciones wrapper que usan `Processor` para compatibilidad hacia atrás
- **pages.go**: Migración a usar `Processor`, actualización de tipos a `types.PageRemovalMode`
- **compress.go**: Wrapper functions manteniendo API pública

**Ventajas**:
- Testeable: se puede mockear el `Processor`
- Configurable: valida y aplica configuración
- Con logging: todas las operaciones se registran
- Mantiene compatibilidad hacia atrás con funciones públicas

### 3. **Refactorización de `cmd/server/`**

#### Nuevo Handlers Package
- **Archivo**: `handlers.go` (nuevo)
- **Cambios principales**:
  - Clase `Handlers` que centraliza toda la lógica HTTP
  - Métodos: `Health()`, `Split()`, `RemovePages()`, `Compress()`
  - Uso de `Processor` para lógica PDF
  - Validación de nombres de archivo (evita panics)
  - Logging estructurado en cada operación
  - Metadata en headers HTTP (X-* headers)
  - Cleaning resource con `ResourceCleaner`

#### Nuevo Main
- **Archivo**: `main.go` (refactorizado)
- **Cambios principales**:
  - Carga de configuración desde env vars
  - Inicialización de logger configurado
  - Creación de `Processor` compartido
  - Inicialización de `Handlers` con dependencias
  - URL con configuración (host, puerto desde config)
  - Mejor manejo de shutdown

**Ventajas**:
- Separación clara de responsabilidades
- Inyección de dependencias
- Configuración externalizada
- Logging consistente
- Mejor manejo de recursos
- Código más testeable

### 4. **Mejoras de Seguridad**

- **Path Traversal Prevention**: `ensureOutputDir()` valida rutas absolutas
- **Filename Sanitization**: `sanitizeFilename()` limpia extensiones de forma segura
- **Panic Prevention**: Validación de bounds en manipulación de strings
- **Temp Directory Permissions**: `os.MkdirTemp()` con permisos 0700
- **Resource Cleanup**: Gestión adecuada de goroutines con contexto

### 5. **Mejoras de Observabilidad**

- Logging centralizado con niveles
- Atributos contextuales en logs
- Formato de log configurable (texto/JSON)
- Metadata en respuestas HTTP (headers X-*)
- Rastreo de operaciones fallidas

## Compatibilidad Hacia Atrás

Todas las funciones públicas en `internal/pdf/` mantienen su firma original:
- `SplitPDFFile(inputPath string) ([]string, error)`
- `GetPDFInfo(inputPath string) (map[string]interface{}, error)`
- `RemovePagesFromFile(inputPath, outputPath, pageSelection string, keepMode bool) (map[string]interface{}, error)`
- `CompressPDFFile(inputPath, outputPath string, opts CompressOptions) (map[string]interface{}, error)`
- `CompressPDFWithDefaults(inputPath, outputPath string) (map[string]interface{}, error)`

Están marcadas como `Deprecated` pero funcionan mediante wrappers que usan `Processor`.

## Variables de Entorno Soportadas

### HTTP Server (`cmd/server`)
```
HTTP_HOST=0.0.0.0 (default)
HTTP_PORT=8080 (default)
HTTP_READ_TIMEOUT=30s (default)
HTTP_WRITE_TIMEOUT=120s (default)
HTTP_IDLE_TIMEOUT=120s (default)
HTTP_MAX_UPLOAD_SIZE=209715200 (200MB, default)
LOG_LEVEL=info|debug|warn|error (default: info)
LOG_FORMAT=text|json (default: text)
PDF_IMAGE_QUALITY=75 (default)
PDF_REMOVE_METADATA=true (default)
PDF_VALIDATION_MODE=relaxed (default)
PDF_TEMP_DIR=/tmp (default: sistema)
```

### MCP Server (`cmd/mcp-server`)
```
LOG_LEVEL=info|debug|warn|error (default: info)
LOG_FORMAT=text|json (default: text)
MCP_BUFFER_SIZE=10485760 (10MB, default)
MCP_PROTOCOL_VERSIONS=2025-11-25,2025-06-18,2025-03-26 (default)
PDF_* variables igual que HTTP server
```

## Patrones Aplicados

### 1. **Dependency Injection**
```go
processor := pdf.NewProcessor(cfg.PDF, logger)
handlers := NewHandlers(processor, logger, cfg)
```

### 2. **Error Handling con Contexto**
```go
if err := p.ValidateFile(inputPath); err != nil {
    return nil, err
}
```

### 3. **Logging Estructurado**
```go
logger.Info("starting HTTP server",
    fmt.Sprintf("addr=%s", addr),
    fmt.Sprintf("log_level=%s", cfg.LogLevel))

logger.Error("server error", err)
```

### 4. **Resource Cleanup**
```go
cleaner := util.NewResourceCleaner(h.logger)
cleaner.AddDirectory(partsDir)
cleaner.CleanupASAPAfterResponse(500 * time.Millisecond)
```

### 5. **Type Safety**
```go
mode, ok := types.ParseMode(modeStr)
if !ok {
    return errors.New("invalid mode")
}
result.Mode = mode
```

## Próximos Pasos Recomendados

1. **Refactorizar `cmd/mcp-server`**: Aplicar mismo patrón de handlers/processor
2. **Añadir tests unitarios**: Para handlers HTTP usando `httptest`
3. **Añadir tests integrales**: Que prueben end-to-end con PDFs reales
4. **Metrics/Observability**: Implementar Prometheus o similar
5. **Graceful Shutdown**: Con context.Context y WaitGroup
6. **Circuit Breaker**: Para manejar fallos de pdfcpu
7. **Caching**: Para evitar re-procesar PDFs iguales

## Referencias de Go Best Practices Aplicadas

- [Effective Go - Error handling](https://golang.org/doc/effective_go#errors)
- [pkg.go.dev/log/slog](https://pkg.go.dev/log/slog)
- [Dependency Injection in Go](https://www.ardanlabs.com/blog/2016/11/using-interfaces-in-go-pt1.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Practical Go](https://dave.cheney.net/practical-go/presentations/dotgo-paris.html)
