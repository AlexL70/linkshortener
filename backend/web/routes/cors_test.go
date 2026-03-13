package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/AlexL70/linkshortener/backend/web/routes"
)

// newCORSRouter builds a minimal Gin router with CORSMiddleware and a stub GET
// handler at "/test", so individual tests can inspect response headers.
func newCORSRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(routes.CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestCORSMiddleware_DevMode_AllowsAnyOrigin(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "dev")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")

	router := newCORSRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_DevMode_PreflightReturns204(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "dev")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")

	router := newCORSRouter(t)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_ProdMode_AllowsMatchingOrigin(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "prod")
	t.Setenv("APP_BASE_URL", "https://example.com")

	router := newCORSRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_ProdMode_BlocksForeignOrigin(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "prod")
	t.Setenv("APP_BASE_URL", "https://example.com")

	router := newCORSRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler still runs (no abort), but no CORS header is set.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_NoOriginHeader_PassesThrough(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "dev")

	router := newCORSRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header — same-origin or non-browser request.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}
