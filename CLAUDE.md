# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Groceries is a personal grocery tracking application written in Go. It uses PostgreSQL for persistent storage, Redis for caching and real-time events via SSE (Server-Sent Events), and serves an HTMX-based web interface.

## Development Commands

This project uses `mise` for task automation. All commands should be run via `mise`:

### Build and Run
```bash
mise run build              # Build the server binary (output: ./groceries)
mise run default        # Run full pipeline: dependencies, test, lint, build, then execute
./groceries             # Run the compiled binary directly
```

### Testing
```bash
mise test               # Run all unit tests with race detector and coverage
go test ./...           # Run tests without race detection
go test ./internal/models  # Run tests for a specific package
```

### Linting
```bash
mise lint               # Run go vet and staticcheck
```

### Dependencies
```bash
mise dependencies       # Download Go modules and HTMX assets
```

### Database Operations
```bash
go tool goose up        # Apply pending migrations
go tool goose down      # Rollback one migration
go tool goose status    # Show migration status
mise run seed               # Populate database with seed data
```

## Architecture

### Core Components

**HTTP Server** (`internal/server/`)
- Server-side rendered HTML using Go templates with HTMX for interactivity
- Template files in `internal/server/templates/` (embedded in binary via `//go:embed`)
- Static assets in `internal/server/assets/` (HTMX libraries, images)
- Routing uses standard library `http.ServeMux` with method-specific patterns (e.g., `GET /items`, `POST /item/add`)
- All routes wrapped with Sentry middleware for error tracking
- Session management via `sessionMiddleware`, admin access via `adminMiddleware`

**Data Models** (`internal/models/`)
- Domain models: User, Group, Item, Category, Store, List
- Each model has CRUD operations that work directly with PostgreSQL via `database/sql`
- Shared database connection initialized in `models.InitDB()`
- Use `insertWithID()` helper for inserts that need the new record's ID

**Database** (`internal/db/`)
- Uses PostgreSQL (pgx driver) with embedded Goose migrations
- Migrations in `internal/db/migrations/*.sql` with `+goose Up/Down` directives
- Schema auto-applies on startup via `db.New()` → `ensureSchema()`
- Connection string via `DATABASE_URL` environment variable

**Cache Layer** (`internal/cache/`)
- Unified `Cache` interface with Redis implementation (`RedisStore`)
- Also provides in-memory cache implementation for testing
- Used for session storage and temporary data
- Redis client initialized via `cache.NewClient()` with TLS support

**Real-time Events** (`internal/events/`)
- SSE (Server-Sent Events) for push notifications to browser clients
- Redis-backed pub/sub via `EventServer`
- Subscribe to channels, publish events as JSON
- SSE endpoint at `GET /sse` handled by `sseHandler`

**Authorization** (`internal/authz/`)
- Cookie-based session management
- Sessions stored in cache (Redis) with configurable expiration
- Session helpers for creation, retrieval, and deletion

### Request Flow

1. Request arrives → Sentry middleware → Session middleware
2. Session loaded from cookie → User retrieved from database
3. Context populated with session/user data via `context.Value`
4. Handler executes business logic
5. Template rendered with data bag (`baseBag` struct)
6. SSE updates pushed to clients for real-time sync

### Key Patterns

- **Middleware stack**: Sentry → Session → Admin/Redirect → Handler
- **Context keys**: `sessionKey`, `userKey`, `redirectKey` for passing data through middleware
- **Template data**: Handlers create a `baseBag` via `s.newBag(ctx)` which includes session/user
- **Error handling**: Custom error pages via `renderHtml()` with HTTP status codes
- **Dev mode**: Set `DEV=true` to load templates from filesystem instead of embedded FS

## Environment Variables

Defaults defined in 'env' block and loaded automatically by mise.

```
PORT=3000                    # HTTP server port
DATABASE_URL=postgresql://...  # PostgreSQL connection string
REDIS_HOST=localhost:6379    # Redis host:port
DB_TYPE=postgres             # Database type (only postgres supported)
DEV=true                     # Enable dev mode (live template reload)
LOG_LEVEL=info               # Logging level (debug, info, warn, error)
SENTRY_DSN=...               # Sentry error tracking DSN
SENTRY_ENVIRONMENT=dev       # Sentry environment tag
```

User defined environment variables may be set in `.env` file (loaded by mise) and are gitignored.

## Secrets

Required in `.env` file (loaded by mise):

No secrets are currently required to be set.

## Development Setup

1. Start dependencies: `docker compose up -d` (launches Redis and PostgreSQL)
2. Ensure `.env` file exists with required variables (see `mise.toml` for template)
3. Run migrations: `go tool goose up` (or automatic on first run)
4. Build and run: `mise run default`
5. Access at `http://localhost:3000`

## Testing Guidelines

- Tests use `_test.go` suffix and live alongside implementation files
- Unit tests should not require external dependencies (use in-memory cache mock)
- Use `t.Parallel()` for tests that can run concurrently
- Table-driven tests preferred for multiple test cases

## Code Organization

```
main.go                      # Application entry point, initialization, graceful shutdown
internal/
  db/                        # Database connection and migrations
    migrations/              # SQL migration files (goose)
  models/                    # Domain models and database operations
  server/                    # HTTP handlers, routing, templates
    templates/               # Go HTML templates
    assets/                  # Static files (embedded)
  cache/                     # Cache interface and Redis implementation
  events/                    # SSE pub/sub for real-time updates
  authz/                     # Session and authentication logic
```
