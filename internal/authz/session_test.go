package authz

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/models"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		session models.Session
		devMode bool
		wantErr bool
	}{
		{
			name: "create session for user 1",
			session: models.Session{
				UserID: 1,
			},
			devMode: true,
			wantErr: false,
		},
		{
			name: "create session for user 42",
			session: models.Session{
				UserID: 42,
			},
			devMode: false,
			wantErr: false,
		},
		{
			name: "create session with zero user ID",
			session: models.Session{
				UserID: 0,
			},
			devMode: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := &cache.MemoryStore{Data: make(map[string][]byte)}

			// Set DEV environment variable
			t.Setenv("DEV", fmt.Sprintf("%t", tt.devMode))

			cookie, err := NewSession(ctx, tt.session, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify cookie properties
			if cookie.Name != "session" {
				t.Errorf("cookie.Name = %v, want %v", cookie.Name, "session")
			}

			if cookie.Value == "" {
				t.Error("cookie.Value should not be empty")
			}

			if cookie.Path != "/" {
				t.Errorf("cookie.Path = %v, want %v", cookie.Path, "/")
			}

			if !cookie.HttpOnly {
				t.Error("cookie.HttpOnly should be true")
			}

			expectedSecure := !tt.devMode
			if cookie.Secure != expectedSecure {
				t.Errorf("cookie.Secure = %v, want %v", cookie.Secure, expectedSecure)
			}

			if cookie.MaxAge != int(defaultSessionExpiration.Seconds()) {
				t.Errorf("cookie.MaxAge = %v, want %v", cookie.MaxAge, int(defaultSessionExpiration.Seconds()))
			}

			// Verify session was stored in cache
			var storedSession models.Session
			err = store.Get(ctx, "session:"+cookie.Value, &storedSession)
			if err != nil {
				t.Errorf("Failed to retrieve session from cache: %v", err)
				return
			}

			if storedSession.UserID != tt.session.UserID {
				t.Errorf("stored session.UserID = %v, want %v", storedSession.UserID, tt.session.UserID)
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name    string
		devMode bool
	}{
		{
			name:    "delete session in dev mode",
			devMode: true,
		},
		{
			name:    "delete session in prod mode",
			devMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set DEV environment variable
			t.Setenv("DEV", fmt.Sprintf("%t", tt.devMode))

			cookie := DeleteSession()

			// Verify cookie properties
			if cookie.Name != "session" {
				t.Errorf("cookie.Name = %v, want %v", cookie.Name, "session")
			}

			if cookie.Value != "" {
				t.Errorf("cookie.Value = %v, want empty string", cookie.Value)
			}

			if cookie.Path != "/" {
				t.Errorf("cookie.Path = %v, want %v", cookie.Path, "/")
			}

			if !cookie.HttpOnly {
				t.Error("cookie.HttpOnly should be true")
			}

			expectedSecure := !tt.devMode
			if cookie.Secure != expectedSecure {
				t.Errorf("cookie.Secure = %v, want %v", cookie.Secure, expectedSecure)
			}

			if cookie.MaxAge != -1 {
				t.Errorf("cookie.MaxAge = %v, want -1", cookie.MaxAge)
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name           string
		setupSession   bool
		sessionUserID  int
		sessionKey     string
		noCookie       bool
		wantSession    *models.Session
		wantErr        bool
		wantErrContain string
	}{
		{
			name:          "get existing session",
			setupSession:  true,
			sessionUserID: 123,
			sessionKey:    "valid-session-key",
			wantSession: &models.Session{
				UserID: 123,
			},
			wantErr: false,
		},
		{
			name:          "get session for different user",
			setupSession:  true,
			sessionUserID: 456,
			sessionKey:    "another-session-key",
			wantSession: &models.Session{
				UserID: 456,
			},
			wantErr: false,
		},
		{
			name:        "no cookie present",
			noCookie:    true,
			wantSession: nil,
			wantErr:     false,
		},
		{
			name:           "cookie present but session not in cache",
			setupSession:   false,
			sessionKey:     "non-existent-key",
			wantSession:    nil,
			wantErr:        true,
			wantErrContain: "failed to load session from backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := &cache.MemoryStore{Data: make(map[string][]byte)}

			// Setup session in cache if needed
			if tt.setupSession {
				sess := models.Session{UserID: tt.sessionUserID}
				err := store.Set(ctx, "session:"+tt.sessionKey, sess, time.Hour)
				if err != nil {
					t.Fatalf("Failed to setup session: %v", err)
				}
			}

			// Create request with or without cookie
			req := httptest.NewRequest("GET", "/", nil)
			if !tt.noCookie {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.sessionKey,
				})
			}

			gotSession, err := GetSession(req, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContain != "" {
				if err == nil || len(err.Error()) == 0 {
					t.Errorf("GetSession() expected error containing %q, got nil", tt.wantErrContain)
				}
			}

			if tt.wantSession == nil {
				if gotSession != nil {
					t.Errorf("GetSession() = %v, want nil", gotSession)
				}
			} else {
				if gotSession == nil {
					t.Error("GetSession() = nil, want non-nil session")
					return
				}
				if gotSession.UserID != tt.wantSession.UserID {
					t.Errorf("GetSession().UserID = %v, want %v", gotSession.UserID, tt.wantSession.UserID)
				}
			}
		})
	}
}

func TestUpdateSession(t *testing.T) {
	tests := []struct {
		name           string
		setupSession   bool
		sessionKey     string
		noCookie       bool
		newUserID      int
		wantErr        bool
		wantErrContain string
	}{
		{
			name:         "update existing session",
			setupSession: true,
			sessionKey:   "valid-key",
			newUserID:    999,
			wantErr:      false,
		},
		{
			name:         "update session to different user",
			setupSession: true,
			sessionKey:   "another-key",
			newUserID:    111,
			wantErr:      false,
		},
		{
			name:           "no cookie present",
			noCookie:       true,
			newUserID:      123,
			wantErr:        true,
			wantErrContain: "no session found to update",
		},
		{
			name:         "update session that doesn't exist in cache yet",
			setupSession: false,
			sessionKey:   "new-key",
			newUserID:    777,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := &cache.MemoryStore{Data: make(map[string][]byte)}

			// Setup initial session in cache if needed
			if tt.setupSession {
				sess := models.Session{UserID: 1}
				err := store.Set(ctx, "session:"+tt.sessionKey, sess, time.Hour)
				if err != nil {
					t.Fatalf("Failed to setup session: %v", err)
				}
			}

			// Create request with or without cookie
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(ctx)
			if !tt.noCookie {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.sessionKey,
				})
			}

			// Update the session
			newSession := &models.Session{UserID: tt.newUserID}
			err := UpdateSession(req, newSession, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContain != "" {
				if err == nil || len(err.Error()) == 0 {
					t.Errorf("UpdateSession() expected error containing %q, got nil", tt.wantErrContain)
				}
				return
			}

			if tt.wantErr {
				return
			}

			// Verify the session was updated in cache
			var storedSession models.Session
			err = store.Get(ctx, "session:"+tt.sessionKey, &storedSession)
			if err != nil {
				t.Errorf("Failed to retrieve updated session from cache: %v", err)
				return
			}

			if storedSession.UserID != tt.newUserID {
				t.Errorf("stored session.UserID = %v, want %v", storedSession.UserID, tt.newUserID)
			}
		})
	}
}

