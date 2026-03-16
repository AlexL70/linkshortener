package routes

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"

	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/web/viewmodels"
)

// redirectResponse is the Huma output for GET /r/{shortcode}.
// Huma writes the Location header and uses DefaultStatus (302 Found).
// No body field — the response body is intentionally empty.
type redirectResponse struct {
	Location string `header:"Location"`
}

// RegisterRedirectRoute registers the public redirect endpoint.
//
// The Huma registration serves two purposes:
//  1. It creates the Gin route (via the humagin adapter) so the endpoint
//     actually handles HTTP requests.
//  2. It adds the operation to the OpenAPI spec.
//
// This endpoint is a permitted non-JSON route per backend-web.md §5.1 because
// it produces a 302 redirect rather than a JSON API response.
func RegisterRedirectRoute(_ *gin.Engine, api huma.API, h *handlers.UrlHandler) {
	huma.Register(api, huma.Operation{
		OperationID:   "redirect-shortcode",
		Method:        http.MethodGet,
		Path:          "/r/{shortcode}",
		Summary:       "Redirect to the long URL for a given shortcode",
		DefaultStatus: http.StatusFound,
	}, func(ctx context.Context, input *viewmodels.RedirectInput) (*redirectResponse, error) {
		url, err := h.ResolveShortcode(ctx, input.Shortcode)
		if err != nil {
			return nil, MapError(err)
		}
		slog.InfoContext(ctx, "redirect", "shortcode", input.Shortcode, "long_url", url.LongUrl, "url_id", url.ID)
		return &redirectResponse{Location: url.LongUrl}, nil
	})
}
