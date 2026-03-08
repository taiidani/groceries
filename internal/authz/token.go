package authz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/models"
)

const defaultTokenExpiration = time.Duration(time.Hour * 720)

// NewAPIToken generates a cryptographically random Bearer token for the given
// user, stores it in the cache with the standard expiration, and returns the
// raw token string. The caller is responsible for delivering the token to the
// client.
func NewAPIToken(ctx context.Context, userID int, backend cache.Cache) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, fmt.Errorf("could not generate token: %w", err)
	}

	token := hex.EncodeToString(raw)
	expiresAt := time.Now().Add(defaultTokenExpiration)

	data := models.APIToken{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := backend.Set(ctx, apiTokenCacheKey(token), data, defaultTokenExpiration); err != nil {
		return "", time.Time{}, fmt.Errorf("could not store token: %w", err)
	}

	return token, expiresAt, nil
}

// RevokeAPIToken removes a token from the cache, immediately invalidating it.
// Returns nil if the token did not exist.
func RevokeAPIToken(ctx context.Context, token string, backend cache.Cache) error {
	// The cache interface only exposes Set/Get, so we overwrite with a zero-value
	// and a 1-second TTL rather than requiring a Delete method.
	return backend.Set(ctx, apiTokenCacheKey(token), models.APIToken{}, time.Second)
}

// apiTokenCacheKey returns the Redis key used to store an API token.
func apiTokenCacheKey(token string) string {
	return "api_token:" + token
}
