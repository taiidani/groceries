package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
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

	slog.Error("Displaying error page", "error", err)
	renderHtml(w, http.StatusNotFound, "error.gohtml", data)
}

func errorResponse(ctx context.Context, writer http.ResponseWriter, code int, err error) {
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

	var hub *sentry.Hub
	if sentry.HasHubOnContext(ctx) {
		hub = sentry.GetHubFromContext(ctx)
	} else {
		hub = sentry.CurrentHub()
	}
	hub.CaptureException(err)

	slog.Error("Displaying error page", "error", err)
	renderHtml(writer, code, "error.gohtml", data)
}
