// Package server provides the HTTP server implementation for the groceries application.
// It handles routing, middleware, template rendering, and HTTP handlers for all application endpoints
// including authentication, shopping lists, items, categories, stores, and admin functionality.
package server

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/db/models"
	"github.com/taiidani/groceries/internal/events"
	"github.com/taiidani/groceries/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	ctx       context.Context
	db        *models.Queries
	cache     cache.Cache
	sseServer events.PubSub
	svc       *service.Service
	*http.Server
}

//go:embed all:templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(ctx context.Context, conn *sql.DB, rds *redis.Client, port string, mux *http.ServeMux) *Server {
	srv := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
		ctx:       ctx,
		db:        models.New(conn),
		cache:     cache.NewRedisCache(rds),
		sseServer: events.NewRedisPubSub(rds),
		svc:       service.New(conn, events.NewRedisPubSub(rds)),
	}
	srv.addRoutes(mux)

	return srv
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	// handle registers a route, wrapping its handler so each request produces a
	// server span named after the matched route. otelhttp also extracts any
	// incoming W3C trace context for distributed tracing.
	handle := func(pattern string, h http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(h, pattern))
	}

	handle("GET /{$}", s.sessionMiddleware(http.HandlerFunc(s.indexHandler)))

	handle("POST /auth", http.HandlerFunc(s.auth))
	handle("GET /login", http.HandlerFunc(s.login))
	handle("GET /logout", http.HandlerFunc(s.logout))

	handle("POST /admin/user/add", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userAddHandler))))
	handle("POST /admin/user/delete/{id}", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userDeleteHandler))))
	handle("POST /admin/user", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userUpdateHandler))))
	handle("POST /admin/group/add", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.groupAddHandler))))
	handle("POST /admin/group/delete/{id}", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.groupDeleteHandler))))
	handle("POST /admin/group", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.groupUpdateHandler))))
	handle("GET /admin", s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.adminHandler))))

	handle("GET /items", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemsHandler))))
	handle("GET /item/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemHandler))))
	handle("POST /item", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemEditHandler))))
	handle("POST /item/add", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemAddHandler))))
	handle("POST /item/delete/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemDeleteHandler))))

	handle("GET /list", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.indexListHandler))))
	handle("POST /list/add", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listAddHandler))))
	handle("POST /list/add/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listAddHandler))))
	handle("POST /list/done", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listDoneHandler))))
	handle("POST /list/undone", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listUnDoneHandler))))
	handle("POST /list/delete/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listDeleteHandler))))
	handle("POST /list/finish", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.finishHandler))))

	handle("GET /cart", s.sessionMiddleware(http.HandlerFunc(s.indexCartHandler)))

	handle("GET /categories", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoriesHandler))))
	handle("GET /category/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryHandler))))
	handle("POST /category", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryEditHandler))))
	handle("POST /category/add", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryAddHandler))))
	handle("POST /category/delete", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryDeleteHandler))))

	handle("GET /stores", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storesHandler))))
	handle("GET /store/{id}", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeHandler))))
	handle("POST /store", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeEditHandler))))
	handle("POST /store/add", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeAddHandler))))
	handle("POST /store/delete", s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeDeleteHandler))))

	handle("GET /sse", s.sessionMiddleware(http.HandlerFunc(s.sseHandler)))

	handle("GET /partials/categories-list-for-store/{id}", s.sessionMiddleware(http.HandlerFunc(s.partialCategoriesListForStoreHandler)))

	handle("/assets/", http.HandlerFunc(s.assetsHandler))
	handle("/apple-touch-icon.png", http.HandlerFunc(s.assetsHandler))

	handle("/", http.HandlerFunc(s.errorNotFoundHandler))
}

func renderHtml(w io.Writer, code int, file string, data any) {
	log := slog.With("name", file, "code", code)

	t, err := getTemplate()
	if err != nil {
		log.Error("Could not parse templates", "error", err)
		return
	}

	log.Debug("Rendering file", "dev", DevMode)
	if writer, ok := w.(http.ResponseWriter); ok {
		writer.WriteHeader(code)
	}
	err = t.ExecuteTemplate(w, file, data)
	if err != nil {
		log.Error("Could not render template", "error", err)
	}
}

func getTemplate() (*template.Template, error) {
	if DevMode {
		return template.ParseGlob("internal/server/templates/**")
	} else {
		return template.ParseFS(templates, "templates/**")
	}
}

type baseBag struct {
	Redirect string
	Session  *authz.Session
	User     *models.User
}

func (s *Server) newBag(ctx context.Context) baseBag {
	ret := baseBag{}

	if redirect, ok := ctx.Value(redirectKey).(string); ok {
		ret.Redirect = redirect
	}

	if sess, ok := ctx.Value(sessionKey).(*authz.Session); ok {
		ret.Session = sess
	}

	if user, ok := ctx.Value(userKey).(*models.User); ok {
		ret.User = user
	}

	return ret
}

func parseId(id string) (int32, error) {
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}
