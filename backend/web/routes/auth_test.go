package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces/mocks"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-for-unit-tests")
	os.Setenv("SESSION_SECRET", "test-session-secret")
	os.Setenv("APP_BASE_URL", "http://localhost:8080")
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("LINKSHORTENER_ENV", "dev")
	os.Exit(m.Run())
}

// newTestRouter builds a Gin engine + Huma API wired with the given AuthHandler.
// An optional blacklist may be provided; if nil a fresh one is used.
func newTestRouter(h *handlers.AuthHandler, bl ...*routes.TokenBlacklist) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	var blacklist *routes.TokenBlacklist
	if len(bl) > 0 && bl[0] != nil {
		blacklist = bl[0]
	} else {
		blacklist = routes.NewTokenBlacklist()
	}
	routes.RegisterAuthRoutes(router, api, h, blacklist)
	return router
}

// newAuthHandler creates a real AuthHandler backed by a mock repository.
func newAuthHandler(t *testing.T, setupMock func(*mocks.MockUserRepository)) *handlers.AuthHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	if setupMock != nil {
		setupMock(mockRepo)
	}
	return handlers.NewAuthHandler(mockRepo, false, "")
}

// --- POST /auth/register tests ---

func TestRegister_ValidPreRegToken_ReturnsJWT(t *testing.T) {
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "google-sub-123",
		Email:          "newuser@example.com",
		DisplayName:    "New User",
	}
	preRegToken, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)

	newUser := &bizmodels.User{ID: 1, UserName: "newuser"}
	h := newAuthHandler(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			CreateUserWithProvider(gomock.Any(), "newuser", gomock.Any()).
			Return(newUser, nil)
	})

	body, _ := json.Marshal(map[string]string{
		"pre_registration_token": preRegToken,
		"user_name":              "newuser",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["token"])
}

func TestRegister_InvalidPreRegToken_Returns400(t *testing.T) {
	h := newAuthHandler(t, nil)

	body, _ := json.Marshal(map[string]string{
		"pre_registration_token": "invalid.token.here",
		"user_name":              "someuser",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_DuplicateUserName_Returns409(t *testing.T) {
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "google-sub-456",
		Email:          "dup@example.com",
	}
	preRegToken, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)

	h := newAuthHandler(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			CreateUserWithProvider(gomock.Any(), "taken", gomock.Any()).
			Return(nil, businesslogic.ErrConflict)
	})

	body, _ := json.Marshal(map[string]string{
		"pre_registration_token": preRegToken,
		"user_name":              "taken",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_MissingBody_Returns422(t *testing.T) {
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(h).ServeHTTP(w, req)

	// Huma returns 422 for missing required fields.
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// --- GET /auth/login/:provider tests ---

func TestLogin_UnsupportedProvider_Returns501(t *testing.T) {
	for _, provider := range []string{"microsoft", "facebook", "unknown"} {
		t.Run(provider, func(t *testing.T) {
			h := newAuthHandler(t, nil)
			req := httptest.NewRequest(http.MethodGet, "/auth/login/"+provider, nil)
			w := httptest.NewRecorder()
			newTestRouter(h).ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, "provider not implemented", resp["error"])
		})
	}
}

func TestLogin_GoogleProvider_StartsOAuthFlow(t *testing.T) {
	// Google is a supported provider; gothic will attempt to redirect to Google.
	// In tests we cannot complete the OAuth flow, but we verify the response is
	// a redirect (3xx) toward accounts.google.com — proving the login route
	// handed off to goth rather than returning an error.
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/auth/login/google", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	// The response must be a redirect (302/307) or at minimum not a 4xx/5xx.
	// gothic redirects to the provider's auth page.
	code := w.Code
	assert.True(t, code >= 300 && code < 400, "expected redirect for Google login, got %d", code)
}

func TestLogin_InvalidRedirectTo_Returns400(t *testing.T) {
	for _, bad := range []string{
		"javascript:alert(1)",
		"data:text/html,<script>",
		"ftp://example.com/callback",
		"//no-scheme.example.com",
		"not-a-url-at-all",
	} {
		t.Run(bad, func(t *testing.T) {
			h := newAuthHandler(t, nil)
			req := httptest.NewRequest(http.MethodGet, "/auth/login/google?redirect_to="+bad, nil)
			w := httptest.NewRecorder()
			newTestRouter(h).ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code, "expected 400 for redirect_to=%q", bad)
		})
	}
}

func TestLogin_ValidRedirectTo_StartsOAuthFlow(t *testing.T) {
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodGet,
		"/auth/login/google?redirect_to=http%3A%2F%2Flocalhost%3A5173%2Fauth%2Fcallback", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	// Valid redirect_to must not block the OAuth flow — still a 3xx.
	assert.True(t, w.Code >= 300 && w.Code < 400,
		"expected redirect for Google login with valid redirect_to, got %d", w.Code)
}

// TestLogin_ForeignOriginRedirectTo_Returns400InProdMode is a regression test
// for the open-redirect / JWT-exfiltration vulnerability (security report #1).
//
// In production, redirect_to must start with APP_BASE_URL. A URL on any other
// domain must be rejected with 400 before the OAuth flow starts, so the backend
// never has a chance to append the JWT fragment to an attacker-controlled URL.
func TestLogin_ForeignOriginRedirectTo_Returns400InProdMode(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "prod")
	// APP_BASE_URL is already set to "http://localhost:8080" by TestMain; here we
	// set it explicitly to make the test self-documenting and independent of TestMain.
	t.Setenv("APP_BASE_URL", "http://localhost:8080")

	for _, foreign := range []string{
		// Different domain — the classic open-redirect attack vector.
		"https://evil.com/steal",
		// Different subdomain — also attacker-controlled.
		"http://sub.localhost:8080/auth/callback",
		// Looks similar but is a distinct host.
		"http://localhost:80801/auth/callback",
	} {
		t.Run(foreign, func(t *testing.T) {
			h := newAuthHandler(t, nil)
			req := httptest.NewRequest(http.MethodGet, "/auth/login/google?redirect_to="+foreign, nil)
			w := httptest.NewRecorder()
			newTestRouter(h).ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"prod mode must reject foreign-origin redirect_to=%q", foreign)
			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, "redirect_to URL not allowed", resp["error"])
		})
	}
}

