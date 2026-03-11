package models

import (
	"time"

	"github.com/uptrace/bun"
)

type UrlClick struct {
	bun.BaseModel `bun:"table:UrlClicks,alias:uc"`

	ID        int64     `bun:"id,pk,autoincrement"`
	UrlID     int64     `bun:"url_id,notnull"`
	ClickedAt time.Time `bun:"clicked_at,notnull,default:current_timestamp"`
	IPAddress string    `bun:"ip_address"`
	UserAgent string    `bun:"user_agent"`
	Referer   string    `bun:"referer"`

	ShortenedUrl *ShortenedUrl `bun:"rel:belongs-to,join:url_id=id"`
}
