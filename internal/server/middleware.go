package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/models"
)

type contextKey string

var (
	sessionKey  contextKey = "session"
	userKey     contextKey = "user"
	redirectKey contextKey = "redirect"
)

func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the context
		user, ok := r.Context().Value(userKey).(*models.User)
		if !ok {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		} else if !user.Admin {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)

		// Do we have a session already?
		sess, err := authz.GetSession(r, s.cache)
		if err != nil {
			slog.Warn("Failed to retrieve session", "error", err)
		}
		if sess == nil || sess.UserID == 0 {
			// No session! Login page
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Got a session! Load the user
		ctx := context.WithValue(r.Context(), sessionKey, sess)
		user, err := models.GetUser(r.Context(), sess.UserID)
		if err != nil {
			slog.Warn("Failed to retrieve user", "error", err)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		ctx = context.WithValue(ctx, userKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) redirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), redirectKey, r.URL.Query().Get("redirect"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
