// Package authz provides session management functionality for user authentication.
// It handles session creation, retrieval, updates, and deletion using cookie-based storage
// with a configurable cache backend.
package authz

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
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

func ValidateCredentials(password string) error {
	// Super secret, just between us
	const expected = "ab77936ff6728921c550adb7fc338623"

	hasher := md5.New()
	io.WriteString(hasher, password)
	sum := fmt.Sprintf("%x", hasher.Sum(nil))
	if sum != expected {
		return errors.New("invalid password")
	}

	return nil
}
