package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) authLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		badRequest(w, "username and password are required")
		return
	}

	user, err := models.GetUserByCredentials(r.Context(), req.Username, req.Password)
	if err != nil {
		// Don't leak whether the user exists vs password was wrong
		errorJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, expiresAt, err := authz.NewAPIToken(r.Context(), user.ID, s.cache)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (s *Server) authLogoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	_, token, _ := strings.Cut(authHeader, " ")

	if err := authz.RevokeAPIToken(r.Context(), token, s.cache); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) authMeHandler(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r.Context())
	if user == nil {
		errorJSON(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Re-fetch to ensure freshness
	fresh, err := models.GetUser(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorJSON(w, http.StatusUnauthorized, "user no longer exists")
			return
		}
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, fresh)
}
