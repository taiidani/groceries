package authz

import (
	"context"
	"testing"
	"time"

	"github.com/taiidani/groceries/internal/cache"
)

func TestResolveAPIToken(t *testing.T) {
	tests := []struct {
		name       string
		setupToken func(ctx context.Context, store cache.Cache) string
		wantErr    bool
	}{
		{
			name: "valid token",
			setupToken: func(ctx context.Context, store cache.Cache) string {
				token, _, err := NewAPIToken(ctx, 42, store)
				if err != nil {
					t.Fatalf("NewAPIToken() error = %v", err)
				}
				return token
			},
			wantErr: false,
		},
		{
			name: "missing token",
			setupToken: func(ctx context.Context, store cache.Cache) string {
				return "does-not-exist"
			},
			wantErr: true,
		},
		{
			name: "revoked token",
			setupToken: func(ctx context.Context, store cache.Cache) string {
				token, _, err := NewAPIToken(ctx, 42, store)
				if err != nil {
					t.Fatalf("NewAPIToken() error = %v", err)
				}
				if err := RevokeAPIToken(ctx, token, store); err != nil {
					t.Fatalf("RevokeAPIToken() error = %v", err)
				}
				return token
			},
			wantErr: true,
		},
		{
			name: "zero-value entry (legacy revoked token)",
			setupToken: func(ctx context.Context, store cache.Cache) string {
				token := "legacy-revoked-token"
				if err := store.Set(ctx, apiTokenCacheKey(token), APIToken{}, time.Second); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
				return token
			},
			wantErr: true,
		},
		{
			name: "expired token",
			setupToken: func(ctx context.Context, store cache.Cache) string {
				token := "expired-token"
				data := APIToken{UserID: 42, ExpiresAt: time.Now().Add(-time.Hour)}
				if err := store.Set(ctx, apiTokenCacheKey(token), data, time.Hour); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
				return token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := &cache.MemoryStore{Data: make(map[string][]byte)}

			token := tt.setupToken(ctx, store)

			got, err := ResolveAPIToken(ctx, token, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAPIToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.UserID != 42 {
				t.Errorf("ResolveAPIToken().UserID = %v, want 42", got.UserID)
			}
		})
	}
}
