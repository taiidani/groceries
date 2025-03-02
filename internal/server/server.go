package server

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/models"
)

type Server struct {
	backend   data.DB
	publicURL string
	port      string
	*http.Server
}

//go:embed templates
var templates embed.FS

// DevMode can be toggled to pull rendered files from the filesystem or the embedded FS.
var DevMode = os.Getenv("DEV") == "true"

func NewServer(backend data.DB, port string) *Server {
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
		publicURL: publicURL,
		port:      port,
		backend:   backend,
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
	mux.Handle("POST /item/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemAddHandler))))
	mux.Handle("POST /item/done", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemDoneHandler))))
	mux.Handle("POST /item/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.itemDeleteHandler))))
	mux.Handle("POST /finish", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.finishHandler))))
	mux.Handle("GET /categories", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoriesHandler))))
	mux.Handle("POST /category/add", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoryAddHandler))))
	mux.Handle("POST /category/delete", sentryHandler.Handle(s.sessionMiddleware(http.HandlerFunc(s.categoryDeleteHandler))))
	mux.Handle("/assets/", sentryHandler.Handle(http.HandlerFunc(s.assetsHandler)))
	mux.Handle("/", sentryHandler.Handle(http.HandlerFunc(s.errorNotFoundHandler)))
}

func renderHtml(writer http.ResponseWriter, code int, file string, data any) {
	log := slog.With("name", file, "code", code)

	var t *template.Template
	var err error
	if DevMode {
		t, err = template.ParseGlob("internal/server/templates/**")
	} else {
		t, err = template.ParseFS(templates, "templates/**")
	}
	if err != nil {
		log.Error("Could not parse templates", "error", err)
		return
	}

	log.Debug("Rendering file", "dev", DevMode)
	writer.WriteHeader(code)
	err = t.ExecuteTemplate(writer, file, data)
	if err != nil {
		log.Error("Could not render template", "error", err)
	}
}

type baseBag struct {
	Session *models.Session
}

func (s *Server) newBag(r *http.Request) baseBag {
	ret := baseBag{}

	if sess, ok := r.Context().Value(sessionKey).(*models.Session); ok {
		ret.Session = sess
	}

	return ret
}
