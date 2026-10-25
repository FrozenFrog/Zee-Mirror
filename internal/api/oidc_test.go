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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"zee-mirror/internal/config"
	"zee-mirror/internal/domain"
	"zee-mirror/internal/repository/mocks"
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

func newOIDCServer(t *testing.T, p *fakeProvider, users []domain.User) *Server {
	t.Helper()
	s := newTestServer(t)
	cfg := s.Service.Config
	cfg.OwnerID = 1
	cfg.OIDCIssuer = p.srv.URL
	cfg.OIDCClientID = "zee"
	cfg.OIDCRedirectURL = "https://bot.example" + config.OIDCCallbackPath
	cfg.OIDCProviderName = "Pocket-ID"
	repo := s.Service.DB.(*mocks.MockRepository)
	repo.On("GetAll", mock.Anything).Return(users, nil).Maybe()
	for i := range users {
		repo.On("GetByID", mock.Anything, users[i].ID).Return(&users[i], nil).Maybe()
	}
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
	frag, ok := strings.CutPrefix(location, "/#"+key+"=")
	if !ok {
		return ""
	}
	v, err := url.QueryUnescape(frag)
	require.NoError(t, err)
	return v
}

func TestOIDC_SignInMapsUsername(t *testing.T) {
	p := newFakeProvider(t)
	s := newOIDCServer(t, p, []domain.User{
		{ID: 5, Username: "bob", Role: "admin", IsActive: true},
		{ID: 7, Username: "Alice", Role: "authorized", IsActive: true},
	})

	token := fragmentValue(t, signIn(t, s, p, false), "oidc_token")
	require.NotEmpty(t, token)

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) { return s.sessionKey, nil })
	require.NoError(t, err)
	assert.Equal(t, "7", claims["uid"])
	assert.Equal(t, "Alice", claims["username"])
	assert.True(t, s.validSession(t.Context(), token))
}

func TestOIDC_Rejections(t *testing.T) {
	tests := []struct {
		name     string
		users    []domain.User
		username string
		groups   []string
		allowed  []string
		tamper   bool
	}{
		{"no matching username", []domain.User{{ID: 7, Username: "bob", Role: "admin", IsActive: true}}, "alice", nil, nil, false},
		{"plain user role", []domain.User{{ID: 7, Username: "alice", Role: "user", IsActive: true}}, "alice", nil, nil, false},
		{"expired user", []domain.User{{ID: 7, Username: "alice", Role: "admin", IsActive: false}}, "alice", nil, nil, false},
		{"outside allowed groups", []domain.User{{ID: 7, Username: "alice", Role: "admin", IsActive: true}}, "alice", []string{"staff"}, []string{"zee-admin"}, false},
		{"state mismatch", []domain.User{{ID: 7, Username: "alice", Role: "admin", IsActive: true}}, "alice", nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFakeProvider(t)
			p.username, p.groups = tt.username, tt.groups
			s := newOIDCServer(t, p, tt.users)
			s.Service.Config.OIDCAllowedGroups = tt.allowed

			location := signIn(t, s, p, tt.tamper)
			assert.Empty(t, fragmentValue(t, location, "oidc_token"))
			assert.NotEmpty(t, fragmentValue(t, location, "oidc_error"))
		})
	}
}

func TestOIDC_AllowedGroupAndOwner(t *testing.T) {
	p := newFakeProvider(t)
	p.groups = []string{"Zee-Admin"}
	// The owner is allowed by OWNER_ID even if the stored role differs.
	s := newOIDCServer(t, p, []domain.User{{ID: 1, Username: "alice", Role: "user", IsActive: true}})
	s.Service.Config.OIDCAllowedGroups = []string{"zee-admin"}

	assert.NotEmpty(t, fragmentValue(t, signIn(t, s, p, false), "oidc_token"))
}

func TestOIDC_DisabledAndConfig(t *testing.T) {
	s := newTestServer(t)
	rec := httptest.NewRecorder()
	s.handleOIDCLogin(rec, httptest.NewRequest("GET", "/api/auth/oidc/login", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	s.handleAuthConfig(rec, httptest.NewRequest("GET", "/api/auth/config", nil))
	assert.JSONEq(t, `{"oidc":false,"providerName":""}`, rec.Body.String())
}
