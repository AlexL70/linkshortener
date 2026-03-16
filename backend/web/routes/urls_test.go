package routes_test

import (
	"bytes"
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

// newUrlHandlerForTest creates a real UrlHandler backed by mock repo and generator.
func newUrlHandlerForTest(t *testing.T, setupRepo func(*mocks.MockUrlRepository), setupGen func(*mocks.MockShortcodeGenerator)) *handlers.UrlHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUrlRepository(ctrl)
	mockGen := mocks.NewMockShortcodeGenerator(ctrl)
	if setupRepo != nil {
		setupRepo(mockRepo)
	}
	if setupGen != nil {
		setupGen(mockGen)
	}
	return handlers.NewUrlHandler(mockRepo, mockGen, 2048, 6, 6, 10)
}

func TestListUserUrls_NoAuth_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil, nil)
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
		{ID: 10, UserID: 1, Shortcode: "aaa111", LongUrl: "https://a.com",  LastUpdated: now},
		{ID: 11, UserID: 1, Shortcode: "bbb222", LongUrl: "https://b.com",  LastUpdated: now},
	}

	h := newUrlHandlerForTest(t, func(m *mocks.MockUrlRepository) {
		m.EXPECT().FindByUserID(gomock.Any(), int64(1), 1, 20).Return(urls, 2, nil)
	}, nil)
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
	}, nil)
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
	}, nil)
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
	}, nil)
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
	}, nil)
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
	h := newUrlHandlerForTest(t, nil, nil)
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
	}, nil)
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
	h := newUrlHandlerForTest(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/user/urls", nil)
	w := httptest.NewRecorder()

	newUrlTestRouterNoAuth(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── POST /user/urls (create-shortened-url) ────────────────────────────────────

func TestCreateUrl_NoAuth_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil, nil)
	bl := routes.NewTokenBlacklist()

	body := bytes.NewBufferString(`{"long_url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/urls", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateUrl_ValidRequest_Returns201(t *testing.T) {
	os.Setenv("APP_BASE_URL", "https://short.example.com")
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 10, UserName: "alice"}
	createdUrl := &bizmodels.ShortenedUrl{ID: 42, UserID: 10, Shortcode: "ab1234", LongUrl: "https://example.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdUrl, nil)
		},
		func(g *mocks.MockShortcodeGenerator) {
			g.EXPECT().GenerateShortcode().Return("ab1234", nil)
		},
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/urls", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ab1234", resp["shortcode"])
	assert.Equal(t, "https://short.example.com/r/ab1234", resp["short_url"])
}

func TestCreateUrl_CustomShortcode_Returns201(t *testing.T) {
	os.Setenv("APP_BASE_URL", "https://short.example.com")
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 11, UserName: "bob"}
	createdUrl := &bizmodels.ShortenedUrl{ID: 43, UserID: 11, Shortcode: "my-sc1", LongUrl: "https://example.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdUrl, nil)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"https://example.com","shortcode":"my-sc1"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/urls", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "my-sc1", resp["shortcode"])
}

func TestCreateUrl_InvalidUrl_Returns400(t *testing.T) {
	user := &bizmodels.User{ID: 12, UserName: "carol"}
	h := newUrlHandlerForTest(t, nil, nil)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"ftp://bad.scheme.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/urls", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUrl_ConflictShortcode_Returns409(t *testing.T) {
	user := &bizmodels.User{ID: 13, UserName: "dave"}
	sc := "taken1"
	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, businesslogic.ErrConflict)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"https://example.com","shortcode":"` + sc + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/urls", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ── PATCH /user/urls/{id} (update-shortened-url) ──────────────────────────────

func TestUpdateUrl_NoAuth_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil, nil)
	bl := routes.NewTokenBlacklist()

	body := bytes.NewBufferString(`{"long_url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateUrl_ValidRequest_Returns200(t *testing.T) {
	os.Setenv("APP_BASE_URL", "https://short.example.com")
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 20, UserName: "alice"}
	existing := &bizmodels.ShortenedUrl{ID: 5, UserID: 20, Shortcode: "old-sc", LongUrl: "https://old.com",  LastUpdated: now}
	updatedUrl := &bizmodels.ShortenedUrl{ID: 5, UserID: 20, Shortcode: "old-sc", LongUrl: "https://new.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().FindByID(gomock.Any(), int64(5)).Return(existing, nil)
			m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(updatedUrl, nil)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	bodyStr := `{"long_url":"https://new.com","last_updated":"` + now.UTC().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/5", bytes.NewBufferString(bodyStr))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "https://new.com", resp["long_url"])
	assert.Equal(t, "https://short.example.com/r/old-sc", resp["short_url"])
	assert.NotEmpty(t, resp["last_updated"])
}

func TestUpdateUrl_NotFound_Returns404(t *testing.T) {
	user := &bizmodels.User{ID: 21, UserName: "bob"}
	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().FindByID(gomock.Any(), int64(99)).Return(nil, businesslogic.ErrNotFound)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"https://example.com","last_updated":"2024-01-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/99", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUrl_WrongOwner_Returns403(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 22, UserName: "carol"}
	existing := &bizmodels.ShortenedUrl{ID: 6, UserID: 99, Shortcode: "abc123", LongUrl: "https://example.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().FindByID(gomock.Any(), int64(6)).Return(existing, nil)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"https://example.com","last_updated":"2024-01-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/6", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUrl_InvalidUrl_Returns400(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 23, UserName: "dave"}
	existing := &bizmodels.ShortenedUrl{ID: 7, UserID: 23, Shortcode: "abc123", LongUrl: "https://example.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().FindByID(gomock.Any(), int64(7)).Return(existing, nil)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"long_url":"ftp://bad-scheme.com","last_updated":"2024-01-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/7", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUrl_NilClaims_Returns401(t *testing.T) {
	h := newUrlHandlerForTest(t, nil, nil)

	body := bytes.NewBufferString(`{"long_url":"https://example.com","last_updated":"2024-01-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouterNoAuth(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateUrl_VersionConflict_Returns409(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	user := &bizmodels.User{ID: 24, UserName: "eve"}
	existing := &bizmodels.ShortenedUrl{ID: 8, UserID: 24, Shortcode: "abc123", LongUrl: "https://example.com",  LastUpdated: now}

	h := newUrlHandlerForTest(t,
		func(m *mocks.MockUrlRepository) {
			m.EXPECT().FindByID(gomock.Any(), int64(8)).Return(existing, nil)
			m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, businesslogic.ErrVersionConflict)
		},
		nil,
	)
	bl := routes.NewTokenBlacklist()
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)

	// Send a stale last_updated
	body := bytes.NewBufferString(`{"long_url":"https://example.com","last_updated":"2024-01-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPatch, "/user/urls/8", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newUrlTestRouter(h, bl).ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
