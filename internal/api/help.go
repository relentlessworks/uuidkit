package api

import (
	"net/http"
)

// helpText is the one-page operating manual for agents.
const helpText = `uuidkit — Agentic-First UUID/ULID Generation Service
=====================================================

uuidkit generates and validates UUIDs (v4, v7) and ULIDs over plain HTTP.
No database, no state, no external dependencies. Single Go binary.

AUTHENTICATION
---------------
No auth required by default. If UUIDKIT_API_KEY is set, send:
  Authorization: Bearer <api-key>

ENDPOINTS
---------
GET /help              This operating manual (also at /.well-known/agent.md)
GET /health            Health check (returns "ok")
GET /uuid              Generate a UUID v4 (default)
GET /uuid?v=4          Generate a UUID v4 (random)
GET /uuid?v=7          Generate a UUID v7 (time-ordered)
GET /uuid?n=5          Generate 5 UUIDs (max 1000)
GET /uuid?v=7&n=10     Generate 10 UUID v7s
GET /ulid              Generate a ULID
GET /ulid?n=5          Generate 5 ULIDs (max 1000)
GET /parse?id=<id>     Parse a UUID or ULID (auto-detected)
GET /validate?id=<id>  Validate a UUID or ULID (auto-detected)
POST /mcp              MCP (Model Context Protocol) endpoint

RESPONSE FORMAT
---------------
Default: plain text, one record per line, space-separated key=value pairs.
  uuid=550e8400-e29b-41d4-a716-446655440000 version=4
  ulid=01H8ZG3F4V5J6K7M8N9P0Q1R2S

JSON: send Accept: application/json header or ?format=json query param.
  {"uuid":"550e8400-...","version":"4"}
  {"ulid":"01H8ZG3F...","timestamp":1694000000000}

ERRORS
------
Errors are plain text with a hint:
  error: missing id parameter | hint: provide an id to parse, e.g. /parse?id=...

EXAMPLES
--------
  curl http://localhost:8471/uuid
  curl http://localhost:8471/uuid?v=7&n=3
  curl http://localhost:8471/ulid?n=5
  curl http://localhost:8471/parse?id=01892b3a-7e3f-7e3f-8e3f-446655440000
  curl http://localhost:8471/validate?id=01H8ZG3F4V5J6K7M8N9P0Q1R2S
  curl -H "Accept: application/json" http://localhost:8471/uuid?v=7

CONFIGURATION
-------------
  UUIDKIT_ADDR     Listen address (default :8471)
  UUIDKIT_API_KEY  API key for auth (default: none, no auth)
  -addr            Override listen address
  -api-key         Override API key

VERSION
-------
0.1.0
`

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(helpText))
}
