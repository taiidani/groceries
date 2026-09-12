# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Overview

Groceries is a personal grocery tracking application written in Go. It uses PostgreSQL for persistent storage, Redis for caching and real-time events via SSE (Server-Sent Events), and serves an HTMX-based web interface alongside a JSON API for the iOS client.

## Architecture

Architecture documentation lives in `docs/` — start at `docs/architecture.md` and treat it as the source of truth for components, request flow, and conventions. Do not duplicate architecture details here.

Quick orientation:

- `internal/api/` — JSON API transport (`/api/v1/`, Bearer token auth) for the iOS client
- `internal/server/` — HTMX web transport (handlers, routing, templates)
- `internal/service/` — all domain logic shared by both transports; handlers contain no business logic
- `internal/db/` — PostgreSQL via sqlc (queries in `queries/`, generated code in `models/` — do not edit)
- `internal/authz/` — credential resolution (API tokens and web sessions, both in Redis)
- `internal/cache/`, `internal/events/` — Redis cache interface and SSE pub/sub

## Development Commands

This project uses `mise` for task automation. All commands should be run via `mise`:

### Build and Run
```bash
mise run build              # Build the server binary (output: ./groceries)
mise run default            # Run full pipeline: dependencies, test, lint, build, then execute
./groceries                 # Run the compiled binary directly
```

### Testing
```bash
mise test                   # Run all unit tests with race detector and coverage
go test ./...               # Run tests without race detection
go test ./internal/service  # Run tests for a specific package
```

### Linting
```bash
mise lint                   # Run go vet and staticcheck
```

### Dependencies
```bash
mise dependencies           # Download Go modules and HTMX assets
mise run generate:sqlc      # Regenerate internal/db/models after editing queries
```

### Database Operations
```bash
goose up        # Apply pending migrations
goose down      # Rollback one migration
goose status    # Show migration status
mise run seed           # Populate database with seed data
```

## Development Setup

1. Start dependencies: `docker compose up -d` (launches Redis and PostgreSQL)
2. Ensure `.env` file exists with required variables (see `mise.toml` for template)
3. Run migrations: `goose up` (or automatic on first run)
4. Build and run: `mise run default`
5. Access at `http://localhost:3000`

## Environment Variables

Defaults defined in 'env' block and loaded automatically by mise.

```
PORT=3000                    # HTTP server port
DATABASE_URL=postgresql://...  # PostgreSQL connection string
REDIS_HOST=localhost:6379    # Redis host:port
DB_TYPE=postgres             # Database type (only postgres supported)
DEV=true                     # Enable dev mode (live template reload; also gates the /auth/dev-login bypass)
LOG_LEVEL=info               # Logging level (debug, info, warn, error)
OIDC_ISSUER_URL=https://auth.taiidani.com  # Authelia issuer URL, used for OIDC discovery at startup
OIDC_CLIENT_ID=groceries     # Authelia client_id registered for this app

# OpenTelemetry tracing (all optional; standard OTEL variables):
OTEL_SERVICE_NAME=groceries  # Service name reported to the trace backend
OTEL_TRACES_EXPORTER=otlp    # Exporter selection: otlp (default), console, or none
OTEL_EXPORTER_OTLP_ENDPOINT=...  # OTLP collector/agent endpoint
OTEL_EXPORTER_OTLP_PROTOCOL=grpc # grpc or http/protobuf
OTEL_TRACES_SAMPLER=parentbased_always_on  # Sampling strategy (default)
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=dev  # Extra resource attributes
```

User defined environment variables may be set in `.env` file (loaded by mise) and are gitignored.

## Secrets

Required in `.env` file (loaded by mise):

```
OIDC_CLIENT_SECRET=...  # Authelia client_secret for the `groceries` client (see 1Password)
```

Only needed to exercise the real Authelia login flow (`/auth/login` → `/auth/callback`).
For local/CI testing without it, use the DevMode-only `/auth/dev-login?username=<name>`
bypass instead (requires `DEV=true`, never registered in production) — in that case
`OIDC_CLIENT_SECRET` just needs to be any non-empty placeholder value, since the server
still requires it to be present at startup even though it's never used.

## Testing Guidelines

- Tests use `_test.go` suffix and live alongside implementation files
- Unit tests should not require external dependencies (use in-memory cache mock, sqlmock for DB)
- Use `t.Parallel()` for tests that can run concurrently
- Table-driven tests preferred for multiple test cases

## Documentation

When writing or updating documentation:

- **Timeless framing**: describe the system as it is, not as a changelog or migration narrative. Point-in-time specs/plans do not belong in `docs/`.
- **Prefer more, smaller articles** focused on one concern over single long documents.
- **Use Mermaid** for all diagrams (```mermaid code blocks). File trees and code samples stay as plain code blocks.
- Keep docs in sync with code changes that alter architecture, request flow, or conventions.
