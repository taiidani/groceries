# Data Layer

All PostgreSQL access goes through **sqlc**. Hand-written SQL in Go code is
not used anywhere in the application.

## Layout

```
internal/db/
  migrations/   # goose migrations (applied automatically on startup)
  queries/      # sqlc query definitions, one .sql file per domain
  models/       # sqlc-generated Go code (gitignored — do not edit)
  seeds/        # seed data for `mise run seed`
  db.go         # connection setup + auto-migration
```

sqlc is configured in `sqlc.yaml` at the repo root: schema from
`internal/db/migrations`, queries from `internal/db/queries`, output to
`internal/db/models` (package `models`, `database/sql` driver).

## Adding or changing a query

1. Edit the appropriate file in `internal/db/queries/` (or create one).
   Annotate each query with a `-- name: TheName :one|:many|:exec` comment.
2. Regenerate: `mise run generate:sqlc`.
3. The generated `*models.Queries` methods are now available to the service
   layer.

Prefer `RETURNING *` on inserts/updates (`:one`) so callers get the row
without a follow-up select. For joined/summary shapes, write an explicit
column list (see `SummarizeItems` in `queries/item.sql`) rather than
`SELECT *` across joins.

## Migrations

- Apply: `goose up` (also runs automatically via `db.New` on startup)
- Roll back one: `goose down`
- Status: `goose status`

Migration files use `+goose Up` / `+goose Down` directives. Since sqlc reads
the migrations as its schema source, queries and migrations must stay in
sync — regenerate after migrating.

## Conventions

- **IDs are `int32`** in generated code. Transports convert URL path values
  with `parseId` helpers.
- **Conflict detection** inspects `pgconn.PgError.Code` (`23505` unique,
  `23503` foreign key) in the service layer — never string-match error text.
- **Transactions** go through `Service.tx` + `queries.WithTx`
  (see [Service Layer](service-layer.md)); don't call `db.Begin` directly
  outside the service package.
- The generated `models` package contains only row types and queries. Domain
  types that compose rows (e.g. an item with its category name and list
  state) live in `internal/service`.
