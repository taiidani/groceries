# Authentication

Two credential types, one storage backend (Redis), shared resolution logic in
`internal/authz`.

## Credential types

| | API token | Web session |
|---|---|---|
| Used by | iOS client (`internal/api`) | Browser (`internal/server`) |
| Transport | `Authorization: Bearer <token>` header | `session` cookie (HttpOnly) |
| Redis key | `api_token:<token>` | `session:<uuid>` |
| Payload | `APIToken{UserID, ExpiresAt}` | `Session{UserID, APIToken}` |
| TTL | 720 h (30 days) | 720 h (30 days) |

A web session **embeds an API token**: login mints both at once. The embedded
token exists so the session and token lifecycles stay coupled — revoking one
on logout revokes both.

## Shared resolution

`authz.ResolveAPIToken(ctx, rawToken, backend)` is the single place that
knows the token cache-key format, performs the lookup, and rejects invalid
entries (`UserID == 0`, expired `ExpiresAt`). The API's `authMiddleware` is a
thin wrapper around it — don't reimplement token lookup elsewhere.

## Middleware chains

**API** (`internal/api/middleware.go`):
`authMiddleware` (Bearer → `ResolveAPIToken` → load user → context) →
optional `adminMiddleware`. Failures return JSON 401/403.

**Web** (`internal/server/middleware.go`):
`sessionMiddleware` (cookie → `GetSession` → load user → context) →
optional `adminMiddleware`. Failures **redirect to `/login`** (307), never
render errors. `redirectMiddleware` additionally threads the `redirect`
query param into context for post-action navigation.

Each package has its own private context-key types (`userKey`, etc.) — the
shared surface is the credential resolution, not the middleware.

## Login / logout lifecycle

**Login** (both): validate credentials (`authz.ValidateCredentials` +
`GetUserByName`) → `authz.NewAPIToken` → (web only) `authz.NewSession`
storing the token in the session. The API returns the token as JSON; the web
sets the cookie.

**Logout** must tear down *everything*:

- API logout: `authz.RevokeAPIToken` (deletes the Redis key outright).
- Web logout: `authz.DeleteSession(ctx, r, backend)` — deletes the Redis
  session key, revokes the embedded API token, and returns an expired
  cookie. (`DeleteSessionCookie` is the cookie-only helper for cases where
  the backend call isn't possible.)

## Cache interface

`cache.Cache` (Redis + in-memory implementations) exposes `Set`, `Get`, and
`Delete`. `Delete` exists specifically to support clean revocation — prefer
it over the old pattern of overwriting keys with zero values and short TTLs.
