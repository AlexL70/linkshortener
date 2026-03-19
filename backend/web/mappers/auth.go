package mappers

import (
	"github.com/markbates/goth"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// GothUserToAuthInput converts a goth.User (returned by the OAuth callback) to
// the AuthInput type expected by the business logic handler.
func GothUserToAuthInput(gu goth.User, provider bizmodels.Provider) *bizmodels.AuthInput {
	return &bizmodels.AuthInput{
		Provider:       provider,
		ProviderUserID: gu.UserID,
		Email:          gu.Email,
		DisplayName:    gu.Name,
	}
}
