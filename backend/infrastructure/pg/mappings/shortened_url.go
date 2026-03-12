package mappings

import (
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
)

// ShortenedUrlToBusinessModel converts a DB ShortenedUrl to a business-layer ShortenedUrl.
func ShortenedUrlToBusinessModel(db *pgmodels.ShortenedUrl) *bizmodels.ShortenedUrl {
	return &bizmodels.ShortenedUrl{
		ID:        db.ID,
		UserID:    db.UserID,
		Shortcode: db.Shortcode,
		LongUrl:   db.LongUrl,
		ExpiresAt: db.ExpiresAt,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}

// ShortenedUrlToDbModel converts a business-layer ShortenedUrl to a DB ShortenedUrl.
func ShortenedUrlToDbModel(biz *bizmodels.ShortenedUrl) *pgmodels.ShortenedUrl {
	return &pgmodels.ShortenedUrl{
		ID:        biz.ID,
		UserID:    biz.UserID,
		Shortcode: biz.Shortcode,
		LongUrl:   biz.LongUrl,
		ExpiresAt: biz.ExpiresAt,
		CreatedAt: biz.CreatedAt,
		UpdatedAt: biz.UpdatedAt,
	}
}

// ShortenedUrlSliceToBusinessModel converts a slice of DB ShortenedUrl models to business-layer models.
func ShortenedUrlSliceToBusinessModel(dbs []*pgmodels.ShortenedUrl) []*bizmodels.ShortenedUrl {
	result := make([]*bizmodels.ShortenedUrl, len(dbs))
	for i, db := range dbs {
		result[i] = ShortenedUrlToBusinessModel(db)
	}
	return result
}
