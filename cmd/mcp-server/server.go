package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/scopweb/mcp-go-pdf-tools/internal/logging"
	"github.com/scopweb/mcp-go-pdf-tools/internal/pdf"
)

// MCPServer es el servidor MCP/stdio.
type MCPServer struct {
	supportedVersions []string
	tools             *ToolsRegistry
	logger            logging.Logger
}

// NewMCPServer crea un nuevo servidor MCP.
func NewMCPServer(processor *pdf.Processor, logger logging.Logger) *MCPServer {
	return &MCPServer{
		supportedVersions: []string{"2025-11-25", "2025-06-18", "2025-03-26"},
		tools:             NewToolsRegistry(processor, logger),
		logger:            logger,
	}
}

// HandleRequest procesa una solicitud MCP.
func (s *MCPServer) HandleRequest(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "ping":
		return s.handlePing(req)
	case "notifications/initialized":
		s.logger.Debug("received notifications/initialized")
		return nil
	case "notifications/cancelled":
		s.logger.Debug("received notifications/cancelled")
		return nil
	default:
		if req.ID != nil {
			s.logger.Warn("unknown method",
				fmt.Sprintf("method=%s", req.Method))
			return NewErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
		}
		return nil
	}
}

// handleInitialize procesa la solicitud initialize.
func (s *MCPServer) handleInitialize(req *Request) *Response {
	var initReq InitializeRequest
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &initReq); err != nil {
			s.logger.Warn("failed to unmarshal initialize params", err)
		}
	}

	// Negociar versión de protocolo
	negotiated := s.negotiateVersion(initReq.ProtocolVersion)

	s.logger.Info("initialized",
		fmt.Sprintf("protocol_version=%s", negotiated),
		fmt.Sprintf("client_version=%s", initReq.ProtocolVersion))

	response := InitializeResponse{
		ProtocolVersion: negotiated,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "mcp-go-pdf-tools",
			Version: "1.0.0",
		},
	}

	return NewSuccessResponse(req.ID, response)
}

// handleToolsList procesa la solicitud tools/list.
func (s *MCPServer) handleToolsList(req *Request) *Response {
	s.logger.Debug("listing tools")
	toolDefs := s.tools.GetToolDefinitions()
	return NewSuccessResponse(req.ID, ToolsListResponse{Tools: toolDefs})
}

// handleToolsCall procesa la solicitud tools/call.
func (s *MCPServer) handleToolsCall(req *Request) *Response {
	var callReq CallToolRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		s.logger.Warn("failed to unmarshal tools/call params", err)
		return NewToolErrorResult(req.ID, fmt.Sprintf("invalid tool call: %v", err))
	}

	s.logger.Debug("calling tool",
		fmt.Sprintf("tool=%s", callReq.Name))

	return s.tools.CallTool(req.ID, callReq.Name, callReq.Arguments)
}

// handlePing procesa la solicitud ping.
func (s *MCPServer) handlePing(req *Request) *Response {
	s.logger.Debug("received ping")
	return NewSuccessResponse(req.ID, map[string]interface{}{})
}

// negotiateVersion negocia la versión de protocolo.
func (s *MCPServer) negotiateVersion(clientVersion string) string {
	// Si el cliente solicita una versión que soportamos, retornarla (MUST per spec).
	for _, v := range s.supportedVersions {
		if v == clientVersion {
			return v
		}
	}
	// De lo contrario, retornar nuestra versión más nueva (SHOULD per spec).
	return s.supportedVersions[0]
}

// IsNotification verifica si una solicitud es una notificación.
func IsNotification(method string) bool {
	return strings.HasPrefix(method, "notifications/")
}

// RequiresID verifica si un método requiere un ID.
func RequiresID(method string) bool {
	return !IsNotification(method)
}
