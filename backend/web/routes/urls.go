package routes

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/danielgtaylor/huma/v2"

	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	webmappers "github.com/AlexL70/linkshortener/backend/web/mappers"
	"github.com/AlexL70/linkshortener/backend/web/viewmodels"
)

// RegisterUrlRoutes registers all URL management endpoints on the given Huma API.
func RegisterUrlRoutes(api huma.API, h *handlers.UrlHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-user-urls",
		Method:      http.MethodGet,
		Path:        "/user/urls",
		Summary:     "List authenticated user's shortened URLs",
	}, func(ctx context.Context, input *viewmodels.ListUrlsInput) (*viewmodels.ListUrlsResponse, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.NewError(http.StatusUnauthorized, "unauthorized")
		}

		pageSize := resolvePageSize(input.PageSize)

		urls, total, err := h.ListUrls(ctx, claims.UserID, input.Page, pageSize)
		if err != nil {
			return nil, MapError(err)
		}

		return &viewmodels.ListUrlsResponse{
			Body: webmappers.ListUrlsToResponse(urls, total, input.Page, pageSize),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-shortened-url",
		Method:        http.MethodPost,
		Path:          "/user/urls",
		Summary:       "Create a new shortened URL",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *viewmodels.CreateUrlInput) (*viewmodels.CreateUrlResponse, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.NewError(http.StatusUnauthorized, "unauthorized")
		}

		created, err := h.CreateUrl(ctx, claims.UserID, input.Body.LongUrl, input.Body.Shortcode, input.Body.ExpiresAt)
		if err != nil {
			return nil, MapError(err)
		}

		baseUrl := os.Getenv("APP_BASE_URL")
		return &viewmodels.CreateUrlResponse{
			Body: webmappers.CreateUrlToResponse(created, baseUrl),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-shortened-url",
		Method:      http.MethodPatch,
		Path:        "/user/urls/{id}",
		Summary:     "Update an existing shortened URL",
	}, func(ctx context.Context, input *viewmodels.UpdateUrlInput) (*viewmodels.UpdateUrlResponse, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.NewError(http.StatusUnauthorized, "unauthorized")
		}

		updated, err := h.UpdateUrl(ctx, input.ID, claims.UserID, input.Body.LongUrl, input.Body.Shortcode, input.Body.ExpiresAt, input.Body.LastUpdated)
		if err != nil {
			return nil, MapError(err)
		}

		baseUrl := os.Getenv("APP_BASE_URL")
		return &viewmodels.UpdateUrlResponse{
			Body: webmappers.UpdateUrlToResponse(updated, baseUrl),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-shortened-url",
		Method:        http.MethodDelete,
		Path:          "/user/urls/{id}",
		Summary:       "Delete an existing shortened URL",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, input *viewmodels.DeleteUrlInput) (*struct{}, error) {
		claims := GetJWTClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.NewError(http.StatusUnauthorized, "unauthorized")
		}

		if err := h.DeleteUrl(ctx, input.ID, claims.UserID, input.LastUpdated); err != nil {
			return nil, MapError(err)
		}
		return nil, nil
	})
}

// resolvePageSize returns the caller-supplied page size when positive, otherwise
// falls back to the DEFAULT_PAGE_SIZE environment variable (default: 20).
func resolvePageSize(requested int) int {
	if requested > 0 {
		return requested
	}
	s := os.Getenv("DEFAULT_PAGE_SIZE")
	if s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 20
}
