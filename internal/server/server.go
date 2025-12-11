package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/events"
	"github.com/taiidani/groceries/internal/models"
)

type Server struct {
	ctx       context.Context
	cache     cache.Cache
	publicURL string
	port      string
	sseServer events.PubSub
	*http.Server
}

//go:embed templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(ctx context.Context, rds *redis.Client, port string) *Server {
	mux := http.NewServeMux()

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:" + port
	}

	srv := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		},
		ctx:       ctx,
		publicURL: publicURL,
		port:      port,
		cache:     cache.NewRedisCache(rds),
		sseServer: events.NewRedisPubSub(rds),
	}
	srv.addRoutes(mux)

	return srv
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	mux.Handle("GET /{$}", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexHandler))))

	mux.Handle("POST /auth", sentryHandler.Handle(http.HandlerFunc(s.auth)))
	mux.Handle("GET /login", sentryHandler.Handle(http.HandlerFunc(s.login)))
	mux.Handle("GET /logout", sentryHandler.Handle(http.HandlerFunc(s.logout)))

	mux.Handle("POST /admin/user/add", sentryHandler.Handle(s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userAddHandler)))))
	mux.Handle("POST /admin/user/delete/{id}", sentryHandler.Handle(s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userDeleteHandler)))))
	mux.Handle("POST /admin/user", sentryHandler.Handle(s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.userUpdateHandler)))))
	mux.Handle("GET /admin", sentryHandler.Handle(s.sessionMiddleware(s.adminMiddleware(http.HandlerFunc(s.adminHandler)))))

	mux.Handle("GET /items", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemsHandler)))))
	mux.Handle("GET /item/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemHandler)))))
	mux.Handle("POST /item", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemEditHandler)))))
	mux.Handle("POST /item/add", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemAddHandler)))))
	mux.Handle("POST /item/delete/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.itemDeleteHandler)))))

	mux.Handle("GET /list", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.indexListHandler)))))
	mux.Handle("POST /list/add", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listAddHandler)))))
	mux.Handle("POST /list/add/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listAddHandler)))))
	mux.Handle("POST /list/done", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listDoneHandler)))))
	mux.Handle("POST /list/undone", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listUnDoneHandler)))))
	mux.Handle("POST /list/delete/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.listDeleteHandler)))))
	mux.Handle("POST /list/finish", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.finishHandler)))))

	mux.Handle("GET /cart", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexCartHandler))))

	mux.Handle("GET /categories", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoriesHandler)))))
	mux.Handle("GET /category/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryHandler)))))
	mux.Handle("POST /category", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryEditHandler)))))
	mux.Handle("POST /category/add", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryAddHandler)))))
	mux.Handle("POST /category/delete", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.categoryDeleteHandler)))))

	mux.Handle("GET /stores", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storesHandler)))))
	mux.Handle("GET /store/{id}", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeHandler)))))
	mux.Handle("POST /store", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeEditHandler)))))
	mux.Handle("POST /store/add", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeAddHandler)))))
	mux.Handle("POST /store/delete", sentryHandler.Handle(s.sessionMiddleware(s.redirectMiddleware(http.HandlerFunc(s.storeDeleteHandler)))))

	mux.Handle("GET /sse", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.sseHandler))))

	mux.Handle("/assets/", sentryHandler.Handle(http.HandlerFunc(s.assetsHandler)))
	mux.Handle("/apple-touch-icon.png", sentryHandler.Handle(http.HandlerFunc(s.assetsHandler)))

	mux.Handle("/", sentryHandler.Handle(http.HandlerFunc(s.errorNotFoundHandler)))
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
	Session  *models.Session
	User     *models.User
}

func (s *Server) newBag(ctx context.Context) baseBag {
	ret := baseBag{}

	if redirect, ok := ctx.Value(redirectKey).(string); ok {
		ret.Redirect = redirect
	}

	if sess, ok := ctx.Value(sessionKey).(*models.Session); ok {
		ret.Session = sess
	}

	if user, ok := ctx.Value(userKey).(*models.User); ok {
		ret.User = user
	}

	return ret
}