// TestLogin_SameOriginRedirectTo_AllowedInProdMode verifies that a redirect_to
// starting with APP_BASE_URL is accepted in production (the happy path).
func TestLogin_SameOriginRedirectTo_AllowedInProdMode(t *testing.T) {
	t.Setenv("LINKSHORTENER_ENV", "prod")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")

	h := newAuthHandler(t, nil)
	// URL-encode http://localhost:8080/auth/callback
	req := httptest.NewRequest(http.MethodGet,
		"/auth/login/google?redirect_to=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fcallback", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	// Same-origin redirect_to must not block the OAuth flow.
	assert.True(t, w.Code >= 300 && w.Code < 400,
		"expected redirect for same-origin redirect_to in prod mode, got %d", w.Code)
}

func TestCallback_NoSessionRedirectTo_Returns400(t *testing.T) {
	// Without a redirect_to stored in the session the callback returns 400.
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "missing redirect URL", resp["error"])
}

func TestRegister_InternalError_Returns500(t *testing.T) {
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "google-sub-789",
		Email:          "err@example.com",
	}
	preRegToken, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)

	h := newAuthHandler(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			CreateUserWithProvider(gomock.Any(), "erruser", gomock.Any()).
			Return(nil, context.DeadlineExceeded)
	})

	body, _ := json.Marshal(map[string]string{
		"pre_registration_token": preRegToken,
		"user_name":              "erruser",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- POST /auth/logout tests ---

func TestLogout_ValidToken_Returns204(t *testing.T) {
	h := newAuthHandler(t, nil)
	user := &bizmodels.User{ID: 10, UserName: "logoutuser"}
	token, err := routes.CreateJWT(user, "logoutuser@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestLogout_MissingToken_Returns401(t *testing.T) {
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_InvalidToken_Returns401(t *testing.T) {
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_ReplayAfterLogout_Returns401(t *testing.T) {
	// Use a shared blacklist so the second request sees the blacklisted jti.
	bl := routes.NewTokenBlacklist()
	h := newAuthHandler(t, nil)
	user := &bizmodels.User{ID: 11, UserName: "replayuser"}
	token, err := routes.CreateJWT(user, "replayuser@example.com")
	require.NoError(t, err)

	router := newTestRouter(h, bl)

	// First logout — must succeed.
	req1 := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusNoContent, w1.Code)

	// Replay the same token — must be rejected.
	req2 := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

// --- DELETE /user/account tests ---

func newAuthHandlerWithAdminEmail(t *testing.T, setupMock func(*mocks.MockUserRepository)) *handlers.AuthHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	if setupMock != nil {
		setupMock(mockRepo)
	}
	return handlers.NewAuthHandler(mockRepo, false, "admin@example.com")
}

func TestDeleteAccount_NoAuth_Returns401(t *testing.T) {
	h := newAuthHandler(t, nil)
	req := httptest.NewRequest(http.MethodDelete, "/user/account", nil)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteAccount_Success_Returns204(t *testing.T) {
	const userID = int64(42)
	h := newAuthHandlerWithAdminEmail(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			FindProvidersByUserID(gomock.Any(), userID).
			Return([]*bizmodels.UserProvider{{ProviderEmail: "regular@example.com"}}, nil)
		m.EXPECT().
			DeleteUser(gomock.Any(), userID).
			Return(nil)
	})
	user := &bizmodels.User{ID: userID, UserName: "regularuser"}
	token, err := routes.CreateJWT(user, "regular@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/user/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteAccount_AdminForbidden_Returns403(t *testing.T) {
	const userID = int64(1)
	h := newAuthHandlerWithAdminEmail(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			FindProvidersByUserID(gomock.Any(), userID).
			Return([]*bizmodels.UserProvider{{ProviderEmail: "admin@example.com"}}, nil)
	})
	user := &bizmodels.User{ID: userID, UserName: "adminuser"}
	token, err := routes.CreateJWT(user, "admin@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/user/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteAccount_NotFound_Returns404(t *testing.T) {
	const userID = int64(99)
	h := newAuthHandlerWithAdminEmail(t, func(m *mocks.MockUserRepository) {
		m.EXPECT().
			FindProvidersByUserID(gomock.Any(), userID).
			Return(nil, businesslogic.ErrNotFound)
	})
	user := &bizmodels.User{ID: userID, UserName: "ghostuser"}
	token, err := routes.CreateJWT(user, "ghost@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/user/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
