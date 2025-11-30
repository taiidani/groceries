package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/models"
)

type contextKey string

var sessionKey contextKey = "session"
var userKey contextKey = "user"

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	bag := s.newBag(r.Context())
	template := "login.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("username") == "" || r.FormValue("password") == "" {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("missing username or password"))
		return
	}

	// They know the password! Load the user
	user, err := models.GetUserByCredentials(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("invalid credentials: %w", err))
	}

	// Yay we're authorized
	sess := models.Session{UserID: user.ID}
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
