package repositories

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmappings "github.com/AlexL70/linkshortener/backend/infrastructure/pg/mappings"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
)

type urlRepository struct {
	db *bun.DB
}

// NewUrlRepository constructs a UrlRepository backed by the given *bun.DB.
func NewUrlRepository(db *bun.DB) interfaces.UrlRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*bizmodels.ShortenedUrl, int, error) {
	total, err := r.db.NewSelect().
		Model((*pgmodels.ShortenedUrl)(nil)).
		Where("su.user_id = ?", userID).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("UrlRepository.FindByUserID count: %w", err)
	}

	if total == 0 {
		return []*bizmodels.ShortenedUrl{}, 0, nil
	}

	var dbs []*pgmodels.ShortenedUrl
	offset := (page - 1) * pageSize
	err = r.db.NewSelect().
		Model(&dbs).
		Column("su.id", "su.user_id", "su.shortcode", "su.long_url", "su.expires_at", "su.created_at", "su.updated_at").
		Where("su.user_id = ?", userID).
		OrderExpr("su.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("UrlRepository.FindByUserID: %w", err)
	}

	return pgmappings.ShortenedUrlSliceToBusinessModel(dbs), total, nil
}
