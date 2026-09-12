package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/taiidani/groceries/internal/authz"
	"golang.org/x/oauth2"
)

// authLoginHandler exchanges an Authelia OIDC access token (obtained by the
// client via the OAuth 2.0 Device Authorization Grant) for this API's own
// long-lived Bearer token. The access token is validated by calling
// Authelia's UserInfo endpoint directly - this requires no client ID or
// secret on our side, since UserInfo validates the token itself regardless
// of which client requested it.
func (s *Server) authLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body")
		return
	}

	req.AccessToken = strings.TrimSpace(req.AccessToken)
	if req.AccessToken == "" {
		badRequest(w, "access_token is required")
		return
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: req.AccessToken})
	userInfo, err := s.oidcProvider.UserInfo(r.Context(), tokenSource)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "invalid or expired access token")
		return
	}

	var claims struct {
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := userInfo.Claims(&claims); err != nil {
		internalError(w, err)
		return
	}
	if claims.PreferredUsername == "" {
		errorJSON(w, http.StatusUnauthorized, "user info is missing the preferred_username claim")
		return
	}

	user, err := authz.SyncUserFromOIDC(r.Context(), s.db, claims.PreferredUsername, claims.Groups)
	if err != nil {
		internalError(w, err)
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
	fresh, err := s.db.GetUser(r.Context(), user.ID)
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
