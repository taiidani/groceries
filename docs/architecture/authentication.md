# Authentication

Every credential is ultimately backed by Authelia (self-hosted OpenID Connect
provider). All three clients — the web app, the Obsidian plugin, and the iOS
app — authenticate against Authelia, then exchange the result for one of the
two credential types below, both stored in Redis with shared resolution
logic in `internal/authz`.

## Credential types

| | API token | Web session |
|---|---|---|
| Used by | API clients (Obsidian plugin, iOS app) (`internal/api`) | Browser (`internal/server`) |
| Transport | `Authorization: Bearer <token>` header | `session` cookie (HttpOnly) |
| Redis key | `api_token:<token>` | `session:<uuid>` |
| Payload | `APIToken{UserID, ExpiresAt}` | `Session{UserID, APIToken}` |
| TTL | 2160 h (90 days) | 2160 h (90 days) |
| Credential source | Authelia OIDC (Device Authorization Grant for Obsidian; Authorization Code + PKCE via `ASWebAuthenticationSession` for iOS) | Authelia OIDC (Authorization Code + PKCE) |

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
`authMiddleware` (Bearer → `ResolveAPIToken` → load user → context). Failures
return JSON 401. There is no admin-only middleware in the API layer — the API
intentionally exposes no admin-only routes (see [Adding an
Endpoint](adding-an-endpoint.md)).

**Web** (`internal/server/middleware.go`):
`sessionMiddleware` (cookie → `GetSession` → load user → context) →
optional `adminMiddleware`. Failures **redirect to `/login`** (307), never
render errors. `redirectMiddleware` additionally threads the `redirect`
query param into context for post-action navigation.

Each package has its own private context-key types (`userKey`, etc.) — the
shared surface is the credential resolution, not the middleware.

## Login / logout lifecycle

**API login** (`internal/api/handlers_auth.go`): both API clients complete
their own OIDC login against Authelia entirely client-side, with no
involvement from the Groceries server until the very last step:

- **Obsidian** uses the OAuth 2.0 Device Authorization Grant — requesting a
  device/user code, showing the user a link + code to approve, and polling
  Authelia's token endpoint (no redirect target needed; works on Obsidian
  mobile too).
- **iOS** uses Authorization Code + PKCE via `ASWebAuthenticationSession`
  (`clients/ios/Sources/Groceries/Features/Auth/OIDCAuthenticator.swift`) —
  the same flow the web app uses, redirecting back to the app via the
  `com.ryannixon.groceries://auth/callback` custom URL scheme.

Either way, once the client has an Authelia access token, it calls
`POST /api/v1/auth/login` with `{ access_token }`. The handler validates that
token by calling Authelia's UserInfo endpoint directly (this requires no
client ID/secret on the server's side, since UserInfo validates the token
itself regardless of which client requested it), reads
`preferred_username`/`groups`, calls `authz.SyncUserFromOIDC`, and mints an
`authz.NewAPIToken`. Returns the token as JSON.

Both `groceries-obsidian` and `groceries-ios` are registered in Authelia as
**public** clients (no secret) — a distributed app/plugin can't keep one
confidential. `groceries-ios` additionally uses `require_pkce` since,
unlike the device flow, its authorization_code flow needs PKCE to protect
the code exchange in the absence of a client secret.

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

Nothing outstanding — the web app, Obsidian plugin, and iOS app all
authenticate through Authelia via `internal/authz.SyncUserFromOIDC`.
