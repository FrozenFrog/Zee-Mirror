package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// Dashboard sign-in through an OpenID Connect provider such as Pocket-ID,
// using the authorization code flow with PKCE. Access is decided by the
// provider alone: an account in one of OIDC_ALLOWED_GROUPS gets the dashboard
// JWT. Bot users and their Telegram usernames play no part.

const (
	oidcStateCookie = "zee_oidc_state"
	oidcCookiePath  = "/api/auth/oidc"
	oidcLoginTTL    = 10 * time.Minute
	// oidcSubjectPrefix marks the sub claim of dashboard sessions.
	oidcSubjectPrefix = "oidc:"
	maxPendingLogins  = 256
)

// Error codes sent to the dashboard in /#oidc_error=<code>. They are part of
// the contract with dashboard/src/i18n (keys "oidcError.<code>"); keep both in
// sync.
const (
	oidcErrProviderUnreachable = "provider_unreachable"
	oidcErrTooManyPending      = "too_many_pending"
	oidcErrCancelled           = "cancelled"
	oidcErrInvalidState        = "invalid_state"
	oidcErrExpired             = "expired"
	oidcErrAuthFailed          = "auth_failed"
	oidcErrClaimsUnreadable    = "claims_unreadable"
	oidcErrGroupsNotConfigured = "groups_not_configured"
	oidcErrNotInGroup          = "not_in_group"
	oidcErrSessionFailed       = "session_failed"
)

var oidcHTTPClient = &http.Client{Timeout: 15 * time.Second}

type pendingLogin struct {
	expires  time.Time
	verifier string
	nonce    string
}

type oidcAuth struct {
	provider *oidc.Provider
	pending  map[string]pendingLogin
	issuer   string
	mu       sync.Mutex
}

// getProvider runs discovery once per issuer, on first use, so the bot starts
// even when the provider is unreachable.
func (a *oidcAuth) getProvider(issuer string) (*oidc.Provider, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.provider != nil && a.issuer == issuer {
		return a.provider, nil
	}
	ctx, cancel := context.WithTimeout(oidc.ClientContext(context.Background(), oidcHTTPClient), 15*time.Second)
	defer cancel()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	a.issuer, a.provider = issuer, provider
	return provider, nil
}

func (a *oidcAuth) begin(state string, login pendingLogin) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending == nil {
		a.pending = make(map[string]pendingLogin)
	}
	now := time.Now()
	for k, p := range a.pending {
		if now.After(p.expires) {
			delete(a.pending, k)
		}
	}
	if len(a.pending) >= maxPendingLogins {
		return false
	}
	a.pending[state] = login
	return true
}

func (a *oidcAuth) finish(state string) (pendingLogin, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	p, ok := a.pending[state]
	delete(a.pending, state)
	return p, ok && time.Now().Before(p.expires)
}

func (s *Server) oauth2Config(provider *oidc.Provider) *oauth2.Config {
	cfg := s.Service.Config
	return &oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.OIDCRedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}
}

// handleAuthConfig tells the login page which sign-in methods are configured.
func (s *Server) handleAuthConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := s.Service.Config
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"oidc":         cfg.OIDCEnabled(),
		"token":        cfg.DashboardToken != "",
		"providerName": cfg.OIDCProviderName,
	})
}