func TestSessionLifecycle(t *testing.T) {
	// Integration test: create, get, update, delete
	ctx := context.Background()
	store := &cache.MemoryStore{Data: make(map[string][]byte)}
	t.Setenv("DEV", "true")

	// Create a session
	originalSession := models.Session{UserID: 100}
	cookie, err := NewSession(ctx, originalSession, store)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	// Create a request with the session cookie
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx)
	req.AddCookie(cookie)

	// Get the session
	gotSession, err := GetSession(req, store)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if gotSession.UserID != 100 {
		t.Errorf("GetSession().UserID = %v, want 100", gotSession.UserID)
	}

	// Update the session
	updatedSession := &models.Session{UserID: 200}
	err = UpdateSession(req, updatedSession, store)
	if err != nil {
		t.Fatalf("UpdateSession() error = %v", err)
	}

	// Verify the update
	gotSession, err = GetSession(req, store)
	if err != nil {
		t.Fatalf("GetSession() after update error = %v", err)
	}
	if gotSession.UserID != 200 {
		t.Errorf("GetSession() after update.UserID = %v, want 200", gotSession.UserID)
	}

	// Delete the session (cookie-wise)
	deleteCookie := DeleteSession()
	if deleteCookie.MaxAge != -1 {
		t.Errorf("DeleteSession().MaxAge = %v, want -1", deleteCookie.MaxAge)
	}
}
