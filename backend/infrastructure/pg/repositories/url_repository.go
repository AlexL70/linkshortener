package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
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

func (r *urlRepository) Create(ctx context.Context, url *bizmodels.ShortenedUrl) (*bizmodels.ShortenedUrl, error) {
	db := pgmappings.ShortenedUrlToDbModel(url)
	now := time.Now()
	db.CreatedAt = now
	db.UpdatedAt = now

	if _, err := r.db.NewInsert().Model(db).Returning("*").Exec(ctx); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("UrlRepository.Create: %w", businesslogic.ErrConflict)
		}
		return nil, fmt.Errorf("UrlRepository.Create: %w", err)
	}

	return pgmappings.ShortenedUrlToBusinessModel(db), nil
}

func (r *urlRepository) FindByID(ctx context.Context, id int64) (*bizmodels.ShortenedUrl, error) {
	db := new(pgmodels.ShortenedUrl)
	err := r.db.NewSelect().
		Model(db).
		Where("su.id = ?", id).
		Scan(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("UrlRepository.FindByID: %w", businesslogic.ErrNotFound)
		}
		return nil, fmt.Errorf("UrlRepository.FindByID: %w", err)
	}
	return pgmappings.ShortenedUrlToBusinessModel(db), nil
}

func (r *urlRepository) Update(ctx context.Context, url *bizmodels.ShortenedUrl) (*bizmodels.ShortenedUrl, error) {
	db := pgmappings.ShortenedUrlToDbModel(url)
	db.UpdatedAt = time.Now()

	if _, err := r.db.NewUpdate().
		Model(db).
		Column("shortcode", "long_url", "expires_at", "updated_at").
		Where("id = ?", db.ID).
		Returning("*").
		Exec(ctx); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("UrlRepository.Update: %w", businesslogic.ErrConflict)
		}
		return nil, fmt.Errorf("UrlRepository.Update: %w", err)
	}

	return pgmappings.ShortenedUrlToBusinessModel(db), nil
}
