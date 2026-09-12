// Package authz provides session management functionality for user authentication.
// It handles session creation, retrieval, updates, and deletion using cookie-based storage
// with a configurable cache backend.
package authz

import (
	"time"
)

// APIToken holds the data persisted in Redis for a given API bearer token.
// The token string itself is the Redis key; this struct is the value.
type APIToken struct {
	UserID    int32     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Session struct {
	UserID   int32
	APIToken string
}
