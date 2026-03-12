package handlers

import (
	"context"
	"fmt"

	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// UrlHandler contains the business logic for managing shortened URLs.
type UrlHandler struct {
	urls interfaces.UrlRepository
}

// NewUrlHandler constructs a UrlHandler.
func NewUrlHandler(urls interfaces.UrlRepository) *UrlHandler {
	return &UrlHandler{urls: urls}
}

// ListUrls returns a paginated list of shortened URLs owned by userID.
// Returns the URL slice, the total count of matching records, and any error.
func (h *UrlHandler) ListUrls(ctx context.Context, userID int64, page, pageSize int) ([]*bizmodels.ShortenedUrl, int, error) {
	urls, total, err := h.urls.FindByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("UrlHandler.ListUrls: %w", err)
	}
	return urls, total, nil
}
