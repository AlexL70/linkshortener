package interfaces

import (
	"context"
	"time"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// UrlRepository is the data access contract for shortened URL operations.
type UrlRepository interface {
	// FindByUserID returns a paginated list of shortened URLs owned by userID,
	// along with the total count for pagination metadata.
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*bizmodels.ShortenedUrl, int, error)

	// FindByID returns the shortened URL with the given ID, or ErrNotFound if it does not exist.
	FindByID(ctx context.Context, id int64) (*bizmodels.ShortenedUrl, error)

	// Create persists a new shortened URL and returns the stored record with its generated ID and timestamps.
	Create(ctx context.Context, url *bizmodels.ShortenedUrl) (*bizmodels.ShortenedUrl, error)

	// Update persists changes to an existing shortened URL and returns the updated record.
	Update(ctx context.Context, url *bizmodels.ShortenedUrl) (*bizmodels.ShortenedUrl, error)

	// Delete removes the shortened URL with the given ID when it is owned by userID and
	// its updated_at timestamp matches lastUpdated (optimistic concurrency check).
	// Returns ErrNotFound if the record does not exist and ErrVersionConflict if it was
	// modified since lastUpdated.
	Delete(ctx context.Context, id, userID int64, lastUpdated time.Time) error
}
