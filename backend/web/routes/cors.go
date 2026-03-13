package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a Gin middleware that adds the necessary CORS response
// headers so the frontend (served from a different port in development) can call
// the backend API.
//
// Behaviour by mode:
//   - dev: any Origin is echoed back as the allowed origin. This accommodates
//     any local frontend port without requiring extra configuration.
//   - prod: only requests whose Origin matches APP_BASE_URL are allowed. If the
//     frontend and backend share the same host in production, no CORS headers are
//     needed at all and this middleware is a no-op for same-origin requests.
//
// The middleware handles OPTIONS preflight requests (which browsers send before
// any cross-origin request that includes custom headers such as Authorization)
// by responding with 204 No Content and the appropriate CORS headers, skipping
// JWT validation.
func CORSMiddleware() gin.HandlerFunc {
	isDevMode := os.Getenv("LINKSHORTENER_ENV") == "dev"
	allowedOriginProd := os.Getenv("APP_BASE_URL")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// Not a cross-origin request — nothing to do.
			c.Next()
			return
		}

		var allow bool
		if isDevMode {
			allow = true
		} else {
			allow = origin == allowedOriginProd
		}

		if !allow {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		// Respond to preflight and stop further handler processing.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
