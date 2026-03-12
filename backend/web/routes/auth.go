package routes

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
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
//   - POST /auth/logout
//
// Huma-registered routes:
//   - POST /auth/register
func RegisterAuthRoutes(router *gin.Engine, api huma.API, h *handlers.AuthHandler, blacklist *TokenBlacklist) {
	// Configure the gorilla/sessions cookie store for transient OAuth state.
	// HttpOnly and SameSite=Lax per the security checklist in app-auth.md §4.
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
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
		req := gothic.GetContextWithProvider(c.Request, providerName)
		gothic.BeginAuthHandler(c.Writer, req)
	})

	router.GET("/auth/callback", func(c *gin.Context) {
		providerName := c.Query("provider")
		if providerName == "" {
			providerName = string(bizmodels.ProviderGoogle)
		}

		req := gothic.GetContextWithProvider(c.Request, providerName)
		gothUser, err := gothic.CompleteUserAuth(c.Writer, req)
		if err != nil {
			slog.Error("auth: OAuth callback failed", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth authentication failed"})
			return
		}

		// Clear the transient OAuth state cookie immediately after use.
		_ = gothic.Logout(c.Writer, c.Request) //nolint:errcheck — best-effort cleanup

		input := webmappers.GothUserToAuthInput(gothUser, bizmodels.Provider(providerName))
		user, resolveErr := h.ResolveUserByProvider(c.Request.Context(), input)

		if resolveErr == nil {
			// Existing user — issue a full JWT.
			token, jwtErr := CreateJWT(user)
			if jwtErr != nil {
				slog.Error("auth: failed to create JWT", "user_id", user.ID, "error", jwtErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			c.JSON(http.StatusOK, webmappers.UserToAuthTokenResponse(token).Body)
			return
		}

		if errors.Is(resolveErr, businesslogic.ErrNewUser) {
			// New user — issue a pre-registration token for the registration form.
			preRegToken, preRegErr := CreatePreRegToken(input)
			if preRegErr != nil {
				slog.Error("auth: failed to create pre-reg token", "error", preRegErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			c.JSON(http.StatusOK, &viewmodels.PreRegistrationTokenBody{
				PreRegistrationToken: preRegToken,
				SuggestedUserName:    input.DisplayName,
			})
			return
		}

		slog.Error("auth: ResolveUserByProvider failed", "error", resolveErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})

	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Complete new-user registration after OAuth callback",
	}, func(ctx context.Context, input *viewmodels.RegisterRequest) (*viewmodels.AuthTokenResponse, error) {
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

		token, err := CreateJWT(user)
		if err != nil {
			slog.Error("auth: failed to create JWT after registration", "user_id", user.ID, "error", err)
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return webmappers.UserToAuthTokenResponse(token), nil
	})

	// Protected group: all routes here require a valid JWT.
	protected := router.Group("/")
	protected.Use(RequireJWT(blacklist))

	// POST /auth/logout — invalidates the caller's JWT by adding its jti to the
	// blacklist. Returns 204 No Content on success.
	protected.POST("/auth/logout", func(c *gin.Context) {
		claims, _ := c.MustGet(string(CtxKeyJWTClaims)).(*JWTClaims)
		blacklist.Add(claims.ID, claims.ExpiresAt.Time)
		slog.Info("auth: user logged out", "user_id", claims.UserID)
		c.Status(http.StatusNoContent)
	})
}
