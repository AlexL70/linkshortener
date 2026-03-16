package models

import "time"

// ShortenedUrl represents a shortened URL at the business layer.
type ShortenedUrl struct {
	ID          int64
	UserID      int64
	Shortcode   string
	LongUrl     string
	ExpiresAt   *time.Time
	LastUpdated time.Time
}
