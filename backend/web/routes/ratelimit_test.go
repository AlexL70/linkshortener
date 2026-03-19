package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/AlexL70/linkshortener/backend/web/routes"
)

// newRateLimitRouter builds a minimal Gin engine with RateLimitMiddleware applied
// and a single catch-all route that returns 200 OK. Using a real IP header so
// that Gin's c.ClientIP() returns a stable value during tests.
func newRateLimitRouter(tiers []routes.RateLimitTier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Disable trusted-proxy logic so c.ClientIP() uses RemoteAddr directly.
	_ = r.SetTrustedProxies(nil)
	r.Use(routes.RateLimitMiddleware(tiers))
	r.Any("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

// doRequest fires a single GET request to the given path with the given
// RemoteAddr and returns the HTTP status code.
func doRequest(r *gin.Engine, path, remoteAddr string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = remoteAddr
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimit_WithinLimit_Allowed(t *testing.T) {
	tiers := []routes.RateLimitTier{
		routes.NewRateLimitTier([]string{"/r/"}, 60, 10),
	}
	r := newRateLimitRouter(tiers)

	// 10 requests (= burst size) should all pass immediately.
	for i := 0; i < 10; i++ {
		assert.Equal(t, http.StatusOK, doRequest(r, "/r/abc123", "1.2.3.4:0"), "request %d should be allowed", i+1)
	}
}

func TestRateLimit_ExceedsLimit_Returns429(t *testing.T) {
	// rpm=1, burst=1 means only one token; the second request must be rejected.
	tiers := []routes.RateLimitTier{
		routes.NewRateLimitTier([]string{"/auth/login"}, 1, 1),
	}
	r := newRateLimitRouter(tiers)

	assert.Equal(t, http.StatusOK, doRequest(r, "/auth/login/google", "5.6.7.8:0"), "first request should pass")
	assert.Equal(t, http.StatusTooManyRequests, doRequest(r, "/auth/login/google", "5.6.7.8:0"), "second request should be rate-limited")
}

func TestRateLimit_DifferentIPs_IndependentBuckets(t *testing.T) {
	// rpm=1, burst=1: each IP gets its own token.
	tiers := []routes.RateLimitTier{
		routes.NewRateLimitTier([]string{"/auth/register"}, 1, 1),
	}
	r := newRateLimitRouter(tiers)

	// Both IPs get one allowed request each because their buckets are separate.
	assert.Equal(t, http.StatusOK, doRequest(r, "/auth/register", "10.0.0.1:0"))
	assert.Equal(t, http.StatusOK, doRequest(r, "/auth/register", "10.0.0.2:0"))

	// Each bucket is now empty — both should be limited.
	assert.Equal(t, http.StatusTooManyRequests, doRequest(r, "/auth/register", "10.0.0.1:0"))
	assert.Equal(t, http.StatusTooManyRequests, doRequest(r, "/auth/register", "10.0.0.2:0"))
}

func TestRateLimit_UnmatchedPath_PassesThrough(t *testing.T) {
	// The tier only covers /r/; requests to /user/ must not be rate-limited.
	tiers := []routes.RateLimitTier{
		routes.NewRateLimitTier([]string{"/r/"}, 1, 1),
	}
	r := newRateLimitRouter(tiers)

	// Fire three requests to an unmatched path — all must succeed regardless of
	// the redirect tier's very tight rpm=1/burst=1 limit.
	for i := 0; i < 3; i++ {
		assert.Equal(t, http.StatusOK, doRequest(r, "/user/urls", "1.2.3.4:0"), "request %d to unmatched path should pass", i+1)
	}
}

func TestRateLimit_PrefixSpecificity_NoFalseMatch(t *testing.T) {
	// Auth tier covers only /auth/login and /auth/register.
	// /auth/me must NOT be matched by those prefixes.
	tiers := []routes.RateLimitTier{
		routes.NewRateLimitTier([]string{"/auth/login", "/auth/register"}, 1, 1),
	}
	r := newRateLimitRouter(tiers)

	// Exhaust the auth/login bucket for this IP.
	doRequest(r, "/auth/login/google", "9.9.9.9:0")

	// /auth/me should be unaffected — it matches neither prefix.
	assert.Equal(t, http.StatusOK, doRequest(r, "/auth/me", "9.9.9.9:0"), "/auth/me must not trigger the auth tier limiter")
}

func TestRateLimit_MultiTier_FirstMatchWins(t *testing.T) {
	// Two tiers with different limits; the first tier that matches is used.
	authTier := routes.NewRateLimitTier([]string{"/auth/login"}, 1, 1) // tight
	apiTier := routes.NewRateLimitTier([]string{"/auth/"}, 60, 60)     // generous (should not be reached for /auth/login)
	tiers := []routes.RateLimitTier{authTier, apiTier}
	r := newRateLimitRouter(tiers)

	// First request consumes the auth-tier token.
	assert.Equal(t, http.StatusOK, doRequest(r, "/auth/login/google", "2.2.2.2:0"))
	// Second request hits the tight limit — the generous fallback must NOT apply.
	assert.Equal(t, http.StatusTooManyRequests, doRequest(r, "/auth/login/google", "2.2.2.2:0"))
}
