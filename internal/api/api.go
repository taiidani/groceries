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

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/db/models"
	"github.com/taiidani/groceries/internal/events"
	"github.com/taiidani/groceries/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server is the API server instance.
type Server struct {
	ctx          context.Context
	db           *models.Queries
	cache        cache.Cache
	sseServer    events.PubSub
	svc          *service.Service
	oidcProvider *oidc.Provider
}

// NewServer creates a new API server and registers all routes onto the provided mux.
// Routes are mounted under /api/v1/.
//
// oidcIssuerURL is Authelia's issuer URL, used only to validate access tokens
// via its UserInfo endpoint during login - unlike the web server, the API
// doesn't need a registered client ID/secret of its own for this.
func NewServer(ctx context.Context, conn *sql.DB, rds *redis.Client, mux *http.ServeMux, oidcIssuerURL string) (*Server, error) {
	provider, err := oidc.NewProvider(ctx, oidcIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not discover OIDC provider %q: %w", oidcIssuerURL, err)
	}

	srv := &Server{
		ctx:          ctx,
		db:           models.New(conn),
		cache:        cache.NewRedisCache(rds),
		sseServer:    events.NewRedisPubSub(rds),
		svc:          service.New(conn, events.NewRedisPubSub(rds)),
		oidcProvider: provider,
	}
	srv.addRoutes(mux)
	return srv, nil
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	// wrap applies auth middleware and registers the route so each request
	// produces a server span named after the matched route.
	wrap := func(h http.Handler) http.Handler {
		return s.authMiddleware(h)
	}
	handle := func(pattern string, h http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(h, pattern))
	}

	// Auth - no token required
	handle("POST /api/v1/auth/login", http.HandlerFunc(s.authLoginHandler))
	handle("POST /api/v1/auth/logout", wrap(http.HandlerFunc(s.authLogoutHandler)))
	handle("GET /api/v1/auth/me", wrap(http.HandlerFunc(s.authMeHandler)))

	// Users (admin only)
	handle("GET /api/v1/users", wrap(s.adminMiddleware(http.HandlerFunc(s.usersListHandler))))
	handle("POST /api/v1/users", wrap(s.adminMiddleware(http.HandlerFunc(s.usersCreateHandler))))
	handle("GET /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersGetHandler))))
	handle("PUT /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersUpdateHandler))))
	handle("DELETE /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersDeleteHandler))))

	// Groups (admin only)
	handle("GET /api/v1/groups", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsListHandler))))
	handle("POST /api/v1/groups", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsCreateHandler))))
	handle("GET /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsGetHandler))))
	handle("PUT /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsUpdateHandler))))
	handle("DELETE /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsDeleteHandler))))

	// Stores
	handle("GET /api/v1/stores", wrap(http.HandlerFunc(s.storesListHandler)))
	handle("POST /api/v1/stores", wrap(http.HandlerFunc(s.storesCreateHandler)))
	handle("GET /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesGetHandler)))
	handle("PUT /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesUpdateHandler)))
	handle("DELETE /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesDeleteHandler)))

	// Categories
	handle("GET /api/v1/categories", wrap(http.HandlerFunc(s.categoriesListHandler)))
	handle("POST /api/v1/categories", wrap(http.HandlerFunc(s.categoriesCreateHandler)))
	handle("GET /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesGetHandler)))
	handle("PUT /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesUpdateHandler)))
	handle("DELETE /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesDeleteHandler)))

	// Items
	handle("GET /api/v1/items", wrap(http.HandlerFunc(s.itemsListHandler)))
	handle("POST /api/v1/items", wrap(http.HandlerFunc(s.itemsCreateHandler)))
	handle("GET /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsGetHandler)))
	handle("PUT /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsUpdateHandler)))
	handle("DELETE /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsDeleteHandler)))

	// Shopping list
	handle("GET /api/v1/list", wrap(http.HandlerFunc(s.listGetHandler)))
	handle("POST /api/v1/list/items", wrap(http.HandlerFunc(s.listAddItemHandler)))
	handle("PUT /api/v1/list/items/{id}", wrap(http.HandlerFunc(s.listUpdateItemHandler)))
	handle("DELETE /api/v1/list/items/{id}", wrap(http.HandlerFunc(s.listRemoveItemHandler)))
	handle("POST /api/v1/list/finish", wrap(http.HandlerFunc(s.listFinishHandler)))

	// Not found handler for /api/v1/ prefix
	handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderError(w, http.StatusNotFound, fmt.Errorf("endpoint not found"))
	}))
}

func parseId(id string) (int32, error) {
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}
