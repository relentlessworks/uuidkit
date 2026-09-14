package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/relentlessworks/uuidkit/internal/store"
)

func setupTestHandler() *Handler {
	return New(store.New())
}

func TestRoot(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.root(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "uuidkit") {
		t.Errorf("body does not contain 'uuidkit': %s", w.Body.String())
	}
}

func TestHealth(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.health(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "ok\n" {
		t.Errorf("body = %q, want %q", w.Body.String(), "ok\n")
	}
}

func TestHelp(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/help", nil)
	w := httptest.NewRecorder()
	h.help(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "uuidkit") {
		t.Errorf("body does not contain 'uuidkit'")
	}
	if !strings.Contains(body, "GET /uuid") {
		t.Errorf("body does not contain endpoint documentation")
	}
}

func TestGenerateUUIDv4(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "uuid=") {
		t.Errorf("body should start with 'uuid=': %s", body)
	}
	if !strings.Contains(body, "version=4") {
		t.Errorf("body should contain 'version=4': %s", body)
	}
}

func TestGenerateUUIDv7(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?v=7", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "version=7") {
		t.Errorf("body should contain 'version=7': %s", body)
	}
}

func TestGenerateUUIDMultiple(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?n=5", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}

func TestGenerateUUIDJSON(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?format=json", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result["version"] != "4" {
		t.Errorf("version = %v, want '4'", result["version"])
	}
	uuids, ok := result["uuids"].([]interface{})
	if !ok || len(uuids) != 1 {
		t.Errorf("expected 1 uuid in JSON, got %v", result["uuids"])
	}
}

func TestGenerateUUIDInvalidVersion(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?v=99", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "hint:") {
		t.Errorf("error should contain hint: %s", w.Body.String())
	}
}

func TestGenerateUUIDInvalidCount(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?n=abc", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGenerateUUIDCountTooLarge(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/uuid?n=1001", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGenerateUUIDWrongMethod(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("POST", "/uuid", nil)
	w := httptest.NewRecorder()
	h.generateUUID(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestGenerateULID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/ulid", nil)
	w := httptest.NewRecorder()
	h.generateULID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "ulid=") {
		t.Errorf("body should start with 'ulid=': %s", body)
	}
}

func TestGenerateULIDMultiple(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/ulid?n=3", nil)
	w := httptest.NewRecorder()
	h.generateULID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestGenerateULIDJSON(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/ulid?format=json", nil)
	w := httptest.NewRecorder()
	h.generateULID(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	ulids, ok := result["ulids"].([]interface{})
	if !ok || len(ulids) != 1 {
		t.Errorf("expected 1 ulid in JSON, got %v", result["ulids"])
	}
}

func TestParseUUID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/parse?id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	h.parse(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "type=uuid") {
		t.Errorf("body should contain 'type=uuid': %s", body)
	}
	if !strings.Contains(body, "version=4") {
		t.Errorf("body should contain 'version=4': %s", body)
	}
	if !strings.Contains(body, "valid=true") {
		t.Errorf("body should contain 'valid=true': %s", body)
	}
}

func TestParseUUIDv7(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/parse?id=01892b3a-7e3f-7e3f-8e3f-446655440000", nil)
	w := httptest.NewRecorder()
	h.parse(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "version=7") {
		t.Errorf("body should contain 'version=7': %s", body)
	}
	if !strings.Contains(body, "timestamp=") {
		t.Errorf("body should contain 'timestamp=': %s", body)
	}
}

func TestParseULID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/parse?id=01H8ZG3F4V5J6K7M8N9P0Q1R2S", nil)
	w := httptest.NewRecorder()
	h.parse(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "type=ulid") {
		t.Errorf("body should contain 'type=ulid': %s", body)
	}
	if !strings.Contains(body, "timestamp=") {
		t.Errorf("body should contain 'timestamp=': %s", body)
	}
}

func TestParseInvalid(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/parse?id=not-valid", nil)
	w := httptest.NewRecorder()
	h.parse(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "hint:") {
		t.Errorf("error should contain hint: %s", w.Body.String())
	}
}

func TestParseMissingID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/parse", nil)
	w := httptest.NewRecorder()
	h.parse(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestValidateUUID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/validate?id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	h.validate(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "type=uuid") {
		t.Errorf("body should contain 'type=uuid': %s", body)
	}
	if !strings.Contains(body, "valid=true") {
		t.Errorf("body should contain 'valid=true': %s", body)
	}
}

func TestValidateInvalid(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/validate?id=not-valid", nil)
	w := httptest.NewRecorder()
	h.validate(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "type=unknown") {
		t.Errorf("body should contain 'type=unknown': %s", body)
	}
	if !strings.Contains(body, "valid=false") {
		t.Errorf("body should contain 'valid=false': %s", body)
	}
}

func TestValidateMissingID(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/validate", nil)
	w := httptest.NewRecorder()
	h.validate(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestValidateJSON(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/validate?id=550e8400-e29b-41d4-a716-446655440000&format=json", nil)
	w := httptest.NewRecorder()
	h.validate(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result["type"] != "uuid" {
		t.Errorf("type = %v, want 'uuid'", result["type"])
	}
	if result["valid"] != true {
		t.Errorf("valid = %v, want true", result["valid"])
	}
}

func TestMCPInitialize(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %s", result.Error.Message)
	}
}

func TestMCPToolsList(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Tools []mcpTool `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(result.Result.Tools) != 4 {
		t.Errorf("expected 4 tools, got %d", len(result.Result.Tools))
	}
}

func TestMCPGenerateUUID(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"generate_uuid","arguments":{"version":"4","count":2}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %s", result.Error.Message)
	}
}

func TestMCPGenerateULID(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"generate_ulid","arguments":{"count":1}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %s", result.Error.Message)
	}
}

func TestMCPValidate(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"validate","arguments":{"id":"550e8400-e29b-41d4-a716-446655440000"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %s", result.Error.Message)
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	h := setupTestHandler()
	body := `{"jsonrpc":"2.0","id":6,"method":"unknown_method"}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.mcp(w, req)
	var result jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Error == nil {
		t.Errorf("expected error for unknown method")
	}
}

func TestMCPWrongMethod(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()
	h.mcp(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
