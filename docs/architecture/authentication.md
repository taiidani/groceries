# Authentication

Every credential is ultimately backed by Authelia (self-hosted OpenID Connect
provider). All three clients — the web app, the Obsidian plugin, and (soon)
the iOS app — authenticate against Authelia, then exchange the result for
one of the two credential types below, both stored in Redis with shared
resolution logic in `internal/authz`.

## Credential types

| | API token | Web session |
|---|---|---|
| Used by | API clients (Obsidian plugin, future iOS client) (`internal/api`) | Browser (`internal/server`) |
| Transport | `Authorization: Bearer <token>` header | `session` cookie (HttpOnly) |
| Redis key | `api_token:<token>` | `session:<uuid>` |
| Payload | `APIToken{UserID, ExpiresAt}` | `Session{UserID, APIToken}` |
| TTL | 2160 h (90 days) | 2160 h (90 days) |
| Credential source | Authelia OIDC (Device Authorization Grant) | Authelia OIDC (Authorization Code + PKCE) |

A web session **embeds an API token**: login mints both at once. The embedded
token exists so the session and token lifecycles stay coupled — revoking one
on logout revokes both.

## Shared resolution

`authz.ResolveAPIToken(ctx, rawToken, backend)` is the single place that
knows the token cache-key format, performs the lookup, and rejects invalid
entries (`UserID == 0`, expired `ExpiresAt`). The API's `authMiddleware` is a
thin wrapper around it — don't reimplement token lookup elsewhere.

`authz.SyncUserFromOIDC(ctx, db, username, groups)` is the single place that
knows how to find-or-create the local `user` row for an Authelia identity and
(re-)sync the `admin` flag from the `groups` claim. Every login path —
the web OIDC callback, the web dev-login bypass, and the API's token
exchange — calls this instead of duplicating the logic.

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

**API login** (`internal/api/handlers_auth.go`): the client (currently the
Obsidian plugin) completes the OAuth 2.0 Device Authorization Grant against
Authelia entirely on its own — requesting a device/user code, showing the
user a link + code to approve, and polling Authelia's token endpoint — with
no involvement from the Groceries server at all. Once it has an Authelia
access token, it calls `POST /api/v1/auth/login` with `{ access_token }`.
The handler validates that token by calling Authelia's UserInfo endpoint
directly (this requires no client ID/secret on the server's side, since
UserInfo validates the token itself regardless of which client requested
it), reads `preferred_username`/`groups`, calls `authz.SyncUserFromOIDC`, and
mints an `authz.NewAPIToken`. Returns the token as JSON.

Authelia's `groceries-obsidian` client is registered as **public** (no
secret) since a distributed plugin can't keep one confidential; the device
flow doesn't require a redirect URI either, which is what makes it work for
clients (like Obsidian, including its mobile app) that have no way to
receive a browser redirect.

**Web login** (`internal/server/auth.go`, `oidc.go`): Authorization Code + PKCE
against Authelia.

1. `GET /login` renders a landing page with a "Log in with Authelia" link.
2. `GET /auth/login` generates a `state`/`nonce`/PKCE `verifier`, stores them
   in Redis (`oidc_state:<state>`, 10 min TTL) with `state` also set as a
   short-lived, HttpOnly `oidc_state` cookie, then redirects to Authelia's
   authorization endpoint.
3. `GET /auth/callback` requires the `state` query param to match the
   `oidc_state` cookie (CSRF / login-fixation protection), looks up and
   deletes the one-time Redis entry, exchanges the code (with the PKCE
   verifier), and verifies the returned ID token (including the `nonce`).
4. `preferred_username` and `groups` are fetched from Authelia's UserInfo
   endpoint (not the ID token) using the access token — by default Authelia
   only places minimal claims (`sub`, `iss`, `aud`, …) directly in the ID
   token, and exposes profile/groups/email claims via UserInfo instead (the
   same reason Grafana's OIDC client points `api_url` at
   `/api/oidc/userinfo`). The UserInfo response's `sub` is checked against
   the ID token's to guard against token substitution.
5. `authz.SyncUserFromOIDC` resolves `preferred_username` to a local `user`
   row (auto-provisioning one if none exists) and (re-)syncs the `admin`
   column from the `groups` claim on every login — Authelia is the source of
   truth for admin status, though the existing manual toggle in the admin UI
   still works until the next login overwrites it.
6. `authz.NewAPIToken` + `authz.NewSession` proceed exactly as before, so
   `sessionMiddleware`/`adminMiddleware` are unchanged.

A DevMode-only `GET /auth/dev-login?username=<name>` bypass (registered only
when `DEV=true`, never in production) skips the Authelia round-trip entirely
for local dev/E2E testing, but still mints sessions through the same
`establishSession` helper the real callback uses.

**Logout** must tear down *everything*:

- API logout: `authz.RevokeAPIToken` (deletes the Redis key outright).
- Web logout: `authz.DeleteSession(ctx, r, backend)` — deletes the Redis
  session key, revokes the embedded API token, and returns an expired
  cookie. (`DeleteSessionCookie` is the cookie-only helper for cases where
  the backend call isn't possible.) This only ends the Groceries session;
  Authelia's own SSO cookie (shared across `*.taiidani.com`) is left intact
  by design.

## Cache interface

`cache.Cache` (Redis + in-memory implementations) exposes `Set`, `Get`, and
`Delete`. `Delete` exists specifically to support clean revocation — prefer
it over the old pattern of overwriting keys with zero values and short TTLs.

## Not yet migrated

The iOS client (`clients/ios`) still calls the old `{ username, password }`
contract, which no longer exists server-side — it's broken until it's
migrated to the same OIDC-based exchange used by the Obsidian plugin (likely
its own public Authelia client, using either the Device Authorization Grant
or a native redirect-based flow via `ASWebAuthenticationSession`).
