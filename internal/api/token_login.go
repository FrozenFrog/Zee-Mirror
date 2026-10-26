package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Dashboard sign-in with WEB_DASHBOARD_TOKEN, the alternative to OIDC for
// setups without an identity provider.

const (
	// tokenSubject is the sub claim of sessions opened with the token.
	tokenSubject = "token:dashboard"
	// tokenClaim binds a session to the token it was opened with, so
	// changing WEB_DASHBOARD_TOKEN signs everyone out.
	tokenClaim = "tkh"
)

// loginFailures slows down token guessing: failed attempts per client.
var loginFailures = newAPIRateLimiter(10, 15*time.Minute)

func (s *Server) tokenFingerprint(token string) string {
	mac := hmac.New(sha256.New, s.sessionKey)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil)[:16])
}

// validTokenSession checks the token binding of a session opened with
// WEB_DASHBOARD_TOKEN.
func (s *Server) validTokenSession(claims jwt.MapClaims) bool {
	token := s.Service.Config.DashboardToken
	got, _ := claims[tokenClaim].(string)
	return token != "" && hmac.Equal([]byte(got), []byte(s.tokenFingerprint(token)))
}

func (s *Server) handleTokenLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	want := s.Service.Config.DashboardToken
	if want == "" {
		http.NotFound(w, r)
		return
	}
	ip := getClientIP(r)
	w.Header().Set("Content-Type", "application/json")
	if loginFailures.count(ip) >= loginFailures.limit {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"too many attempts"}`))
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request"}`))
		return
	}
	// Comparing digests keeps the token length out of the timing.
	got, exp := sha256.Sum256([]byte(req.Token)), sha256.Sum256([]byte(want))
	if subtle.ConstantTimeCompare(got[:], exp[:]) != 1 {
		loginFailures.Allow(ip)
		slog.Warn("Dashboard token login failed", "ip", ip)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid token"}`))
		return
	}

	signed, err := s.issueDashboardToken(jwt.MapClaims{
		"sub":      tokenSubject,
		tokenClaim: s.tokenFingerprint(want),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to create session"}`))
		return
	}
	slog.Info("Dashboard token login", "ip", ip)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":     signed,
		"expiresIn": int(dashboardTokenTTL.Seconds()),
	})
}
