package routes

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	webmappers "github.com/AlexL70/linkshortener/backend/web/mappers"
	"github.com/AlexL70/linkshortener/backend/web/viewmodels"
)

// RegisterAuthRoutes sets up the OAuth2/OIDC authentication routes.
//
// Bare Gin routes (permitted exceptions per app-auth.md §2.4):
//   - GET /auth/login/:provider
//   - GET /auth/callback
//
// Huma-registered routes:
//   - POST /auth/register
//   - POST /auth/logout
//   - GET  /auth/me

func RegisterAuthRoutes(router *gin.Engine, api huma.API, h *handlers.AuthHandler, blacklist *TokenBlacklist) {
	// Configure the gorilla/sessions cookie store for transient OAuth state.
	// HttpOnly and SameSite=Lax per the security checklist in app-auth.md §4.
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
		Path:     "/", // must be "/" so the cookie is sent on the callback path too
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("LINKSHORTENER_ENV") != "dev",
		MaxAge:   300, // 5 minutes; OAuth state is transient
	}
	gothic.Store = store

	// Register the Google provider. Other providers are not yet implemented.
	callbackURL := os.Getenv("APP_BASE_URL") + "/auth/callback"
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			callbackURL,
		),
	)

	router.GET("/auth/login/:provider", func(c *gin.Context) {
		providerName := c.Param("provider")
		if bizmodels.Provider(providerName) != bizmodels.ProviderGoogle {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "provider not implemented"})
			return
		}

		// Validate redirect_to and persist it in the session so the callback can
		// pick it up after the OAuth round-trip.
		if redirectTo := c.Query("redirect_to"); redirectTo != "" {
			parsed, err := url.ParseRequestURI(redirectTo)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect_to URL"})
				return
			}

			// SECURITY: Enforce same-origin constraint on redirect_to in production.
			//
			// Without this check, any syntactically valid http/https URL passes the
			// format guard above. After OAuth completes, the backend appends the
			// user's JWT to that URL as a fragment (#token=…) and redirects the
			// browser there — effectively exfiltrating the token to an arbitrary
			// attacker-controlled domain (open redirect + credential leak).
			//
			// In production, Caddy routes both the frontend and the backend under the
			// same domain (APP_BASE_URL), so every legitimate redirect_to must
			// originate from that domain. We compare scheme and host exactly so that
			// look-alike hostnames (e.g. localhost:80801 ≠ localhost:8080) are rejected.
			//
			// In dev mode the frontend runs on a different port (e.g. :5173) while
			// the backend runs on :8080, so APP_BASE_URL does not match the frontend
			// origin. We skip the check in dev — this mirrors the same carve-out in
			// CORSMiddleware (cors.go) and is acceptable because dev is not publicly
			// reachable.
			if os.Getenv("LINKSHORTENER_ENV") != "dev" {
				appBase, appBaseErr := url.ParseRequestURI(os.Getenv("APP_BASE_URL"))
				if appBaseErr != nil || parsed.Scheme != appBase.Scheme || parsed.Host != appBase.Host {
					c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_to URL not allowed"})
					return
				}
			}

			session, err := store.Get(c.Request, "linkshortener_state")
			if err != nil {
				slog.Error("auth: failed to get session for login", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			session.Values["redirect_to"] = redirectTo
			if err := session.Save(c.Request, c.Writer); err != nil {
				slog.Error("auth: failed to save session for login", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}

		req := gothic.GetContextWithProvider(c.Request, providerName)
		gothic.BeginAuthHandler(c.Writer, req)
	})

	router.GET("/auth/callback", func(c *gin.Context) {
		providerName := c.Query("provider")
		if providerName == "" {
			providerName = string(bizmodels.ProviderGoogle)
		}

		// Retrieve and clear redirect_to from the session before any other work so
		// we can use it in error redirects too.
		session, _ := store.Get(c.Request, "linkshortener_state")
		redirectTo, _ := session.Values["redirect_to"].(string)
		delete(session.Values, "redirect_to")
		_ = session.Save(c.Request, c.Writer) //nolint:errcheck — best-effort cleanup

		if redirectTo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing redirect URL"})
			return
		}

		req := gothic.GetContextWithProvider(c.Request, providerName)
		gothUser, err := gothic.CompleteUserAuth(c.Writer, req)
		if err != nil {
			slog.Error("auth: OAuth callback failed", "error", err)
			c.Redirect(http.StatusFound, redirectTo+"#error=authentication_failed")
			return
		}

		// Clear the transient OAuth state cookie immediately after use.
		_ = gothic.Logout(c.Writer, c.Request) //nolint:errcheck — best-effort cleanup

		input := webmappers.GothUserToAuthInput(gothUser, bizmodels.Provider(providerName))
		user, resolveErr := h.ResolveUserByProvider(c.Request.Context(), input)

		if resolveErr == nil {
			// Existing user — issue a full JWT, set it as an HttpOnly session cookie,
			// and redirect to the frontend without exposing the token in the URL.
			token, jwtErr := CreateJWT(user, input.Email)
			if jwtErr != nil {
				slog.Error("auth: failed to create JWT", "user_id", user.ID, "error", jwtErr)
				c.Redirect(http.StatusFound, redirectTo+"#error=internal_error")
				return
			}
			setSessionCookie(c.Writer, token)
			c.Redirect(http.StatusFound, redirectTo)
			return
		}

		if errors.Is(resolveErr, businesslogic.ErrNewUser) {
			// New user — issue a pre-registration token; the frontend will show a
			// registration form and POST to /auth/register.
			preRegToken, preRegErr := CreatePreRegToken(input)
			if preRegErr != nil {
				slog.Error("auth: failed to create pre-reg token", "error", preRegErr)
				c.Redirect(http.StatusFound, redirectTo+"#error=internal_error")
				return
			}
			fragment := "#pre_registration_token=" + preRegToken +
				"&suggested_user_name=" + url.QueryEscape(input.DisplayName)
			c.Redirect(http.StatusFound, redirectTo+fragment)
			return
		}

		slog.Error("auth: ResolveUserByProvider failed", "error", resolveErr)
		c.Redirect(http.StatusFound, redirectTo+"#error=authentication_failed")
	})

	huma.Register(api, huma.Operation{
		OperationID:   "register-user",
		Method:        http.MethodPost,
		Path:          "/auth/register",
		Summary:       "Complete new-user registration after OAuth callback",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, input *viewmodels.RegisterRequest) (*viewmodels.RegisterResponse, error) {
		if input.Body == nil {
			return nil, huma.Error422UnprocessableEntity("request body is required")
		}
		authInput, err := ParsePreRegToken(input.Body.PreRegistrationToken)
		if err != nil {
			return nil, MapError(err)
		}

		user, err := h.CreateUser(ctx, input.Body.UserName, authInput)
		if err != nil {
			return nil, MapError(err)
		}

		token, err := CreateJWT(user, authInput.Email)
		if err != nil {
			slog.Error("auth: failed to create JWT after registration", "user_id", user.ID, "error", err)
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return &viewmodels.RegisterResponse{SetCookie: buildSessionCookieHeader(token)}, nil
	})

	// Protected Huma API: shares the main OpenAPI spec but registers routes on a
	// JWT-gated Gin group. Empty OpenAPIPath/DocsPath/SchemasPath prevent duplicate
	// spec-serving endpoints; huma.DefaultFormats keeps the same JSON serialisation.
	protectedGroup := router.Group("/")
	protectedGroup.Use(RequireJWT(blacklist))
	protectedAPI := humagin.NewWithGroup(router, protectedGroup, huma.Config{
		OpenAPI:       api.OpenAPI(),
		OpenAPIPath:   "",
		DocsPath:      "",
		SchemasPath:   "",
		Formats:       huma.DefaultFormats,
		DefaultFormat: "application/json",
	})

	// POST /auth/logout — invalidates the caller's JWT by adding its jti to the
	// blacklist, and clears the session cookie. Returns 204 No Content on success.
	// JWT validation is enforced by the protectedGroup middleware; claims are
	// forwarded into the stdlib context by RequireJWT for retrieval here.
	huma.Register(protectedAPI, huma.Operation{
		Method:        http.MethodPost,
		Path:          "/auth/logout",
		Summary:       "Log out the authenticated user by blacklisting the JWT",
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"session": {}}},
	}, func(ctx context.Context, _ *struct{}) (*viewmodels.LogoutResponse, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			// Should never happen — RequireJWT middleware always sets claims before
			// this handler runs; defensive guard only.
			return nil, huma.Error401Unauthorized("unauthorized")
		}
		blacklist.Add(claims.ID, claims.ExpiresAt.Time)
		slog.Info("auth: user logged out", "user_id", claims.UserID)
		return &viewmodels.LogoutResponse{SetCookie: buildClearCookieHeader()}, nil
	})

	// DELETE /user/account — permanently deletes the authenticated user's account
	// and all associated data via DB cascade. Returns 204 No Content on success.
	// Super-admin is blocked with 403 Forbidden.
	huma.Register(protectedAPI, huma.Operation{
		OperationID:   "delete-account",
		Method:        http.MethodDelete,
		Path:          "/user/account",
		Summary:       "Permanently delete the authenticated user's account and all associated data",
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"session": {}}},
	}, func(ctx context.Context, _ *struct{}) (*struct{}, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.Error401Unauthorized("unauthorized")
		}
		if err := h.DeleteAccount(ctx, claims.UserID); err != nil {
			return nil, MapError(err)
		}
		return nil, nil
	})

	// GET /auth/me — returns the currently authenticated user's info decoded from
	// the session cookie JWT. Used by the frontend to restore auth state on load.
	huma.Register(protectedAPI, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Return the authenticated user's profile info",
		Security:    []map[string][]string{{"session": {}}},
	}, func(ctx context.Context, _ *struct{}) (*viewmodels.MeResponse, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.Error401Unauthorized("unauthorized")
		}
		return &viewmodels.MeResponse{
			Body: &viewmodels.MeBody{
				UserID:        claims.UserID,
				UserName:      claims.UserName,
				ProviderEmail: claims.Email,
			},
		}, nil
	})
}
