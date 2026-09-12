package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCConfig holds the settings needed to talk to the Authelia OIDC provider.
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	// BaseURL is this application's own externally-reachable URL (e.g.
	// "https://groceries.taiidani.com" or "http://localhost:3000"), used to
	// build the redirect_uri.
	BaseURL string
}

// oidcAuth bundles the pieces needed to run the Authorization Code + PKCE
// flow against Authelia.
type oidcAuth struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth2   oauth2.Config
}

// oidcState is the short-lived, server-side record of an in-flight login,
// keyed by an opaque value handed to the browser via cookie and matched
// against the `state` query param Authelia echoes back on /auth/callback.
type oidcState struct {
	Verifier string
	Nonce    string
}

func newOIDCAuth(ctx context.Context, cfg OIDCConfig) (*oidcAuth, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not discover OIDC provider %q: %w", cfg.IssuerURL, err)
	}

	return &oidcAuth{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth2: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.BaseURL + "/auth/callback",
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		},
	}, nil
}

// randomString returns a URL-safe, base64-encoded string backed by n bytes
// of crypto/rand, suitable for use as an OAuth2 state or nonce value.
func randomString(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("could not generate random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
