package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/relentlessworks/uuidkit/internal/model"
	"github.com/relentlessworks/uuidkit/internal/store"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	Store *store.Store
}

// New creates a new API handler.
func New(s *store.Store) *Handler {
	return &Handler{Store: s}
}

// RegisterRoutes wires all endpoints onto the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.root)
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/uuid", h.generateUUID)
	mux.HandleFunc("/ulid", h.generateULID)
	mux.HandleFunc("/parse", h.parse)
	mux.HandleFunc("/validate", h.validate)
	mux.HandleFunc("/mcp", h.mcp)
}

// --- Helpers ---

func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func writeError(w http.ResponseWriter, r *http.Request, msg, hint string, code int) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

func writeText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(text))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// --- Handlers ---

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	writeText(w, "uuidkit — agentic-first UUID/ULID service | hint: GET /help for the operating manual\n")
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeText(w, "ok\n")
}

func (h *Handler) generateUUID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, "method not allowed", "use GET /uuid to generate UUIDs", http.StatusMethodNotAllowed)
		return
	}

	version := r.URL.Query().Get("v")
	if version == "" {
		version = "4"
	}

	n := 1
	if nStr := r.URL.Query().Get("n"); nStr != "" {
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil || n < 1 {
			writeError(w, r, "invalid count parameter", "n must be a positive integer, e.g. /uuid?n=5", http.StatusBadRequest)
			return
		}
		if n > 1000 {
			writeError(w, r, "count too large", "maximum 1000 UUIDs per request, e.g. /uuid?n=1000", http.StatusBadRequest)
			return
		}
	}

	var ids []string
	for i := 0; i < n; i++ {
		var id string
		var err error
		switch version {
		case "4":
			id, err = model.GenerateUUIDv4()
		case "7":
			id, err = model.GenerateUUIDv7()
		default:
			writeError(w, r, "unsupported UUID version", "use v=4 for random UUIDs or v=7 for time-ordered UUIDs, e.g. /uuid?v=7", http.StatusBadRequest)
			return
		}
		if err != nil {
			writeError(w, r, "failed to generate UUID", "try again, the random source may be temporarily unavailable", http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
		h.Store.IncrUUID()
	}

	if wantsJSON(r) {
		result := map[string]interface{}{
			"version": version,
			"count":   len(ids),
			"uuids":   ids,
		}
		writeJSON(w, result)
		return
	}

	var sb strings.Builder
	for _, id := range ids {
		sb.WriteString("uuid=")
		sb.WriteString(id)
		sb.WriteString(" version=")
		sb.WriteString(version)
		sb.WriteString("\n")
	}
	writeText(w, sb.String())
}

func (h *Handler) generateULID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, "method not allowed", "use GET /ulid to generate ULIDs", http.StatusMethodNotAllowed)
		return
	}

	n := 1
	if nStr := r.URL.Query().Get("n"); nStr != "" {
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil || n < 1 {
			writeError(w, r, "invalid count parameter", "n must be a positive integer, e.g. /ulid?n=5", http.StatusBadRequest)
			return
		}
		if n > 1000 {
			writeError(w, r, "count too large", "maximum 1000 ULIDs per request, e.g. /ulid?n=1000", http.StatusBadRequest)
			return
		}
	}

	var ids []string
	for i := 0; i < n; i++ {
		id, err := model.GenerateULID()
		if err != nil {
			writeError(w, r, "failed to generate ULID", "try again, the random source may be temporarily unavailable", http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
		h.Store.IncrULID()
	}

	if wantsJSON(r) {
		result := map[string]interface{}{
			"count": len(ids),
			"ulids": ids,
		}
		writeJSON(w, result)
		return
	}

	var sb strings.Builder
	for _, id := range ids {
		sb.WriteString("ulid=")
		sb.WriteString(id)
		sb.WriteString("\n")
	}
	writeText(w, sb.String())
}

func (h *Handler) parse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, "method not allowed", "use GET /parse?id=<uuid-or-ulid> to parse an identifier", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, r, "missing id parameter", "provide an id to parse, e.g. /parse?id=550e8400-e29b-41d4-a716-446655440000", http.StatusBadRequest)
		return
	}

	h.Store.IncrParse()

	// Try UUID first
	if u, err := model.ParseUUID(id); err == nil {
		ts := ""
		if u.Version == model.Version7 {
			// Extract 48-bit timestamp from first 6 bytes
			t := int64(u.Raw[0])<<40 | int64(u.Raw[1])<<32 | int64(u.Raw[2])<<24 |
				int64(u.Raw[3])<<16 | int64(u.Raw[4])<<8 | int64(u.Raw[5])
			ts = strconv.FormatInt(t, 10)
		}
		if wantsJSON(r) {
			result := map[string]interface{}{
				"id":      id,
				"type":    "uuid",
				"version": u.Version.String(),
				"variant": string(u.Variant),
				"valid":   true,
			}
			if ts != "" {
				result["timestamp"] = ts
			}
			writeJSON(w, result)
			return
		}
		var sb strings.Builder
		sb.WriteString("id=")
		sb.WriteString(id)
		sb.WriteString(" type=uuid version=")
		sb.WriteString(u.Version.String())
		sb.WriteString(" variant=")
		sb.WriteString(string(u.Variant))
		sb.WriteString(" valid=true")
		if ts != "" {
			sb.WriteString(" timestamp=")
			sb.WriteString(ts)
		}
		sb.WriteString("\n")
		writeText(w, sb.String())
		return
	}

	// Try ULID
	if u, err := model.ParseULID(id); err == nil {
		if wantsJSON(r) {
			result := map[string]interface{}{
				"id":        id,
				"type":      "ulid",
				"timestamp": u.Timestamp,
				"valid":     true,
			}
			writeJSON(w, result)
			return
		}
		writeText(w, fmt.Sprintf("id=%s type=ulid timestamp=%d valid=true\n", id, u.Timestamp))
		return
	}

	writeError(w, r, "unrecognized identifier format", "provide a valid UUID (8-4-4-4-12 hex) or ULID (26 Crockford base32 chars), e.g. /parse?id=550e8400-e29b-41d4-a716-446655440000", http.StatusBadRequest)
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, "method not allowed", "use GET /validate?id=<uuid-or-ulid> to validate an identifier", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, r, "missing id parameter", "provide an id to validate, e.g. /validate?id=550e8400-e29b-41d4-a716-446655440000", http.StatusBadRequest)
		return
	}

	h.Store.IncrValid()

	idType := model.DetectType(id)
	valid := idType != "unknown"

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"id":    id,
			"type":  idType,
			"valid": valid,
		})
		return
	}

	writeText(w, fmt.Sprintf("id=%s type=%s valid=%t\n", id, idType, valid))
}
