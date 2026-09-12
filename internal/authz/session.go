package authz

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/taiidani/groceries/internal/cache"
)

const defaultSessionExpiration = time.Duration(time.Hour * 2160) // 90 days, kept in sync with authz.defaultTokenExpiration

func NewSession(ctx context.Context, sess Session, backend cache.Cache) (*http.Cookie, error) {
	sessionKey := uuid.New().String()
	err := backend.Set(ctx, "session:"+sessionKey, sess, defaultSessionExpiration)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Secure:   os.Getenv("DEV") != "true",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(defaultSessionExpiration.Seconds()),
	}
	return &cookie, nil
}

// DeleteSessionCookie returns an expired cookie that clears the client's
// session cookie.
func DeleteSessionCookie() *http.Cookie {
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Secure:   os.Getenv("DEV") != "true",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	return &cookie
}

// DeleteSession removes the session referenced by the request's cookie from
// the cache backend and returns an expired cookie for the client. If the
// session embeds an API token, that token is revoked as well.
func DeleteSession(ctx context.Context, r *http.Request, backend cache.Cache) (*http.Cookie, error) {
	cookie, err := r.Cookie("session")
	if err == nil {
		var sess Session
		if err := backend.Get(ctx, "session:"+cookie.Value, &sess); err == nil && sess.APIToken != "" {
			if err := RevokeAPIToken(ctx, sess.APIToken, backend); err != nil {
				return nil, fmt.Errorf("could not revoke session token: %w", err)
			}
		}
		if err := backend.Delete(ctx, "session:"+cookie.Value); err != nil {
			return nil, fmt.Errorf("could not delete session: %w", err)
		}
	}

	return DeleteSessionCookie(), nil
}

func GetSession(r *http.Request, cache cache.Cache) (*Session, error) {
	var sess *Session
	cookie, err := r.Cookie("session")
	if err != nil {
		// No cookie 🍪
		return nil, nil
	}

	err = cache.Get(r.Context(), "session:"+cookie.Value, &sess)
	if err != nil {
		return nil, fmt.Errorf("failed to load session from backend: %w", err)
	}

	return sess, nil
}

func UpdateSession(r *http.Request, sess *Session, backend cache.Cache) error {
	cookie, err := r.Cookie("session")
	if err != nil {
		// No cookie 🍪
		return fmt.Errorf("no session found to update")
	}

	err = backend.Set(r.Context(), "session:"+cookie.Value, &sess, defaultSessionExpiration)
	if err != nil {
		return fmt.Errorf("failed to update session in backend: %w", err)
	}

	return nil
}
