package routes

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

// RequireJWT returns a Gin middleware that validates the Bearer token from the
// Authorization header and rejects the request with 401 Unauthorized if:
//   - the header is absent or malformed,
//   - the token signature is invalid or the token is expired, or
//   - the token's jti is on the blacklist.
//
// On success it stores *JWTClaims in the Gin context under CtxKeyJWTClaims and
// calls c.Next().
func RequireJWT(blacklist *TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseJWT(tokenStr)
		if err != nil {
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
