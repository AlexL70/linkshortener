package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

const maxShortcodeRetries = 10

// UrlHandler contains the business logic for managing shortened URLs.
type UrlHandler struct {
	urls            interfaces.UrlRepository
	generator       interfaces.ShortcodeGenerator
	maxUrlLen       int
	minShortcodeLen int
	maxShortcodeLen int
}

// NewUrlHandler constructs a UrlHandler.
// maxUrlLen, minShortcodeLen, and maxShortcodeLen are read from configuration
// (MAX_URL_LENGTH, MIN_SHORTCODE_LENGTH, MAX_SHORTCODE_LENGTH) and applied during validation.
func NewUrlHandler(urls interfaces.UrlRepository, generator interfaces.ShortcodeGenerator, maxUrlLen, minShortcodeLen, maxShortcodeLen int) *UrlHandler {
	return &UrlHandler{
		urls:            urls,
		generator:       generator,
		maxUrlLen:       maxUrlLen,
		minShortcodeLen: minShortcodeLen,
		maxShortcodeLen: maxShortcodeLen,
	}
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

// CreateUrl creates a new shortened URL for the given user.
// If customShortcode is nil, a random Base62 shortcode is generated with up to
// maxShortcodeRetries retries on collision.
// Long URL validation and, when provided, custom shortcode validation are applied
// before any DB operation.
func (h *UrlHandler) CreateUrl(ctx context.Context, userID int64, longUrl string, customShortcode *string, expiresAt *time.Time) (*bizmodels.ShortenedUrl, error) {
	if err := businesslogic.ValidateLongUrl(longUrl, h.maxUrlLen); err != nil {
		return nil, err
	}

	if customShortcode != nil {
		if err := businesslogic.ValidateCustomShortcode(*customShortcode, h.minShortcodeLen, h.maxShortcodeLen); err != nil {
			return nil, err
		}
		record := &bizmodels.ShortenedUrl{
			UserID:    userID,
			Shortcode: *customShortcode,
			LongUrl:   longUrl,
			ExpiresAt: expiresAt,
		}
		created, err := h.urls.Create(ctx, record)
		if err != nil {
			return nil, fmt.Errorf("UrlHandler.CreateUrl: %w", err)
		}
		return created, nil
	}

	for attempt := 0; attempt < maxShortcodeRetries; attempt++ {
		sc, err := h.generator.GenerateShortcode()
		if err != nil {
			return nil, fmt.Errorf("UrlHandler.CreateUrl: shortcode generation failed: %w", err)
		}
		record := &bizmodels.ShortenedUrl{
			UserID:    userID,
			Shortcode: sc,
			LongUrl:   longUrl,
			ExpiresAt: expiresAt,
		}
		created, err := h.urls.Create(ctx, record)
		if err == nil {
			return created, nil
		}
		if !errors.Is(err, businesslogic.ErrConflict) {
			return nil, fmt.Errorf("UrlHandler.CreateUrl: %w", err)
		}
		slog.WarnContext(ctx, "shortcode collision, retrying", "attempt", attempt+1, "user_id", userID)
	}

	return nil, fmt.Errorf("UrlHandler.CreateUrl: %w: exhausted %d shortcode retries", businesslogic.ErrConflict, maxShortcodeRetries)
}
