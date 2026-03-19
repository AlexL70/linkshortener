package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultPublicPaths is the set of URL path prefixes that bypass JWT validation.
// Extend this slice when new public-access endpoints are added to the application.
var DefaultPublicPaths = []string{
	"/auth/",        // OAuth2 login / callback / logout / register
	"/docs",         // Swagger UI
	"/openapi.json", // OpenAPI spec
	"/r/",           // Public redirect endpoint (future)
}

// RequireJWTGlobal returns a Gin middleware that enforces JWT authentication on
// all routes whose path does NOT start with one of the given publicPaths prefixes.
// Paths that match a prefix are passed through without any token check.
//
// On success for protected routes it stores *JWTClaims in both the Gin context
// (under CtxKeyJWTClaims) and the stdlib request context (for Huma handlers).
func RequireJWTGlobal(blacklist *TokenBlacklist, publicPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, prefix := range publicPaths {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				c.Next()
				return
			}
		}

		tokenStr, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, parseErr := ParseJWT(tokenStr)
		if parseErr != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if blacklist.IsBlacklisted(claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(string(CtxKeyJWTClaims), claims)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), ctxKeyJWTClaimsStd{}, claims),
		)
		c.Next()
	}
}

// ctxKeyJWTClaimsStd is the key used to store JWT claims in the stdlib request
// context so that Huma handlers (which receive context.Context, not gin.Context)
// can access the validated claims set by RequireJWT middleware.
type ctxKeyJWTClaimsStd struct{}

// GetJWTClaimsFromContext retrieves JWT claims from a stdlib context. Returns
// nil if no claims are present (i.e., the request did not pass through
// RequireJWT middleware).
func GetJWTClaimsFromContext(ctx context.Context) *JWTClaims {
	claims, _ := ctx.Value(ctxKeyJWTClaimsStd{}).(*JWTClaims)
	return claims
}

// contextKey is an unexported type used for keys stored in a Gin context to
// prevent collisions with keys from other packages.
type contextKey string

// CtxKeyJWTClaims is the Gin context key under which the validated *JWTClaims
// are stored by RequireJWT middleware. Downstream handlers retrieve claims with:
//
//	claims, _ := c.MustGet(string(routes.CtxKeyJWTClaims)).(*routes.JWTClaims)
const CtxKeyJWTClaims contextKey = "jwt_claims"

// RequireJWT returns a Gin middleware that validates the JWT from the
// session cookie and rejects the request with 401 Unauthorized if:
//   - the cookie is absent,
//   - the token signature is invalid or the token is expired, or
//   - the token's jti is on the blacklist.
//
// On success it stores *JWTClaims in the Gin context under CtxKeyJWTClaims and
// calls c.Next().
func RequireJWT(blacklist *TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, parseErr := ParseJWT(tokenStr)
		if parseErr != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if blacklist.IsBlacklisted(claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(string(CtxKeyJWTClaims), claims)
		// Also store in the request's stdlib context so Huma handlers can access claims.
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), ctxKeyJWTClaimsStd{}, claims),
		)
		c.Next()
	}
}
