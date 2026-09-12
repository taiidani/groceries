package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/taiidani/groceries/internal/authz"
	"github.com/taiidani/groceries/internal/db/models"
	"golang.org/x/oauth2"
)

// oidcStateCookieName is the short-lived cookie that ties an in-flight OIDC
// login to the browser that started it, preventing an attacker from
// completing their own login flow inside a victim's browser session.
const oidcStateCookieName = "oidc_state"
const oidcStateExpiration = 10 * time.Minute

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	bag := s.newBag(r.Context())
	template := "login.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

// authLogin kicks off the Authorization Code + PKCE flow: it generates and
// persists the state/nonce/verifier, drops a matching cookie, and redirects
// the browser to Authelia.
func (s *Server) authLogin(w http.ResponseWriter, r *http.Request) {
	verifier := oauth2.GenerateVerifier()

	nonce, err := randomString(32)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not generate nonce: %w", err))
		return
	}

	state, err := randomString(32)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not generate state: %w", err))
		return
	}

	stored := oidcState{Verifier: verifier, Nonce: nonce}
	if err := s.cache.Set(r.Context(), "oidc_state:"+state, stored, oidcStateExpiration); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not persist login state: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookieName,
		Value:    state,
		Secure:   !DevMode,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oidcStateExpiration.Seconds()),
	})

	authCodeURL := s.oidc.oauth2.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oidc.Nonce(nonce),
	)
	http.Redirect(w, r, authCodeURL, http.StatusFound)
}

// authCallback completes the Authorization Code + PKCE flow: it validates
// the returned state, exchanges the code, verifies the ID token, and
// provisions/loads the local user before establishing a session exactly as
// the old password flow did.
func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear the short-lived state cookie regardless of outcome - it's single use.
	defer http.SetCookie(w, expireCookie(oidcStateCookieName))

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("authentication failed: %s: %s", errParam, r.URL.Query().Get("error_description")))
		return
	}

	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(oidcStateCookieName)
	if err != nil || state == "" || cookie.Value != state {
		errorResponse(w, r, http.StatusBadRequest, errors.New("invalid or missing login state; please try logging in again"))
		return
	}

	var stored oidcState
	if err := s.cache.Get(ctx, "oidc_state:"+state, &stored); err != nil {
		errorResponse(w, r, http.StatusBadRequest, fmt.Errorf("login session expired; please try again: %w", err))
		return
	}
	_ = s.cache.Delete(ctx, "oidc_state:"+state) // one-time use

	code := r.URL.Query().Get("code")
	if code == "" {
		errorResponse(w, r, http.StatusBadRequest, errors.New("missing authorization code"))
		return
	}

	oauth2Token, err := s.oidc.oauth2.Exchange(ctx, code, oauth2.VerifierOption(stored.Verifier))
	if err != nil {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("could not exchange authorization code: %w", err))
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("token response did not include an id_token"))
		return
	}

	idToken, err := s.oidc.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("could not verify id token: %w", err))
		return
	}

	if idToken.Nonce != stored.Nonce {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("id token nonce mismatch"))
		return
	}

	// Authelia only places minimal claims (sub, iss, aud, ...) directly in the
	// ID token by default; profile/groups/email claims are only available via
	// the UserInfo endpoint (this is the same reason Grafana's OIDC client
	// points at /api/oidc/userinfo instead of decoding the ID token).
	userInfo, err := s.oidc.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("could not fetch user info: %w", err))
		return
	}
	if userInfo.Subject != idToken.Subject {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("user info subject does not match id token subject"))
		return
	}

	var claims struct {
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := userInfo.Claims(&claims); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not parse user info claims: %w", err))
		return
	}
	if claims.PreferredUsername == "" {
		errorResponse(w, r, http.StatusUnauthorized, errors.New("user info is missing the preferred_username claim"))
		return
	}

	user, err := authz.SyncUserFromOIDC(ctx, s.db, claims.PreferredUsername, claims.Groups)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := s.establishSession(ctx, w, user); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// devLogin bypasses the Authelia round-trip entirely, logging in as an
// existing seeded user by name. It is only ever registered when DevMode is
// true (see addRoutes), so it can never be reached in production, and it
// mints sessions through the exact same code path as the real OIDC callback.
func (s *Server) devLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name := r.URL.Query().Get("username")
	if name == "" {
		errorResponse(w, r, http.StatusBadRequest, errors.New("missing username query parameter"))
		return
	}

	user, err := s.db.GetUserByName(ctx, name)
	if err != nil {
		errorResponse(w, r, http.StatusUnauthorized, fmt.Errorf("could not find user %q: %w", name, err))
		return
	}

	if err := s.establishSession(ctx, w, user); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// establishSession mints an API token and a web session for the given user,
// setting the resulting session cookie on the response. Used by both the
// real OIDC callback and the DevMode-only login bypass.
func (s *Server) establishSession(ctx context.Context, w http.ResponseWriter, user models.User) error {
	apiToken, _, err := authz.NewAPIToken(ctx, user.ID, s.cache)
	if err != nil {
		return fmt.Errorf("could not create API token: %w", err)
	}

	sess := authz.Session{UserID: user.ID, APIToken: apiToken}
	sessCookie, err := authz.NewSession(ctx, sess, s.cache)
	if err != nil {
		return fmt.Errorf("could not create session: %w", err)
	}

	http.SetCookie(w, sessCookie)
	return nil
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := authz.DeleteSession(r.Context(), r, s.cache)
	if err != nil {
		slog.WarnContext(r.Context(), "failed to fully delete session", "error", err)
		cookie = authz.DeleteSessionCookie()
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// expireCookie returns an already-expired cookie for the given name, used to
// clear short-lived, single-use cookies (e.g. the OIDC state cookie).
func expireCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Secure:   !DevMode,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
}
