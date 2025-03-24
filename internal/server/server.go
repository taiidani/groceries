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
	"strings"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/models"
)

type Server struct {
	ctx       context.Context
	cache     cache.Cache
	publicURL string
	port      string
	sseServer *sseServer
	*http.Server
}

//go:embed templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(ctx context.Context, cache cache.Cache, port string) *Server {
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
		cache:     cache,
		sseServer: newSSEServer(),
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
	mux.Handle("GET /items", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemsHandler))))
	mux.Handle("POST /item/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemAddHandler))))
	mux.Handle("POST /item/bag", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemBagHandler))))
	mux.Handle("POST /item/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemDeleteHandler))))
	mux.Handle("GET /bag", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexBagHandler))))
	mux.Handle("POST /bag/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.bagAddHandler))))
	mux.Handle("POST /bag/update", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.bagUpdateHandler))))
	mux.Handle("POST /bag/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.bagDeleteHandler))))
	mux.Handle("POST /bag/done", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.bagDoneHandler))))
	mux.Handle("GET /list", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexListHandler))))
	mux.Handle("POST /list/done", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemDoneHandler))))
	mux.Handle("POST /list/undone", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemUnDoneHandler))))
	mux.Handle("POST /list/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.listDeleteHandler))))
	mux.Handle("POST /list/finish", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.finishHandler))))
	mux.Handle("GET /cart", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.indexCartHandler))))
	mux.Handle("GET /categories", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoriesHandler))))
	mux.Handle("POST /category/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoryAddHandler))))
	mux.Handle("POST /category/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoryDeleteHandler))))
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

func returnHtml(file string, data any) string {
	w := &strings.Builder{}
	renderHtml(w, 0, file, data)
	return w.String()
}

func getTemplate() (*template.Template, error) {
	if DevMode {
		return template.ParseGlob("internal/server/templates/**")
	} else {
		return template.ParseFS(templates, "templates/**")
	}
}

type baseBag struct {
	Session *models.Session
}

func (s *Server) newBag(ctx context.Context) baseBag {
	ret := baseBag{}

	if sess, ok := ctx.Value(sessionKey).(*models.Session); ok {
		ret.Session = sess
	}

	return ret
}
