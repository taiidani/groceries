package models

import "time"

// APIToken holds the data persisted in Redis for a given API bearer token.
// The token string itself is the Redis key; this struct is the value.
type APIToken struct {
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
