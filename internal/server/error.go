package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type errorBag struct {
	baseBag
	Title   string
	Message error
}

func (s *Server) errorNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := errors.New("this page does not exist")

	data := errorBag{
		baseBag: baseBag{},
		Title:   "404 Page Not Found",
		Message: err,
	}

	slog.Warn("Displaying error page", "error", err)
	renderHtml(w, http.StatusNotFound, "error.gohtml", data)
}

func errorResponse(w http.ResponseWriter, r *http.Request, code int, err error) {
	title := "Error"
	switch code {
	case http.StatusNotFound:
		title = "404 Page Not Found"
	case http.StatusInternalServerError:
		title = "500 Internal Server Error"
	case http.StatusBadRequest:
		title = "400 Bad Request"
	}

	data := errorBag{
		baseBag: baseBag{},
		Title:   title,
		Message: err,
	}

	// Record the error on the active request span so it surfaces in the
	// trace backend. No-op when no span is present.
	trace.SpanFromContext(r.Context()).RecordError(err)

	if r.Header.Get("HX-Request") != "" {
		slog.WarnContext(r.Context(), "Displaying error message", "error", err)
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(code)
		fmt.Fprintln(w, err.Error())
	} else {
		slog.WarnContext(r.Context(), "Displaying error page", "error", err)
		renderHtml(w, code, "error.gohtml", data)
	}
}
