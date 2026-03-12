package routes_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// newUrlTestRouter builds a Gin engine with RequireJWTGlobal applied (so that
// /user/urls requires a JWT) and the URL routes registered.
func newUrlTestRouter(h *handlers.UrlHandler, bl *routes.TokenBlacklist) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Protect everything except docs/openapi paths; /user/urls requires JWT.
	router.Use(routes.RequireJWTGlobal(bl, []string{"/docs", "/openapi.json"}))
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	routes.RegisterUrlRoutes(api, h)
	return router
}

// newUrlTestRouterNoAuth builds a router with URL routes but WITHOUT JWT middleware.
// Used to test the handler's own defensive nil-claims check.
func newUrlTestRouterNoAuth(h *handlers.UrlHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	routes.RegisterUrlRoutes(api, h)
	return router
}

// newUrlHandler creates a real UrlHandler backed by a mock repository.
func newUrlHandlerForTest(t *testing.T, setupMock func(*mocks.MockUrlRepository)) *handlers.UrlHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUrlRepository(ctrl)
	if setupMock != nil {
		setupMock(mockRepo)
	}
	return handlers.NewUrlHandler(mockRepo)
}

func TestListUserUrls_NoAuth_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil)
	bl := routes.NewTokenBlacklist()

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListUserUrls_ValidJWT_ReturnsURLs(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 1, UserName: "alice"}
	urls := []*bizmodels.ShortenedUrl{
		{ID: 10, UserID: 1, Shortcode: "aaa111", LongUrl: "https://a.com", CreatedAt: now, UpdatedAt: now},
		{ID: 11, UserID: 1, Shortcode: "bbb222", LongUrl: "https://b.com", CreatedAt: now, UpdatedAt: now},
	}

	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(1), 1, 20).Return(urls, 2, nil)
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items    []map[string]interface{} `json:"items"`
		Total    int                      `json:"total"`
		Page     int                      `json:"page"`
		PageSize int                      `json:"page_size"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 2, body.Total)
	assert.Equal(t, 1, body.Page)
	assert.Equal(t, 20, body.PageSize)
	assert.Len(t, body.Items, 2)
}

func TestListUserUrls_CustomPageSize_Respected(t *testing.T) {
	os.Setenv("DEFAULT_PAGE_SIZE", "20")
	user := &bizmodels.User{ID: 2, UserName: "bob"}
	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(2), 1, 5).Return([]*bizmodels.ShortenedUrl{}, 0, nil)
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls?page_size=5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		PageSize int `json:"page_size"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 5, body.PageSize)
}

func TestListUserUrls_PageParam_Respected(t *testing.T) {
	user := &bizmodels.User{ID: 3, UserName: "carol"}
	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(3), 3, 20).Return([]*bizmodels.ShortenedUrl{}, 0, nil)
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls?page=3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListUserUrls_RepoError_Returns500(t *testing.T) {
	user := &bizmodels.User{ID: 4, UserName: "dave"}
	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(4), 1, 20).Return(nil, 0, errors.New("db failure"))
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListUserUrls_NotFoundError_Returns404(t *testing.T) {
	user := &bizmodels.User{ID: 5, UserName: "eve"}
	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(5), 1, 20).Return(nil, 0, businesslogic.ErrNotFound)
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListUserUrls_BlacklistedToken_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil)
	bl := routes.NewTokenBlacklist()

	user := &bizmodels.User{ID: 6, UserName: "frank"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	// Parse and blacklist the token's JTI.
	claims, err := routes.ParseJWT(token)
	require.NoError(t, err)
	bl.Add(claims.ID, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListUserUrls_EmptyList_Returns200(t *testing.T) {
	user := &bizmodels.User{ID: 7, UserName: "grace"}
	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(7), 1, 20).Return([]*bizmodels.ShortenedUrl{}, 0, nil)
	})
	bl := routes.NewTokenBlacklist()

	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []interface{} `json:"items"`
		Total int           `json:"total"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Total)
	assert.Empty(t, body.Items)
}

func TestListUserUrls_NilClaims_Returns401(t *testing.T) {
	// Router without JWT middleware: claims will be nil in the handler,
	// exercising the defensive nil check inside the Huma handler function.
	h := newUrlHandlerForTest(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	w := httptest.NewRecorder()

	newUrlTestRouterNoAuth(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
