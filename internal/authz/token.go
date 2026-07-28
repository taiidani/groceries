package authz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/taiidani/groceries/internal/cache"
)

const defaultTokenExpiration = time.Duration(time.Hour * 720)

// NewAPIToken generates a cryptographically random Bearer token for the given
// user, stores it in the cache with the standard expiration, and returns the
// raw token string. The caller is responsible for delivering the token to the
// client.
func NewAPIToken(ctx context.Context, userID int32, backend cache.Cache) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, fmt.Errorf("could not generate token: %w", err)
	}

	token := hex.EncodeToString(raw)
	expiresAt := time.Now().Add(defaultTokenExpiration)

	data := APIToken{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := backend.Set(ctx, apiTokenCacheKey(token), data, defaultTokenExpiration); err != nil {
		return "", time.Time{}, fmt.Errorf("could not store token: %w", err)
	}

	return token, expiresAt, nil
}

// ResolveAPIToken validates a raw Bearer token against the cache backend and
// returns the associated token data. Returns an error if the token is missing,
// expired, or has been revoked.
func ResolveAPIToken(ctx context.Context, rawToken string, backend cache.Cache) (*APIToken, error) {
	var data APIToken
	if err := backend.Get(ctx, apiTokenCacheKey(rawToken), &data); err != nil {
		return nil, fmt.Errorf("could not look up token: %w", err)
	}

	// Revoked tokens briefly persist as a zero-value entry with a 1s TTL;
	// treat them as invalid.
	if data.UserID == 0 {
		return nil, fmt.Errorf("token is revoked or invalid")
	}

	if !data.ExpiresAt.IsZero() && time.Now().After(data.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	return &data, nil
}

// RevokeAPIToken removes a token from the cache, immediately invalidating it.
// Returns nil if the token did not exist.
func RevokeAPIToken(ctx context.Context, token string, backend cache.Cache) error {
	return backend.Delete(ctx, apiTokenCacheKey(token))
}

// apiTokenCacheKey returns the Redis key used to store an API token.
func apiTokenCacheKey(token string) string {
	return "api_token:" + token
}
