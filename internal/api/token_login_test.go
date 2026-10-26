package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDashboardToken = "dashboard-token-0123456789"

func tokenLogin(s *Server, token, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"token":"`+token+`"}`))
	req.RemoteAddr = ip
	rec := httptest.NewRecorder()
	s.handleTokenLogin(rec, req)
	return rec
}

func TestTokenLogin(t *testing.T) {
	s := newTestServer(t)
	s.Service.Config.DashboardToken = testDashboardToken

	rec := tokenLogin(s, "wrong", "10.1.0.1:1")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = tokenLogin(s, testDashboardToken, "10.1.0.1:1")
	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, s.validSession(resp.Token))

	// Changing the token ends sessions opened with the old one.
	s.Service.Config.DashboardToken = testDashboardToken + "-rotated"
	assert.False(t, s.validSession(resp.Token))
	s.Service.Config.DashboardToken = ""
	assert.False(t, s.validSession(resp.Token))
}

func TestTokenLogin_Disabled(t *testing.T) {
	s := newTestServer(t)
	rec := tokenLogin(s, "", "10.1.0.2:1")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTokenLogin_LimitsFailures(t *testing.T) {
	s := newTestServer(t)
	s.Service.Config.DashboardToken = testDashboardToken
	ip := "10.1.0.3:1"
	for i := 0; i < loginFailures.limit; i++ {
		assert.Equal(t, http.StatusUnauthorized, tokenLogin(s, "guess", ip).Code)
	}
	// Even the right token is refused until the window passes.
	assert.Equal(t, http.StatusTooManyRequests, tokenLogin(s, testDashboardToken, ip).Code)
}
