# MCP Server Upgrade to Latest Specification

Esta documento describe la actualización del MCP server a la última especificación con soporte mejorado para Claude Desktop.

## Especificación MCP Soportada

- **2025-11-25** (versión más nueva)
- **2025-06-18**
- **2025-03-26**

La versión más nueva es negociada automáticamente durante la inicialización.

## Cambios Principales

### 1. **Refactorización de Protocolo** (`protocol.go`)

Se creó un modelo completo de protocolo JSON-RPC 2.0 + MCP:

```go
// Tipos clave
type Request struct {
    JSONRPC string
    ID      RequestID       // Puede ser número o string
    Method  string
    Params  json.RawMessage // Raw bytes para parsing flexible
}

type Response struct {
    JSONRPC string
    ID      RequestID
    Result  interface{}
    Error   *RpcError
}

type ToolResult struct {
    Content []ToolContent
    IsError bool
}
```

**Beneficios**:
- Validación de tipos JSON-RPC 2.0 completa
- Soporte para `json.RawMessage` (parsing flexible)
- Códigos de error estándar definidos
- Estructuras genéricas para todos los métodos

### 2. **Registro de Herramientas** (`tools.go`)

Se implementó un patrón de registry con handlers individuales:

```go
type ToolHandler interface {
    Handle(id RequestID, rawArgs json.RawMessage) *Response
    GetDefinition() Tool
}

type ToolsRegistry struct {
    tools map[string]*ToolHandler
    logger logging.Logger
}
```

**Ventajas**:
- Cada herramienta en su propio handler
- Fácil de añadir nuevas herramientas
- Validación de argumentos tipada
- Logging dedicado por herramienta
- Manejo de errores consistente

**Herramientas implementadas**:
1. `pdf_split` - Divide PDF en páginas individuales
2. `pdf_info` - Información del PDF (páginas, tamaño)
3. `pdf_compress` - Comprime PDF optimizando
4. `pdf_remove_pages` - Elimina o conserva páginas específicas

### 3. **Servidor MCP** (`server.go`)

Se creó una clase `MCPServer` que maneja el protocolo:

```go
type MCPServer struct {
    supportedVersions []string
    tools             *ToolsRegistry
    logger            logging.Logger
}

func (s *MCPServer) HandleRequest(req *Request) *Response
```

**Funcionalidades**:
- Manejo de `initialize` con negociación de versión
- Respuesta a `tools/list` dinámicamente
- Ejecución de `tools/call` a través del registry
- Manejo de `ping` para health check
- Manejo de notificaciones (`notifications/initialized`, `notifications/cancelled`)
- Validación de protocolo per spec

### 4. **Main Refactorizado** (`main.go`)

Nuevo main más limpio:
- Carga configuración desde env vars
- Inicializa logger centralizado
- Crea `MCPServer` con inyección de dependencias
- Manejo mejorado de EOF y errores
- Logging en cada paso

## Compatibilidad con Claude Desktop

### Requisitos

1. **Ejecutable en PATH o configurado en config.json**
   ```json
   {
     "mcpServers": {
       "pdf-tools": {
         "command": "/path/to/mcp-server"
       }
     }
   }
   ```

2. **Protocolo JSON-RPC 2.0** ✓
   - Validación de estructuras JSON
   - Códigos de error estándar
   - IDs de solicitud soportados (número o string)

3. **MCP 2025-11-25** ✓
   - `initialize` con `protocolVersion`
   - `tools/list` con InputSchema
   - `tools/call` con Arguments
   - Manejo de notificaciones

4. **Buffer de stdio** ✓
   - 10MB máximo para payloads grandes
   - Manejo de base64 en respuestas (zip_b64)

5. **Error Handling** ✓
   - Códigos de error JSON-RPC estándar
   - `isError: true` en resultados de herramientas
   - Mensajes descriptivos

### Configuración en Claude Desktop

Ejemplo `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "pdf-tools": {
      "command": "/path/to/mcp-go-pdf-tools/bin/mcp-server",
      "env": {
        "LOG_LEVEL": "info",
        "LOG_FORMAT": "text",
        "PDF_VALIDATION_MODE": "relaxed"
      }
    }
  }
}
```

### Variables de Entorno Soportadas

```bash
LOG_LEVEL=info|debug|warn|error (default: info)
LOG_FORMAT=text|json (default: text)
MCP_BUFFER_SIZE=10485760 (10MB, default)
MCP_PROTOCOL_VERSIONS=2025-11-25,2025-06-18,2025-03-26
PDF_IMAGE_QUALITY=75 (default)
PDF_REMOVE_METADATA=true (default)
PDF_VALIDATION_MODE=relaxed (default)
PDF_TEMP_DIR=/tmp (default: sistema)
```

## Ejemplos de Uso

### Desde Claude Desktop

Claude Desktop puede ahora llamar automáticamente:

```
pdf_split: Divide un PDF en páginas individuales
pdf_info: Obtiene información del PDF
pdf_compress: Comprime un PDF
pdf_remove_pages: Elimina o conserva páginas específicas
```

### Manualmente (stdio)

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}' | ./mcp-server
```

Respuesta:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-11-25",
    "capabilities": {
      "tools": {
        "listChanged": false
      }
    },
    "serverInfo": {
      "name": "mcp-go-pdf-tools",
      "version": "1.0.0"
    }
  }
}
```

## Mejoras de Observabilidad

- **Logging estructurado**: Cada solicitud, herramienta y error se registra
- **Debug logs**: Información detallada para troubleshooting
- **Error messages**: Descriptivos y actionables
- **Formato JSON opcional**: Para integración con sistemas de logging

## Arquitectura

```
┌──────────────────────────────┐
│      main.go                 │
│  (config, logger, server)    │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│   MCPServer (server.go)      │
│  (protocol, method dispatch) │
└──────────────┬───────────────┘
               │
       ┌───────┼───────┬──────────┐
       ▼       ▼       ▼          ▼
    init   tools  tools/call   ping
       │      list     │         │
       │       │       │         │
       └───────┼───────┼─────────┘
               ▼
    ┌────────────────────────┐
    │  ToolsRegistry         │
    │  (tools.go)            │
    └────────────────────────┘
               │
       ┌───────┼───────┬──────────┐
       ▼       ▼       ▼          ▼
    pdf_  pdf_  pdf_  pdf_
    split info  compress remove
                        _pages

       ▼ (cada uno)
    Handler.Handle()
       │
       ▼
    Processor (DI)
       │
       ▼
    pdfcpu API
```

## Próximos Pasos

1. **Compilar nuevo binario**:
   ```bash
   go build -o bin/mcp-server ./cmd/mcp-server
   ```

2. **Probar con Claude Desktop**:
   - Añadir a `claude_desktop_config.json`
   - Reiniciar Claude Desktop
   - Usar las herramientas PDF

3. **Monitorear logs**:
   ```bash
   LOG_LEVEL=debug ./bin/mcp-server < test.json
   ```

## Compatibilidad Hacia Atrás

- ✓ Todas las herramientas mantienen su firma original
- ✓ Soporta múltiples versiones de protocolo
- ✓ Manejo de `json.RawMessage` para flexibilidad
- ✓ Códigos de error estándar JSON-RPC

## Referencia

- [MCP Specification](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0](https://www.jsonrpc.org/specification)
- [Claude Desktop Configuration](https://github.com/anthropics/claude-desktop)
