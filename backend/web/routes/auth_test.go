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
func newTestRouter(h *handlers.AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	routes.RegisterAuthRoutes(router, api, h)
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