func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	cfg := s.Service.Config
	if !cfg.OIDCEnabled() {
		http.NotFound(w, r)
		return
	}
	provider, err := s.sso.getProvider(cfg.OIDCIssuer)
	if err != nil {
		slog.Error("OIDC discovery failed", "issuer", cfg.OIDCIssuer, "error", err)
		oidcFail(w, r, oidcErrProviderUnreachable, cfg.OIDCProviderName)
		return
	}

	state, nonce := randomToken(), randomToken()
	verifier := oauth2.GenerateVerifier()
	if !s.sso.begin(state, pendingLogin{verifier: verifier, nonce: nonce, expires: time.Now().Add(oidcLoginTTL)}) {
		oidcFail(w, r, oidcErrTooManyPending, "")
		return
	}

	// The cookie ties the state to this browser, so a callback URL started
	// by someone else cannot log the victim in.
	http.SetCookie(w, s.stateCookie(state, int(oidcLoginTTL.Seconds())))
	http.Redirect(w, r, s.oauth2Config(provider).AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	cfg := s.Service.Config
	if !cfg.OIDCEnabled() {
		http.NotFound(w, r)
		return
	}
	http.SetCookie(w, s.stateCookie("", -1))

	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		slog.Warn("OIDC provider returned an error", "error", e, "description", q.Get("error_description"))
		oidcFail(w, r, oidcErrCancelled, "")
		return
	}
	state := q.Get("state")
	cookie, err := r.Cookie(oidcStateCookie)
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		oidcFail(w, r, oidcErrInvalidState, "")
		return
	}
	login, ok := s.sso.finish(state)
	if !ok {
		oidcFail(w, r, oidcErrExpired, "")
		return
	}
	provider, err := s.sso.getProvider(cfg.OIDCIssuer)
	if err != nil {
		slog.Error("OIDC discovery failed", "issuer", cfg.OIDCIssuer, "error", err)
		oidcFail(w, r, oidcErrProviderUnreachable, cfg.OIDCProviderName)
		return
	}

	ctx := oidc.ClientContext(r.Context(), oidcHTTPClient)
	token, err := s.oauth2Config(provider).Exchange(ctx, q.Get("code"), oauth2.VerifierOption(login.verifier))
	if err != nil {
		slog.Warn("OIDC code exchange failed", "error", err)
		oidcFail(w, r, oidcErrAuthFailed, cfg.OIDCProviderName)
		return
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	idToken, err := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID}).Verify(ctx, rawIDToken)
	if err != nil || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(login.nonce)) != 1 {
		slog.Warn("OIDC ID token rejected", "error", err)
		oidcFail(w, r, oidcErrAuthFailed, cfg.OIDCProviderName)
		return
	}
	var claims struct {
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err = idToken.Claims(&claims); err != nil {
		oidcFail(w, r, oidcErrClaimsUnreadable, "")
		return
	}

	if len(cfg.OIDCAllowedGroups) == 0 {
		slog.Error("OIDC login refused: OIDC_ALLOWED_GROUPS is not set", "username", claims.PreferredUsername)
		oidcFail(w, r, oidcErrGroupsNotConfigured, "")
		return
	}
	if !inAllowedGroups(claims.Groups, cfg.OIDCAllowedGroups) {
		slog.Warn("OIDC login outside allowed groups", "username", claims.PreferredUsername, "groups", claims.Groups)
		oidcFail(w, r, oidcErrNotInGroup, "")
		return
	}

	signed, err := s.issueDashboardToken(jwt.MapClaims{
		"sub":      oidcSubjectPrefix + idToken.Subject,
		"username": claims.PreferredUsername,
	})
	if err != nil {
		oidcFail(w, r, oidcErrSessionFailed, "")
		return
	}
	slog.Info("Dashboard OIDC login", "username", claims.PreferredUsername, "sub", idToken.Subject, "ip", getClientIP(r))
	// A fragment never reaches the server or its logs; the dashboard moves
	// it into storage and drops it from the URL.
	http.Redirect(w, r, "/#oidc_token="+url.QueryEscape(signed), http.StatusFound)
}

// stateCookie builds the login state cookie; a negative maxAge deletes it.
// Secure follows the callback URL so plain-http setups can still sign in.
func (s *Server) stateCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{ //nolint:gosec // G124: Secure is off only for http:// callback URLs
		Name:     oidcStateCookie,
		Value:    value,
		Path:     oidcCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.Service.Config.OIDCRedirectURL, "https://"),
		SameSite: http.SameSiteLaxMode,
	}
}

func inAllowedGroups(groups, allowed []string) bool {
	for _, g := range groups {
		for _, a := range allowed {
			if strings.EqualFold(g, a) {
				return true
			}
		}
	}
	return false
}

// oidcFail sends the browser back to the dashboard with a stable error code
// (one of the oidcErr* constants) that the login page translates. provider,
// when set, names the identity provider for messages that mention it.
func oidcFail(w http.ResponseWriter, r *http.Request, code, provider string) {
	q := url.Values{"oidc_error": {code}}
	if provider != "" {
		q.Set("oidc_provider", provider)
	}
	http.Redirect(w, r, "/#"+q.Encode(), http.StatusFound)
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
