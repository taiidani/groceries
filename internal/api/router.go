// Package api provides the REST API server for the groceries application.
// It handles token-based authentication and JSON responses for all API endpoints,
// serving as the backend for native app clients and future frontend migrations.
package api

import (
	"context"
	"fmt"
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/events"
)

// Server is the API server instance.
type Server struct {
	ctx       context.Context
	cache     cache.Cache
	sseServer events.PubSub
}

// NewServer creates a new API server and registers all routes onto the provided mux.
// Routes are mounted under /api/v1/.
func NewServer(ctx context.Context, rds *redis.Client, mux *http.ServeMux) *Server {
	srv := &Server{
		ctx:       ctx,
		cache:     cache.NewRedisCache(rds),
		sseServer: events.NewRedisPubSub(rds),
	}
	srv.addRoutes(mux)
	return srv
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	wrap := func(h http.Handler) http.Handler {
		return sentryHandler.Handle(s.authMiddleware(h))
	}

	// Auth - no token required
	mux.Handle("POST /api/v1/auth/login", sentryHandler.Handle(http.HandlerFunc(s.authLoginHandler)))
	mux.Handle("POST /api/v1/auth/logout", wrap(http.HandlerFunc(s.authLogoutHandler)))
	mux.Handle("GET /api/v1/auth/me", wrap(http.HandlerFunc(s.authMeHandler)))

	// Users (admin only)
	mux.Handle("GET /api/v1/users", wrap(s.adminMiddleware(http.HandlerFunc(s.usersListHandler))))
	mux.Handle("POST /api/v1/users", wrap(s.adminMiddleware(http.HandlerFunc(s.usersCreateHandler))))
	mux.Handle("GET /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersGetHandler))))
	mux.Handle("PUT /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersUpdateHandler))))
	mux.Handle("DELETE /api/v1/users/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.usersDeleteHandler))))

	// Groups (admin only)
	mux.Handle("GET /api/v1/groups", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsListHandler))))
	mux.Handle("POST /api/v1/groups", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsCreateHandler))))
	mux.Handle("GET /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsGetHandler))))
	mux.Handle("PUT /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsUpdateHandler))))
	mux.Handle("DELETE /api/v1/groups/{id}", wrap(s.adminMiddleware(http.HandlerFunc(s.groupsDeleteHandler))))

	// Stores
	mux.Handle("GET /api/v1/stores", wrap(http.HandlerFunc(s.storesListHandler)))
	mux.Handle("POST /api/v1/stores", wrap(http.HandlerFunc(s.storesCreateHandler)))
	mux.Handle("GET /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesGetHandler)))
	mux.Handle("PUT /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesUpdateHandler)))
	mux.Handle("DELETE /api/v1/stores/{id}", wrap(http.HandlerFunc(s.storesDeleteHandler)))

	// Categories
	mux.Handle("GET /api/v1/categories", wrap(http.HandlerFunc(s.categoriesListHandler)))
	mux.Handle("POST /api/v1/categories", wrap(http.HandlerFunc(s.categoriesCreateHandler)))
	mux.Handle("GET /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesGetHandler)))
	mux.Handle("PUT /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesUpdateHandler)))
	mux.Handle("DELETE /api/v1/categories/{id}", wrap(http.HandlerFunc(s.categoriesDeleteHandler)))

	// Items
	mux.Handle("GET /api/v1/items", wrap(http.HandlerFunc(s.itemsListHandler)))
	mux.Handle("POST /api/v1/items", wrap(http.HandlerFunc(s.itemsCreateHandler)))
	mux.Handle("GET /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsGetHandler)))
	mux.Handle("PUT /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsUpdateHandler)))
	mux.Handle("DELETE /api/v1/items/{id}", wrap(http.HandlerFunc(s.itemsDeleteHandler)))

	// Shopping list
	mux.Handle("GET /api/v1/list", wrap(http.HandlerFunc(s.listGetHandler)))
	mux.Handle("POST /api/v1/list/items", wrap(http.HandlerFunc(s.listAddItemHandler)))
	mux.Handle("PUT /api/v1/list/items/{id}", wrap(http.HandlerFunc(s.listUpdateItemHandler)))
	mux.Handle("DELETE /api/v1/list/items/{id}", wrap(http.HandlerFunc(s.listRemoveItemHandler)))
	mux.Handle("POST /api/v1/list/finish", wrap(http.HandlerFunc(s.listFinishHandler)))

	// Recipes
	mux.Handle("GET /api/v1/recipes", wrap(http.HandlerFunc(s.recipesListHandler)))
	mux.Handle("POST /api/v1/recipes", wrap(http.HandlerFunc(s.recipesCreateHandler)))
	mux.Handle("GET /api/v1/recipes/{id}", wrap(http.HandlerFunc(s.recipesGetHandler)))
	mux.Handle("PUT /api/v1/recipes/{id}", wrap(http.HandlerFunc(s.recipesUpdateHandler)))
	mux.Handle("DELETE /api/v1/recipes/{id}", wrap(http.HandlerFunc(s.recipesDeleteHandler)))
	mux.Handle("POST /api/v1/recipes/{id}/items", wrap(http.HandlerFunc(s.recipesAddItemHandler)))
	mux.Handle("DELETE /api/v1/recipes/{id}/items/{itemId}", wrap(http.HandlerFunc(s.recipesRemoveItemHandler)))
	mux.Handle("POST /api/v1/recipes/{id}/add-to-list", wrap(http.HandlerFunc(s.recipesAddToListHandler)))

	// Not found handler for /api/v1/ prefix
	mux.Handle("/api/", sentryHandler.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderError(w, http.StatusNotFound, fmt.Errorf("endpoint not found"))
	})))
}
