package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func TestHello_Returns200WithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := humagin.New(router, huma.DefaultConfig("Test API", "0.0.1"))
	routes.RegisterHello(api)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Hello World", body["message"])
}
