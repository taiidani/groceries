package server

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/models"
)

type contextKey string

var sessionKey contextKey = "session"

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	bag := s.newBag(r.Context())
	template := "login.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	// Super secret, just between us
	const expected = "ab77936ff6728921c550adb7fc338623"

	hasher := md5.New()
	io.WriteString(hasher, r.FormValue("password"))
	sum := fmt.Sprintf("%x", hasher.Sum(nil))

	if sum != expected {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("bad password"))
		return
	}

	// Yay we're authorized
	sess := models.Session{}
	cookie, err := authz.NewSession(r.Context(), sess, s.cache)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not create session: %w", err))
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie := authz.DeleteSession()
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)

		// Do we have a session already?
		sess, err := authz.GetSession(r, s.cache)
		if err != nil {
			slog.Warn("Failed to retrieve session", "error", err)
		}
		if sess == nil {
			// No session! Login page
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Got a session!
		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
