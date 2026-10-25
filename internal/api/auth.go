package api

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Web dashboard sessions are HS256 JWTs issued after an OIDC sign-in
// (oidc.go). API clients can instead send a user's personal API key.

const dashboardTokenTTL = 24 * time.Hour

// newSessionKey derives the JWT signing key from ENCRYPTION_KEY so sessions
// survive restarts. Without it a random key is used and every restart signs
// everyone out.
func newSessionKey(encryptionKey []byte) []byte {
	if len(encryptionKey) > 0 {
		mac := hmac.New(sha256.New, encryptionKey)
		mac.Write([]byte("zee-mirror dashboard session"))
		return mac.Sum(nil)
	}
	slog.Warn("ENCRYPTION_KEY not set: dashboard sessions end when the bot restarts")
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

func (s *Server) issueDashboardToken(claims jwt.MapClaims) (string, error) {
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(dashboardTokenTTL).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.sessionKey)
}

// validSession checks a session JWT. Group membership was checked at sign-in,
// so removing someone from the group takes effect when their session expires.
func (s *Server) validSession(tokenStr string) bool {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return s.sessionKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return false
	}
	sub, _ := claims["sub"].(string)
	return strings.HasPrefix(sub, oidcSubjectPrefix)
}

// authorizeToken accepts a dashboard session or an active user's personal API key.
func (s *Server) authorizeToken(ctx context.Context, tokenStr string) bool {
	if tokenStr == "" {
		return false
	}
	if s.validSession(tokenStr) {
		return true
	}
	u, err := s.Service.DB.GetUserByAPIKey(ctx, tokenStr)
	return err == nil && u != nil && u.IsActive
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if s.authorizeToken(r.Context(), bearer) || s.authorizeToken(r.Context(), r.Header.Get("X-API-Key")) {
			next(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
	}
}
