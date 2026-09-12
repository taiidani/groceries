# Adding an Endpoint

Checklist for shipping a new feature that touches both the JSON API and the
HTMX web UI. The golden rule: **logic goes in the service; transports only
parse and render.**

## 1. Data access (if needed)

- Add the query to `internal/db/queries/<domain>.sql`, regenerate with
  `mise run generate:sqlc`. See [Data Layer](data-layer.md).

## 2. Service method

- Add the method to `internal/service/<domain>.go`:
  - Validate input → return `ErrValidation` (wrapped with `%w`).
  - Missing resources → `ErrNotFound`; duplicates/in-use → `ErrConflict`
    (detect via pg codes `23505`/`23503`).
  - Multiple writes → wrap in `s.tx`.
  - Publish SSE events per [Realtime Events](realtime-events.md).
- Add sqlmock tests alongside (`<domain>_test.go`).

## 3. API transport (`internal/api`)

- Handler in `handlers_<domain>.go`: decode JSON → call `s.svc` → map errors
  → `writeJSON`. Reuse the existing sentinel-mapping helper
  (`listServiceError` pattern) rather than hand-rolling status codes.
- Register the route in `api.go` under `/api/v1/`, wrapped with
  `authMiddleware`.
- Note: `internal/api` intentionally only exposes what the iOS app and
  Obsidian plugin need (item/list management, read-only store/category
  browsing, auth). Admin-only concerns (user/group management, store and
  category CRUD) are web-app-only — add those to `internal/server` instead,
  not `internal/api`.
- Preserve existing response shapes; if you must change one, check the iOS
  client in `clients/ios/` — breaking changes are acceptable pre-1.0 but
  should be deliberate.
- Document the endpoint in `openapi.yaml`.

## 4. Web transport (`internal/server`)

- Handler in `<domain>.go`: parse form values → call `s.svc` → on error,
  `errorResponse(w, r, status, err)` (it handles `HX-Request` branching); on
  success, render a template or `http.Redirect(..., http.StatusFound)`.
- Register the route in `server.go` with `sessionMiddleware` +
  `redirectMiddleware` for form-posting routes.
- Honor the `redirect` form value when present (fall back to a sensible
  default path) — this is how HTMX flows control post-action navigation.
- Templates live in `internal/server/templates/`; run with `DEV=true` for
  live reload.

## 5. Verify

- `mise test` and `mise lint`.
- Manual smoke: `docker compose up -d && mise run default`, exercise the
  flow in the browser, and (if it mutates list/category state) confirm a
  second open tab updates via SSE.
