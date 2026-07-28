# Realtime Events (SSE)

Browser clients receive live updates via Server-Sent Events, backed by Redis
pub/sub. The integration is deliberately minimal: events are signals that
tell the page to refresh a region — they carry no payload.

## Moving parts

- `internal/events` — `PubSub` interface (`Publish`/`Subscribe`) with a
  Redis implementation.
- `internal/service` — the **only** publisher. Event names are constants
  here and are the single source of truth (`internal/server/sse.go` aliases
  them; don't define event names elsewhere).
- `internal/server/sse.go` — the `GET /sse` endpoint. Subscribes to all
  three channels, forwards events to the browser, and sends a `ping` every
  2 s to keep the connection alive.
- Templates use HTMX's `sse` extension to swap content when an event
  arrives.

## Channels

| Constant | Name | Published when |
|----------|------|----------------|
| `EventList` | `list` | Item added to / removed from the shopping list; catalog item created, edited, or deleted |
| `EventCart` | `cart` | List item marked done/undone; shopping finished |
| `EventCategory` | `category` | Category created, edited, or deleted |

Because publishing happens in the service layer, mutations from **either**
transport notify browsers — including changes made by the iOS client.

## Rules for publishers

1. Publish from service methods, never from handlers.
2. Publishing is **best-effort**: `Service.publish` logs failures and never
   fails the operation. Event delivery must not be load-bearing.
3. Don't add payload data — clients only use events as a refresh trigger.

## Adding a new event

1. Add a constant in `internal/service/service.go` (e.g. `EventStore =
   "store"`).
2. Publish it from the relevant service methods.
3. Subscribe to it in `internal/server/sse.go`.
4. Wire the HTMX `sse-swap` (or equivalent) in the templates that should
   react.
