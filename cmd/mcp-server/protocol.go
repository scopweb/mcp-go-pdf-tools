package main

import (
	"encoding/json"
	"fmt"
)

// MCPProtocol contiene tipos y funciones de protocolo MCP/stdio.

// RequestID puede ser un número o string según la especificación JSON-RPC 2.0.
type RequestID interface{}

// Request es una solicitud JSON-RPC 2.0 (MCP).
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response es una respuesta JSON-RPC 2.0 (MCP).
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      RequestID   `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RpcError   `json:"error,omitempty"`
}

// RpcError es una estructura de error JSON-RPC 2.0.
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// StandardErrorCodes son códigos de error estándar JSON-RPC 2.0.
const (
	ParseError      = -32700
	InvalidRequest  = -32600
	MethodNotFound  = -32601
	InvalidParams   = -32602
	InternalError   = -32603
	ServerErrorBase = -32000
)

// ServerInfo contiene información sobre el servidor MCP.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeRequest es la solicitud de inicialización.
type InitializeRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities,omitempty"`
	ClientInfo      ServerInfo `json:"clientInfo,omitempty"`
}

// InitializeResponse es la respuesta de inicialización.
type InitializeResponse struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo `json:"serverInfo"`
}

// Tool es una herramienta disponible a través de MCP.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"` // Puede ser un JSON schema
}

// ToolsListResponse es la respuesta de tools/list.
type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

// CallToolRequest es la solicitud para llamar una herramienta.
type CallToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolContent es el contenido de una herramienta en la respuesta.
type ToolContent struct {
	Type string `json:"type"` // "text", "image", "resource"
	Text string `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// ToolResult es el resultado de llamar una herramienta.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// NewSuccessResponse crea una respuesta exitosa.
func NewSuccessResponse(id RequestID, result interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse crea una respuesta de error JSON-RPC.
func NewErrorResponse(id RequestID, code int, message string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RpcError{
			Code:    code,
			Message: message,
		},
	}
}

// NewToolResult crea un resultado de herramienta exitoso.
func NewToolResult(id RequestID, text string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: text},
			},
			IsError: false,
		},
	}
}

// NewToolErrorResult crea un resultado de herramienta con error.
func NewToolErrorResult(id RequestID, text string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{Type: "text", Text: text},
			},
			IsError: true,
		},
	}
}

// UnmarshalParams desmarshalea los parámetros JSON en la estructura especificada.
func UnmarshalParams(params json.RawMessage, v interface{}) error {
	if len(params) == 0 {
		return nil
	}
	return json.Unmarshal(params, v)
}
