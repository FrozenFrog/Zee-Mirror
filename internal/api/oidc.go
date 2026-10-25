package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"zee-mirror/internal/domain"
	"zee-mirror/internal/service"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// Dashboard sign-in through an OpenID Connect provider such as Pocket-ID,
// using the authorization code flow with PKCE. The provider's
// preferred_username is matched against the Telegram username in the users
// table; a match with a dashboard role gets the regular dashboard JWT, so the
// rest of the API needs no changes.

const (
	oidcStateCookie  = "zee_oidc_state"
	oidcCookiePath   = "/api/auth/oidc"
	oidcLoginTTL     = 10 * time.Minute
	maxPendingLogins = 256
)

var oidcHTTPClient = &http.Client{Timeout: 15 * time.Second}

// dashboardRoles are the roles allowed to sign in through OIDC.
var dashboardRoles = map[string]bool{
	service.RoleOwner:      true,
	service.RoleAdmin:      true,
	service.RoleAuthorized: true,
}

type pendingLogin struct {
	verifier string
	nonce    string
	expires  time.Time
}

type oidcAuth struct {
	mu       sync.Mutex
	issuer   string
	provider *oidc.Provider
	pending  map[string]pendingLogin
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

// handleAuthConfig tells the login page whether sign-in is configured.
func (s *Server) handleAuthConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := s.Service.Config
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"oidc":         cfg.OIDCEnabled(),
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
		oidcFail(w, r, "Không kết nối được "+cfg.OIDCProviderName)
		return
	}

	state, nonce := randomToken(), randomToken()
	verifier := oauth2.GenerateVerifier()
	if !s.sso.begin(state, pendingLogin{verifier: verifier, nonce: nonce, expires: time.Now().Add(oidcLoginTTL)}) {
		oidcFail(w, r, "Quá nhiều lượt đăng nhập đang chờ, hãy thử lại sau")
		return
	}

	// The cookie ties the state to this browser, so a callback URL started
	// by someone else cannot log the victim in.
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		Path:     oidcCookiePath,
		MaxAge:   int(oidcLoginTTL.Seconds()),
		HttpOnly: true,
		Secure:   strings.HasPrefix(cfg.OIDCRedirectURL, "https://"),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, s.oauth2Config(provider).AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	cfg := s.Service.Config
	if !cfg.OIDCEnabled() {
		http.NotFound(w, r)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oidcStateCookie, Path: oidcCookiePath, MaxAge: -1, HttpOnly: true})

	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		slog.Warn("OIDC provider returned an error", "error", e, "description", q.Get("error_description"))
		oidcFail(w, r, "Đăng nhập bị hủy hoặc bị từ chối")
		return
	}
	state := q.Get("state")
	cookie, err := r.Cookie(oidcStateCookie)
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		oidcFail(w, r, "Phiên đăng nhập không hợp lệ, hãy thử lại")
		return
	}
	login, ok := s.sso.finish(state)
	if !ok {
		oidcFail(w, r, "Phiên đăng nhập đã hết hạn, hãy thử lại")
		return
	}
	provider, err := s.sso.getProvider(cfg.OIDCIssuer)
	if err != nil {
		slog.Error("OIDC discovery failed", "issuer", cfg.OIDCIssuer, "error", err)
		oidcFail(w, r, "Không kết nối được "+cfg.OIDCProviderName)
		return
	}

	ctx := oidc.ClientContext(r.Context(), oidcHTTPClient)
	token, err := s.oauth2Config(provider).Exchange(ctx, q.Get("code"), oauth2.VerifierOption(login.verifier))
	if err != nil {
		slog.Warn("OIDC code exchange failed", "error", err)
		oidcFail(w, r, "Không xác thực được với "+cfg.OIDCProviderName)
		return
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	idToken, err := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID}).Verify(ctx, rawIDToken)
	if err != nil || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(login.nonce)) != 1 {
		slog.Warn("OIDC ID token rejected", "error", err)
		oidcFail(w, r, "Không xác thực được với "+cfg.OIDCProviderName)
		return
	}
	var claims struct {
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		oidcFail(w, r, "Không đọc được thông tin tài khoản")
		return
	}

	if !inAllowedGroups(claims.Groups, cfg.OIDCAllowedGroups) {
		slog.Warn("OIDC login outside allowed groups", "username", claims.PreferredUsername, "groups", claims.Groups)
		oidcFail(w, r, "Tài khoản không thuộc nhóm được phép")
		return
	}
	user, err := s.userByUsername(r.Context(), claims.PreferredUsername)
	if err != nil {
		slog.Warn("OIDC login without a matching bot user", "username", claims.PreferredUsername, "error", err)
		oidcFail(w, r, "Không có người dùng bot nào có username \""+claims.PreferredUsername+"\" và quyền vào dashboard")
		return
	}

	signed, err := s.issueDashboardToken(jwt.MapClaims{
		"sub":      "oidc:" + idToken.Subject,
		"uid":      strconv.FormatInt(user.ID, 10),
		"username": user.Username,
	})
	if err != nil {
		oidcFail(w, r, "Không tạo được phiên đăng nhập")
		return
	}
	slog.Info("Dashboard OIDC login", "username", user.Username, "uid", user.ID, "ip", getClientIP(r))
	// A fragment never reaches the server or its logs; the dashboard moves
	// it into storage and drops it from the URL.
	http.Redirect(w, r, "/#oidc_token="+url.QueryEscape(signed), http.StatusFound)
}

// userByUsername finds the bot user whose Telegram username matches the
// provider's username (case-insensitive, leading "@" ignored) and who may use
// the dashboard.
func (s *Server) userByUsername(ctx context.Context, username string) (*domain.User, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" {
		return nil, errors.New("provider sent no preferred_username")
	}
	users, err := s.Service.DB.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if strings.EqualFold(strings.TrimPrefix(users[i].Username, "@"), username) {
			if !s.canUseDashboard(&users[i]) {
				return nil, errors.New("user has no dashboard role or is expired")
			}
			return &users[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Server) canUseDashboard(u *domain.User) bool {
	return u.IsActive && (u.ID == s.Service.Config.OwnerID || dashboardRoles[u.Role])
}

func inAllowedGroups(groups, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, g := range groups {
		for _, a := range allowed {
			if strings.EqualFold(g, a) {
				return true
			}
		}
	}
	return false
}

func oidcFail(w http.ResponseWriter, r *http.Request, msg string) {
	http.Redirect(w, r, "/#oidc_error="+url.QueryEscape(msg), http.StatusFound)
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
