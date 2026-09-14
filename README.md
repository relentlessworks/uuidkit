# uuidkit

> Agentic-first UUID and ULID generation service. Generate UUID v4, v7, ULIDs, parse and validate identifiers. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
# Build
make build

# Run
./uuidkit

# Generate a UUID v4
curl http://localhost:8471/uuid

# Generate 5 UUID v7s
curl http://localhost:8471/uuid?v=7&n=5

# Generate ULIDs
curl http://localhost:8471/ulid?n=3

# Parse an identifier
curl http://localhost:8471/parse?id=01892b3a-7e3f-7e3f-8e3f-446655440000

# Validate an identifier
curl http://localhost:8471/validate?id=01H8ZG3F4V5J6K7M8N9P0Q1R2S

# Get JSON output
curl -H "Accept: application/json" http://localhost:8471/uuid?v=7
```

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/help` | Operating manual (also at `/.well-known/agent.md`) |
| GET | `/health` | Health check |
| GET | `/uuid` | Generate UUID v4 (default) |
| GET | `/uuid?v=4` | Generate UUID v4 (random) |
| GET | `/uuid?v=7` | Generate UUID v7 (time-ordered) |
| GET | `/uuid?n=5` | Generate 5 UUIDs (max 1000) |
| GET | `/ulid` | Generate a ULID |
| GET | `/ulid?n=5` | Generate 5 ULIDs (max 1000) |
| GET | `/parse?id=<id>` | Parse a UUID or ULID (auto-detected) |
| GET | `/validate?id=<id>` | Validate a UUID or ULID (auto-detected) |
| POST | `/mcp` | MCP (Model Context Protocol) endpoint |

### Response Format

**Plain text (default):** One record per line, space-separated key=value pairs.

```
uuid=550e8400-e29b-41d4-a716-446655440000 version=4
ulid=01H8ZG3F4V5J6K7M8N9P0Q1R2S
```

**JSON:** Send `Accept: application/json` header or `?format=json` query param.

```json
{"uuid":"550e8400-e29b-41d4-a716-446655440000","version":"4"}
```

### Errors

Errors include a hint for self-correction:

```
error: missing id parameter | hint: provide an id to parse, e.g. /parse?id=550e8400-e29b-41d4-a716-446655440000
```

## Configuration

| Source | Variable | Default | Description |
|--------|----------|---------|-------------|
| Env | `UUIDKIT_ADDR` | `:8471` | Listen address |
| Env | `UUIDKIT_API_KEY` | (empty) | API key for auth (no auth if empty) |
| Flag | `-addr` | `:8471` | Override listen address |
| Flag | `-api-key` | (empty) | Override API key |

Priority: defaults < env vars < flags.

## MCP Integration

uuidkit speaks Model Context Protocol at `POST /mcp` for chat client integrations (Claude, Cursor, etc.).

**Tools:**
- `generate_uuid` — Generate UUIDs (params: `version`, `count`)
- `generate_ulid` — Generate ULIDs (params: `count`)
- `parse` — Parse an identifier (params: `id`)
- `validate` — Validate an identifier (params: `id`)

## Build

```bash
make build    # CGO_ENABLED=0, single static binary
make test     # go test -race
make vet      # go vet
make run      # build + run
```

## Design Principles

- **Agent IS the interface** — No UI, no SDK. The API is the product.
- **Plain text by default** — Token-cheap, grepable, one record per line.
- **Instructive errors** — Every 4xx includes a hint for self-correction.
- **Self-documenting** — `GET /help` returns the full operating manual.
- **Single static binary** — Go, zero external dependencies, CGO_ENABLED=0.
- **Zero config** — Runs out of the box with sensible defaults.

## License

MIT
