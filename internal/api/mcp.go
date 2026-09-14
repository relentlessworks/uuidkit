package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/relentlessworks/uuidkit/internal/model"
)

// JSON-RPC 2.0 types

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

// MCP tool definitions

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type mcpToolsResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpToolCallParams struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
	ID      string `json:"id"`
}

func mcpTools() []mcpTool {
	return []mcpTool{
		{
			Name:        "generate_uuid",
			Description: "Generate one or more UUIDs. Version 4 is random, version 7 is time-ordered.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "UUID version: '4' (random) or '7' (time-ordered). Default: '4'",
						"default":     "4",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of UUIDs to generate (1-1000). Default: 1",
						"default":     1,
					},
				},
			},
		},
		{
			Name:        "generate_ulid",
			Description: "Generate one or more ULIDs (26-character Crockford base32, time-ordered).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of ULIDs to generate (1-1000). Default: 1",
						"default":     1,
					},
				},
			},
		},
		{
			Name:        "parse",
			Description: "Parse a UUID or ULID and return its components (version, variant, timestamp).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The UUID or ULID string to parse",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "validate",
			Description: "Validate whether a string is a valid UUID or ULID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The string to validate",
					},
				},
				"required": []string{"id"},
			},
		},
	}
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, "method not allowed", "POST JSON-RPC 2.0 requests to /mcp", http.StatusMethodNotAllowed)
		return
	}

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONRPCError(w, nil, -32700, "parse error: invalid JSON")
		return
	}

	if req.JSONRPC != "2.0" {
		writeJSONRPCError(w, req.ID, -32600, "invalid request: jsonrpc must be '2.0'")
		return
	}

	switch req.Method {
	case "initialize":
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "uuidkit",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		})

	case "tools/list":
		writeJSONRPCResult(w, req.ID, mcpToolsResult{Tools: mcpTools()})

	case "tools/call":
		h.handleMCPToolCall(w, r, req)

	default:
		writeJSONRPCError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (h *Handler) handleMCPToolCall(w http.ResponseWriter, r *http.Request, req jsonRPCRequest) {
	// Extract tool name from params
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeJSONRPCError(w, req.ID, -32602, "invalid params: expected name and arguments")
		return
	}

	var args mcpToolCallParams
	if len(params.Arguments) > 0 {
		_ = json.Unmarshal(params.Arguments, &args)
	}

	switch params.Name {
	case "generate_uuid":
		version := args.Version
		if version == "" {
			version = "4"
		}
		count := args.Count
		if count < 1 {
			count = 1
		}
		if count > 1000 {
			writeJSONRPCError(w, req.ID, -32602, "count must be 1-1000")
			return
		}

		var ids []string
		for i := 0; i < count; i++ {
			var id string
			var err error
			switch version {
			case "4":
				id, err = model.GenerateUUIDv4()
			case "7":
				id, err = model.GenerateUUIDv7()
			default:
				writeJSONRPCError(w, req.ID, -32602, "version must be '4' or '7'")
				return
			}
			if err != nil {
				writeJSONRPCError(w, req.ID, -32603, "failed to generate UUID")
				return
			}
			ids = append(ids, id)
			h.Store.IncrUUID()
		}

		text := ""
		for _, id := range ids {
			text += fmt.Sprintf("uuid=%s version=%s\n", id, version)
		}
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": text},
			},
		})

	case "generate_ulid":
		count := args.Count
		if count < 1 {
			count = 1
		}
		if count > 1000 {
			writeJSONRPCError(w, req.ID, -32602, "count must be 1-1000")
			return
		}

		var ids []string
		for i := 0; i < count; i++ {
			id, err := model.GenerateULID()
			if err != nil {
				writeJSONRPCError(w, req.ID, -32603, "failed to generate ULID")
				return
			}
			ids = append(ids, id)
			h.Store.IncrULID()
		}

		text := ""
		for _, id := range ids {
			text += fmt.Sprintf("ulid=%s\n", id)
		}
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": text},
			},
		})

	case "parse":
		if args.ID == "" {
			writeJSONRPCError(w, req.ID, -32602, "missing required argument: id")
			return
		}
		h.Store.IncrParse()

		if u, err := model.ParseUUID(args.ID); err == nil {
			result := fmt.Sprintf("id=%s type=uuid version=%s variant=%s valid=true", args.ID, u.Version.String(), u.Variant)
			if u.Version == model.Version7 {
				ts := int64(u.Raw[0])<<40 | int64(u.Raw[1])<<32 | int64(u.Raw[2])<<24 |
					int64(u.Raw[3])<<16 | int64(u.Raw[4])<<8 | int64(u.Raw[5])
				result += " timestamp=" + strconv.FormatInt(ts, 10)
			}
			writeJSONRPCResult(w, req.ID, map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": result},
				},
			})
			return
		}
		if u, err := model.ParseULID(args.ID); err == nil {
			result := fmt.Sprintf("id=%s type=ulid timestamp=%d valid=true", args.ID, u.Timestamp)
			writeJSONRPCResult(w, req.ID, map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": result},
				},
			})
			return
		}
		writeJSONRPCError(w, req.ID, -32602, "unrecognized identifier format: provide a valid UUID or ULID")

	case "validate":
		if args.ID == "" {
			writeJSONRPCError(w, req.ID, -32602, "missing required argument: id")
			return
		}
		h.Store.IncrValid()

		idType := model.DetectType(args.ID)
		valid := idType != "unknown"
		result := fmt.Sprintf("id=%s type=%s valid=%t", args.ID, idType, valid)
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": result},
			},
		})

	default:
		writeJSONRPCError(w, req.ID, -32601, fmt.Sprintf("unknown tool: %s", params.Name))
	}
}

func writeJSONRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	})
}
