# MCP / Claude Desktop Integration

# Integración con Claude Desktop (MCP)

Este directorio contiene la configuración para integrar este proyecto con Claude Desktop como un MCP server.

## ✅ Estado actual

El servidor MCP stdio ya está completamente implementado en `cmd/mcp-server` y compilado como `bin/mcp-server.exe`.

**Herramientas disponibles:**
- `pdf_split`: divide un PDF en páginas individuales
- `pdf_info`: obtiene información del PDF (páginas, tamaño)
- `pdf_compress`: comprime un PDF optimizando imágenes y metadatos

## 🚀 Configuración rápida

1. **Copiar la configuración:**
   - Abre tu archivo `claude_desktop_config.json` (ubicación típica: `%APPDATA%\Claude\claude_desktop_config.json`)
   - Copia el contenido de `claude_desktop_config_sample.json` en la sección `mcpServers`

2. **Ejemplo de configuración completa:**
   ```json
   {
     "mcpServers": {
       "mcp-pdf-tools": {
         "command": "C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\bin\\mcp-server.exe",
         "env": {}
       }
     }
   }
   ```

3. **Reiniciar Claude Desktop** para activar el servidor

## 📝 Uso en Claude

Una vez integrado, puedes usar las herramientas directamente en Claude:

### pdf_compress
```
Quiero comprimir un PDF grande. Aquí está ubicado en C:\path\to\large.pdf
Guárdalo comprimido en C:\path\to\large-compressed.pdf
```

### pdf_split
```
Divide este PDF en páginas individuales: C:\path\to\document.pdf
Crea un ZIP con las páginas en C:\path\to\output
```

### pdf_info
```
Dame información sobre este PDF: C:\path\to\file.pdf
```

## 🔧 Alternativa: Usar HTTP directamente

Si prefieres usar el servidor HTTP (`bin/server.exe`) en lugar del MCP:

```powershell
# Terminal 1: Iniciar servidor HTTP
C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools\bin\server.exe

# Terminal 2: Hacer solicitudes
curl -F "file=@C:\path\to\document.pdf" http://localhost:8080/api/v1/pdf/compress -o compressed.pdf
```

## 📚 Desarrollo

- **MCP Server code**: `cmd/mcp-server/main.go`
- **PDF logic**: `internal/pdf/` (split.go, compress.go)
- **HTTP Server code**: `cmd/server/main.go`

Para compilar manualmente:
```bash
go build -o bin/mcp-server.exe ./cmd/mcp-server
go build -o bin/server.exe ./cmd/server
```

## ⚠️ Rutas de archivos

Claude Desktop necesita rutas **absolutas** para acceder a los PDFs. Ejemplos válidos:
- Windows: `C:\Users\tu_usuario\Documents\document.pdf`
- UNC: `\\servidor\compartido\document.pdf`

No funcionan rutas relativas (como `./documento.pdf`).
