package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces/mocks"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/routes"

	"github.com/golang/mock/gomock"
)

// newRedirectTestRouter builds a Gin engine with /r/ in the public paths and
// the redirect route registered.
func newRedirectTestRouter(h *handlers.UrlHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(routes.RequireJWTGlobal(routes.NewTokenBlacklist(), routes.DefaultPublicPaths))
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	routes.RegisterRedirectRoute(router, api, h)
	return router
}

// newRedirectHandlerForTest builds a UrlHandler with a mock repo wired for a
// specific FindByShortcode expectation.
func newRedirectHandlerForTest(t *testing.T, setup func(*mocks.MockUrlRepository)) *handlers.UrlHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUrlRepository(ctrl)
	mockGen := mocks.NewMockShortcodeGenerator(ctrl)
	if setup != nil {
		setup(mockRepo)
	}
	return handlers.NewUrlHandler(mockRepo, mockGen, 2048, 6, 6, 10, nil, false)
}

func TestRedirect_Success(t *testing.T) {
	now := time.Now()
	url := &bizmodels.ShortenedUrl{ID: 1, UserID: 5, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	h := newRedirectHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByShortcode(gomock.Any(), "abc123").Return(url, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/r/abc123", nil)
	w := httptest.NewRecorder()

	newRedirectTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}

func TestRedirect_NotFound_Returns404(t *testing.T) {
	h := newRedirectHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByShortcode(gomock.Any(), "unknwn").Return(nil, businesslogic.ErrNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/r/unknwn", nil)
	w := httptest.NewRecorder()

	newRedirectTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRedirect_Expired_Returns410(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	url := &bizmodels.ShortenedUrl{ID: 2, UserID: 5, Shortcode: "exprd1", LongUrl: "https://example.com", ExpiresAt: &past}

	h := newRedirectHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByShortcode(gomock.Any(), "exprd1").Return(url, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/r/exprd1", nil)
	w := httptest.NewRecorder()

	newRedirectTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
}

func TestRedirect_InternalError_Returns500(t *testing.T) {
	h := newRedirectHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByShortcode(gomock.Any(), "errsc1").Return(nil, assert.AnError)
	})

	req := httptest.NewRequest(http.MethodGet, "/r/errsc1", nil)
	w := httptest.NewRecorder()

	newRedirectTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirect_NoAuthRequired(t *testing.T) {
	// Verify the redirect endpoint is publicly accessible (no JWT needed).
	now := time.Now()
	url := &bizmodels.ShortenedUrl{ID: 3, UserID: 7, Shortcode: "pub123", LongUrl: "https://public.com", LastUpdated: now}

	h := newRedirectHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByShortcode(gomock.Any(), "pub123").Return(url, nil)
	})

	// No Authorization header — should succeed with 302.
	req := httptest.NewRequest(http.MethodGet, "/r/pub123", nil)
	w := httptest.NewRecorder()

	newRedirectTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://public.com", w.Header().Get("Location"))
}
