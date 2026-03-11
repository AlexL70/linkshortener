package models

import (
	"time"

	"github.com/uptrace/bun"
)

type UserProvider struct {
	bun.BaseModel `bun:"table:UserProviders,alias:up"`

	ID             int64     `bun:"id,pk,autoincrement"`
	UserID         int64     `bun:"user_id,notnull"`
	Provider       string    `bun:"provider,notnull"`
	ProviderUserID string    `bun:"provider_user_id,notnull"`
	ProviderEmail  string    `bun:"provider_email"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}
