package mappings

import (
	"time"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
)

// ToBusinessModel converts a DB User to a business-layer User.
func UserToBusinessModel(db *pgmodels.User) *bizmodels.User {
	return &bizmodels.User{
		ID:       db.ID,
		UserName: db.UserName,
	}
}

// UserToDbModel converts a business-layer User to a DB User.
// now is injected by the caller so timestamp handling stays at the repository boundary.
func UserToDbModel(biz *bizmodels.User, now time.Time) *pgmodels.User {
	return &pgmodels.User{
		ID:        biz.ID,
		UserName:  biz.UserName,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
