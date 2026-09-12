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
| Credential source | Shared static password (`authz.ValidateCredentials`) | Authelia OIDC (Authorization Code + PKCE) |

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

**API login**: validate credentials (`authz.ValidateCredentials` + `GetUserByName`)
→ `authz.NewAPIToken`. Returns the token as JSON. This is the legacy shared-password
flow, retained for the iOS client and Obsidian plugin until they're migrated to OIDC
too.

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
5. `preferred_username` is resolved to a local `user` row via
   `GetUserByName`, auto-provisioning one via `CreateUser` if none exists.
   The `groups` claim is checked for `admins` membership and used to
   (re-)sync the local `user.admin` column on every login — Authelia is the
   source of truth for admin status, though the existing manual toggle in the
   admin UI still works until the next login overwrites it.
6. `authz.NewAPIToken` + `authz.NewSession` proceed exactly as before, so
   `sessionMiddleware`/`adminMiddleware` are unchanged.

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
