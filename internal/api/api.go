// Package api provides the REST API server for the groceries application.
// It implements a token-authenticated JSON API under the /api/v1/ prefix,
// designed to be consumed by both native app clients and webapp frontends.
//
// Authentication is performed via Bearer tokens in the Authorization header.
// Tokens are generated at login time and stored in Redis alongside the web
// session, sharing the same expiration window.
//
// All handlers are independent of the web server's session-based auth and
// HTMX rendering pipeline. The two servers share only the models layer.

package api
