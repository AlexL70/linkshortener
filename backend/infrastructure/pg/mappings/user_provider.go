package mappings

import (
	"time"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
)

// UserProviderToBusinessModel converts a DB UserProvider to a business-layer UserProvider.
func UserProviderToBusinessModel(db *pgmodels.UserProvider) *bizmodels.UserProvider {
	return &bizmodels.UserProvider{
		ID:             db.ID,
		UserID:         db.UserID,
		Provider:       bizmodels.Provider(db.Provider),
		ProviderUserID: db.ProviderUserID,
		ProviderEmail:  db.ProviderEmail,
	}
}

// UserProviderToDbModel converts a business-layer UserProvider to a DB UserProvider.
// now is injected by the caller so timestamp handling stays at the repository boundary.
func UserProviderToDbModel(biz *bizmodels.UserProvider, now time.Time) *pgmodels.UserProvider {
	return &pgmodels.UserProvider{
		ID:             biz.ID,
		UserID:         biz.UserID,
		Provider:       string(biz.Provider),
		ProviderUserID: biz.ProviderUserID,
		ProviderEmail:  biz.ProviderEmail,
		CreatedAt:      now,
	}
}
