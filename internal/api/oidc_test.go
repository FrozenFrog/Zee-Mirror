package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zee-mirror/internal/config"
)

// fakeProvider is a minimal OIDC provider: discovery, JWKS and a token
// endpoint that enforces PKCE.
type fakeProvider struct {
	srv       *httptest.Server
	key       *rsa.PrivateKey
	challenge string
	nonce     string
	username  string
	groups    []string
}

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p := &fakeProvider{key: key, username: "alice"}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                                p.srv.URL,
			"authorization_endpoint":                p.srv.URL + "/authorize",
			"token_endpoint":                        p.srv.URL + "/token",
			"jwks_uri":                              p.srv.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		b64 := base64.RawURLEncoding.EncodeToString
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"keys": []map[string]string{{
			"kty": "RSA", "alg": "RS256", "use": "sig", "kid": "k1",
			"n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if r.Form.Get("code") != "good-code" || base64.RawURLEncoding.EncodeToString(sum[:]) != p.challenge {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": p.srv.URL, "aud": "zee", "sub": "pocket-123",
			"iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
			"nonce": p.nonce, "preferred_username": p.username, "groups": p.groups,
		})
		idToken.Header["kid"] = "k1"
		signed, _ := idToken.SignedString(key)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "at", "token_type": "Bearer", "expires_in": 3600, "id_token": signed,
		})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func newOIDCServer(t *testing.T, p *fakeProvider) *Server {
	t.Helper()
	s := newTestServer(t)
	cfg := s.Service.Config
	cfg.OIDCIssuer = p.srv.URL
	cfg.OIDCClientID = "zee"
	cfg.OIDCRedirectURL = "https://bot.example" + config.OIDCCallbackPath
	cfg.OIDCProviderName = "Pocket-ID"
	cfg.OIDCAllowedGroups = []string{"zee-admin"}
	return s
}

// signIn runs login and callback and returns the callback's redirect target.
func signIn(t *testing.T, s *Server, p *fakeProvider, tamperState bool) string {
	t.Helper()
	rec := httptest.NewRecorder()
	s.handleOIDCLogin(rec, httptest.NewRequest("GET", "/api/auth/oidc/login", nil))
	require.Equal(t, http.StatusFound, rec.Code)

	authURL, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	q := authURL.Query()
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.Equal(t, "zee", q.Get("client_id"))
	assert.Contains(t, q.Get("scope"), "openid")
	p.challenge, p.nonce = q.Get("code_challenge"), q.Get("nonce")

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.True(t, cookies[0].HttpOnly)
	assert.True(t, cookies[0].Secure)

	state := q.Get("state")
	if tamperState {
		state = "attacker-state"
	}
	req := httptest.NewRequest("GET", config.OIDCCallbackPath+"?code=good-code&state="+url.QueryEscape(state), nil)
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	s.handleOIDCCallback(rec, req)
	require.Equal(t, http.StatusFound, rec.Code)
	return rec.Header().Get("Location")
}

func fragmentValue(t *testing.T, location, key string) string {
	t.Helper()
	frag, ok := strings.CutPrefix(location, "/#")
	if !ok {
		return ""
	}
	q, err := url.ParseQuery(frag)
	require.NoError(t, err)
	return q.Get(key)
}

// Group membership alone grants access: no bot user exists for "alice", and
// the repository mock would fail the test if the sign-in looked one up.
func TestOIDC_SignInByGroup(t *testing.T) {
	p := newFakeProvider(t)
	p.groups = []string{"staff", "Zee-Admin"}
	s := newOIDCServer(t, p)

	token := fragmentValue(t, signIn(t, s, p, false), "oidc_token")
	require.NotEmpty(t, token)

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) { return s.sessionKey, nil })
	require.NoError(t, err)
	assert.Equal(t, "oidc:pocket-123", claims["sub"])
	assert.Equal(t, "alice", claims["username"])
	assert.True(t, s.validSession(token))
}

func TestOIDC_Rejections(t *testing.T) {
	tests := []struct {
		name    string
		groups  []string
		allowed []string
		tamper  bool
		code    string
	}{
		{"no groups claim", nil, []string{"zee-admin"}, false, oidcErrNotInGroup},
		{"outside allowed groups", []string{"staff"}, []string{"zee-admin"}, false, oidcErrNotInGroup},
		{"allowed groups not configured", []string{"zee-admin"}, nil, false, oidcErrGroupsNotConfigured},
		{"state mismatch", []string{"zee-admin"}, []string{"zee-admin"}, true, oidcErrInvalidState},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFakeProvider(t)
			p.groups = tt.groups
			s := newOIDCServer(t, p)
			s.Service.Config.OIDCAllowedGroups = tt.allowed

			location := signIn(t, s, p, tt.tamper)
			assert.Empty(t, fragmentValue(t, location, "oidc_token"))
			// A stable code the dashboard translates, not display text.
			assert.Equal(t, tt.code, fragmentValue(t, location, "oidc_error"))
			assert.Empty(t, fragmentValue(t, location, "oidc_provider"))
		})
	}
}

func TestOIDC_UnreachableProviderNamesIt(t *testing.T) {
	p := newFakeProvider(t)
	s := newOIDCServer(t, p)
	p.srv.Close()

	rec := httptest.NewRecorder()
	s.handleOIDCLogin(rec, httptest.NewRequest("GET", "/api/auth/oidc/login", nil))
	require.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Equal(t, oidcErrProviderUnreachable, fragmentValue(t, location, "oidc_error"))
	assert.Equal(t, "Pocket-ID", fragmentValue(t, location, "oidc_provider"))
}

func TestOIDC_DisabledAndConfig(t *testing.T) {
	s := newTestServer(t)
	rec := httptest.NewRecorder()
	s.handleOIDCLogin(rec, httptest.NewRequest("GET", "/api/auth/oidc/login", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	s.handleAuthConfig(rec, httptest.NewRequest("GET", "/api/auth/config", nil))
	assert.JSONEq(t, `{"oidc":false,"token":false,"providerName":""}`, rec.Body.String())
}
