# 🚀 Setup: mcp-go-pdf-tools con Claude Desktop

Esta guía te ayudará a integrar el servidor MCP de herramientas PDF con Claude Desktop.

## ✅ Requisitos previos

- Claude Desktop instalado
- Go 1.24+ (si quieres compilar desde fuente)
- Windows/macOS/Linux

## 📦 Paso 1: Obtener el binario

### Opción A: Usar el binario precompilado (recomendado)
El binario `bin/mcp-server.exe` ya está compilado y listo para usar.

### Opción B: Compilar desde fuente
```bash
cd c:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools
go build -o bin/mcp-server.exe ./cmd/mcp-server
```

## 🔧 Paso 2: Configurar Claude Desktop

1. **Localizar el archivo de configuración:**
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

2. **Copiar la configuración:**

   Si el archivo no existe, créalo con este contenido:
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

   Si el archivo ya existe, añade esta sección a `mcpServers`:
   ```json
   "mcp-pdf-tools": {
     "command": "C:\\MCPs\\clone_PROYECTOS\\mcp-go-pdf-tools\\bin\\mcp-server.exe",
     "env": {}
   }
   ```

3. **Ajustar la ruta si es necesario:**
   - Reemplaza `C:\MCPs\clone_PROYECTOS\mcp-go-pdf-tools` con tu ruta actual
   - Usa `\\` en JSON para barras invertidas Windows

## 🔄 Paso 3: Reiniciar Claude Desktop

Cierra Claude Desktop completamente y vuelve a abrirlo. Las nuevas herramientas deberían aparecer automáticamente.

## ✨ Paso 4: Probar las herramientas

En Claude Desktop, intenta estos comandos:

### Ejemplo 1: Comprimir un PDF
```
Quiero comprimir el PDF ubicado en C:\Users\tu_usuario\Documents\large.pdf
Guárdalo como C:\Users\tu_usuario\Documents\large-compressed.pdf
```

### Ejemplo 2: Dividir un PDF
```
Divide el PDF en C:\Users\tu_usuario\Documents\document.pdf en páginas individuales
Crea un ZIP en C:\Users\tu_usuario\Documents\split_output
```

### Ejemplo 3: Obtener información del PDF
```
Dame información sobre el PDF en C:\Users\tu_usuario\Documents\file.pdf
Quiero saber cuántas páginas tiene y su tamaño
```

## 📋 Herramientas disponibles

### pdf_compress
- **Descripción**: Comprime un PDF optimizando imágenes y metadatos
- **Parámetros**:
  - `pdf_path` (requerido): Ruta absoluta al PDF de entrada
  - `output_path` (requerido): Ruta absoluta donde guardar el PDF comprimido
- **Resultado**: Estadísticas de compresión (tamaño original, comprimido, porcentaje)

### pdf_split
- **Descripción**: Divide un PDF en páginas individuales
- **Parámetros**:
  - `pdf_path` (requerido): Ruta absoluta al PDF
  - `output_dir` (opcional): Directorio para las páginas
  - `zip` (opcional): Crear ZIP con las páginas (true/false)
  - `zip_name` (opcional): Nombre del ZIP
  - `zip_b64` (opcional): Devolver ZIP como base64
- **Resultado**: Lista de archivos PDF creados

### pdf_info
- **Descripción**: Obtiene información del PDF
- **Parámetros**:
  - `pdf_path` (requerido): Ruta absoluta al PDF
- **Resultado**: Número de páginas y tamaño del archivo

## ⚠️ Notas importantes

1. **Rutas absolutas**: Claude Desktop necesita rutas completas (ej: `C:\Users\tu_usuario\Documents\file.pdf`)
   - ❌ No funcionan: `./documento.pdf`, `~/file.pdf`
   - ✅ Sí funcionan: `C:\Users\tu_usuario\Documents\file.pdf`

2. **Permisos**: El proceso MCP necesita acceso de lectura/escritura a los directorios

3. **Velocidad**: La primera herramienta que uses tardará ~1-2 segundos en ejecutar (carga del proceso Go)

## 🐛 Solución de problemas

### Las herramientas no aparecen en Claude Desktop
- ✅ Verifica que la ruta en `claude_desktop_config.json` es correcta
- ✅ Asegúrate de reiniciar Claude Desktop después de cambios
- ✅ Revisa los logs: `Claude > Help > Show Logs`

### Error: "archivo no encontrado"
- ✅ Usa rutas absolutas completas
- ✅ Verifica que el PDF existe en esa ruta
- ✅ Intenta con la ruta en comillas: `"C:\path\to\file.pdf"`

### Error: "permission denied"
- ✅ Verifica que tienes permisos de lectura/escritura
- ✅ Intenta ejecutar Claude Desktop con permisos de administrador

## 📞 Soporte

- GitHub: https://github.com/scopweb/mcp-go-pdf-tools
- Issues: https://github.com/scopweb/mcp-go-pdf-tools/issues

---

**¡Listo!** Ahora puedes usar las herramientas PDF directamente desde Claude Desktop. 🎉
