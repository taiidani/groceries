# Service Layer

`internal/service` is the single home for all domain logic. Both transports
(`internal/api` and `internal/server`) delegate to it; neither contains
validation rules, raw DB calls, or event publishes of its own.

## Package layout

One file per domain:

| File | Domain | Notable methods |
|------|--------|-----------------|
| `list.go` | Shopping list | `GetList`, `AddItem`, `UpdateItem`, `RemoveItem`, `MarkDone`, `Finish` |
| `items.go` | Item catalog | `ListItems`, `GetItem`, `CreateItem`, `UpdateCatalogItem`, `DeleteItem` |
| `categories.go` | Categories | `ListCategories`, `GetCategory`, `CreateCategory`, `UpdateCategory`, `DeleteCategory` |
| `stores.go` | Stores | `ListStores`, `GetStore`, `CreateStore`, `UpdateStore`, `DeleteStore` |
| `hierarchy.go` | Store→category→item aggregation | `LoadStoreHierarchy` |
| `service.go` | Shared plumbing | `Service`, `New`, sentinel errors, `Publisher`, `tx` |

## The Service struct

```go
svc := service.New(conn, publisher) // publisher may be nil (no-op)
```

Constructed once in each transport's `NewServer` and stored on the server
struct (`s.svc`). It holds the `*sql.DB`, the sqlc `*models.Queries`, and the
event publisher.

## Error contract

The service returns sentinel errors wrapped with `%w`; transports classify
them with `errors.Is`:

| Sentinel | Meaning | API maps to | Web maps to |
|----------|---------|-------------|-------------|
| `ErrValidation` | Bad input (missing name, zero ID) | 400 | 400 error partial |
| `ErrNotFound` | Resource doesn't exist | 404 | 404 error partial |
| `ErrConflict` | State conflict (duplicate, in-use FK) | 409 | 409 error partial |

Anything else is an internal error (500). The service never imports
`net/http` — status codes are a transport concern.

Conflict detection uses real PostgreSQL error codes (`23505` unique
violation, `23503` FK violation) rather than string matching.

## Transactions

Multi-step writes run inside `s.tx(ctx, func(q *models.Queries) error {...})`,
which begins a transaction, runs the callback with `queries.WithTx(tx)`, and
commits/rolls back. Operations that are transactional today:

- List add with get-or-create-by-name (fixes a read-then-write race)
- List partial update (quantity + done in one write)
- Item edit that also updates the on-list quantity (`UpdateCatalogItem`)

If a new method performs more than one write, use `s.tx`.

## Event publishing

The service publishes SSE events through the injected `Publisher` interface
(`events.PubSub` satisfies it directly). Publishing is **best-effort**:
failures are logged, never returned. See
[Realtime Events](realtime-events.md) for the channel mapping.

## Testing

Service tests use `github.com/DATA-DOG/go-sqlmock` to stub the database —
see `list_test.go` for the pattern. Construct the service with a nil
publisher (or a fake) so tests don't need Redis.

## Adding a new domain operation

1. Write/extend the sqlc query if needed (see [Data Layer](data-layer.md)).
2. Add the method to the appropriate `internal/service/*.go` file: validate
   input (`ErrValidation`), check existence (`ErrNotFound`), perform writes
   (in `s.tx` if multiple), publish events, return domain types.
3. Add sqlmock tests.
4. Wire both transports (see [Adding an Endpoint](adding-an-endpoint.md)).
