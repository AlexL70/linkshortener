package interfaces

import (
	"context"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// UrlRepository is the data access contract for shortened URL operations.
type UrlRepository interface {
	// FindByUserID returns a paginated list of shortened URLs owned by userID,
	// along with the total count for pagination metadata.
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*bizmodels.ShortenedUrl, int, error)
}
