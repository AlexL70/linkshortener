package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ShortenedUrl struct {
	bun.BaseModel `bun:"table:ShortenedUrls,alias:su"`

	ID        int64      `bun:"id,pk,autoincrement"`
	UserID    int64      `bun:"user_id,notnull"`
	Shortcode string     `bun:"shortcode,notnull,unique"`
	LongUrl   string     `bun:"long_url,notnull"`
	ExpiresAt *time.Time `bun:"expires_at"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}
