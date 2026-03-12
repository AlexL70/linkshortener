package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/routes"
)

// newMiddlewareRouter builds a minimal Gin engine that applies RequireJWT and
// returns 200 with the user_id claim on success.
func newMiddlewareRouter(bl *routes.TokenBlacklist) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", routes.RequireJWT(bl), func(c *gin.Context) {
		claims, _ := c.MustGet(string(routes.CtxKeyJWTClaims)).(*routes.JWTClaims)
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})
	return r
}

func TestRequireJWT_ValidToken_Returns200(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 1, UserName: "alice"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireJWT_MissingHeader_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWT_MalformedHeader_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token not-a-bearer")
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWT_InvalidToken_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.valid")
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWT_BlacklistedToken_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 2, UserName: "bob"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	// Parse the token to extract the jti, then blacklist it.
	claims, err := routes.ParseJWT(token)
	require.NoError(t, err)
	bl.Add(claims.ID, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWT_ClaimsStoredInContext(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 7, UserName: "carol"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":7`)
}

// --- RequireJWTGlobal tests ---

// newGlobalMiddlewareRouter builds a minimal Gin engine with RequireJWTGlobal applied.
// /public is in the skip-list; /protected is not.
func newGlobalMiddlewareRouter(bl *routes.TokenBlacklist) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(routes.RequireJWTGlobal(bl, []string{"/public"}))
	r.GET("/public", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/protected", func(c *gin.Context) {
		claims, _ := c.MustGet(string(routes.CtxKeyJWTClaims)).(*routes.JWTClaims)
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})
	return r
}

func TestRequireJWTGlobal_PublicPath_NoToken_Returns200(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireJWTGlobal_ProtectedPath_NoToken_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWTGlobal_ProtectedPath_ValidToken_Returns200(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 10, UserName: "heidi"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":10`)
}

func TestRequireJWTGlobal_ProtectedPath_BlacklistedToken_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 11, UserName: "ivan"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	claims, err := routes.ParseJWT(token)
	require.NoError(t, err)
	bl.Add(claims.ID, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWTGlobal_ProtectedPath_InvalidToken_Returns401(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.invalid")
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWTGlobal_ClaimsStoredInContext(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	user := &bizmodels.User{ID: 12, UserName: "judy"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newGlobalMiddlewareRouter(bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":12`)
}
