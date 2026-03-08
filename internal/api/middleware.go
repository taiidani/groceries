package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/models"
)

type contextKey string

var (
	tokenKey contextKey = "api_token"
	userKey  contextKey = "api_user"
)

// authMiddleware validates a Bearer token from the Authorization header,
// looks up the associated user, and places both into the request context.
// Requests without a valid token receive a 401 response.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorJSON(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		scheme, token, found := strings.Cut(authHeader, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
			errorJSON(w, http.StatusUnauthorized, "Authorization header must use the Bearer scheme")
			return
		}

		// Look up the token in the cache
		var tokenData models.APIToken
		err := s.cache.Get(r.Context(), tokenCacheKey(token), &tokenData)
		if err != nil {
			if err == cache.ErrKeyNotFound {
				errorJSON(w, http.StatusUnauthorized, "invalid or expired token")
			} else {
				slog.ErrorContext(r.Context(), "failed to look up API token", "error", err)
				errorJSON(w, http.StatusInternalServerError, "could not validate token")
			}
			return
		}

		// Load the associated user
		user, err := models.GetUser(r.Context(), tokenData.UserID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to load user for API token", "userID", tokenData.UserID, "error", err)
			errorJSON(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), tokenKey, &tokenData)
		ctx = context.WithValue(ctx, userKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminMiddleware ensures the authenticated user has the admin flag set.
// Must be used after authMiddleware.
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userKey).(*models.User)
		if !ok || user == nil {
			errorJSON(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		if !user.Admin {
			errorJSON(w, http.StatusForbidden, "admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// userFromContext retrieves the authenticated user from the request context.
// Returns nil if no user is present (should not happen after authMiddleware).
func userFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userKey).(*models.User)
	return user
}

// tokenCacheKey returns the Redis cache key for a given raw token string.
func tokenCacheKey(token string) string {
	return "api_token:" + token
}
