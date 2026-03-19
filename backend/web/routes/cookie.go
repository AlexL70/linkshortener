package routes

import (
	"fmt"
	"net/http"
	"os"
)

// SessionCookieName is the name of the HttpOnly session cookie that carries the JWT.
const SessionCookieName = "linkshortener_session"

// setSessionCookie writes an HttpOnly SameSite=Strict cookie carrying the JWT
// to the response. The cookie expires after 24 hours (matching the JWT expiry).
// The Secure flag is set in all environments except dev.
func setSessionCookie(w http.ResponseWriter, token string) {
	isDev := os.Getenv("LINKSHORTENER_ENV") == "dev"
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   !isDev,
		MaxAge:   86400, // 24 hours — matches JWT expiry
	}
	http.SetCookie(w, cookie)
}

// buildSessionCookieHeader returns a formatted Set-Cookie header value for use
// in Huma response structs (which set headers via tagged fields rather than the
// raw http.ResponseWriter). The value matches what setSessionCookie would write.
func buildSessionCookieHeader(token string) string {
	isDev := os.Getenv("LINKSHORTENER_ENV") == "dev"
	secure := ""
	if !isDev {
		secure = "; Secure"
	}
	return fmt.Sprintf(
		"%s=%s; Path=/; HttpOnly; SameSite=Strict; Max-Age=86400%s",
		SessionCookieName, token, secure,
	)
}

// buildClearCookieHeader returns a formatted Set-Cookie header value that
// instructs the browser to delete the session cookie immediately.
func buildClearCookieHeader() string {
	isDev := os.Getenv("LINKSHORTENER_ENV") == "dev"
	secure := ""
	if !isDev {
		secure = "; Secure"
	}
	return fmt.Sprintf(
		"%s=; Path=/; HttpOnly; SameSite=Strict; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT%s",
		SessionCookieName, secure,
	)
}
