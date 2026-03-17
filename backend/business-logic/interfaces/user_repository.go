package interfaces

import (
	"context"

	"github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// UserRepository defines the persistence operations required by the auth business handler.
type UserRepository interface {
	// FindByProviderID looks up the user and provider record matching the given
	// (provider, providerUserID) pair. Returns ErrNotFound if no match exists.
	FindByProviderID(ctx context.Context, provider models.Provider, providerUserID string) (*models.User, *models.UserProvider, error)

	// FindByProviderEmailWithSeedID looks up the user and provider record for an
	// admin user whose ProviderUserID is still set to the dev-seed placeholder value.
	// Returns ErrNotFound if no such record exists.
	FindByProviderEmailWithSeedID(ctx context.Context, provider models.Provider, email string) (*models.User, *models.UserProvider, error)

	// UpdateProviderUserID replaces the ProviderUserID on the given UserProvider record.
	UpdateProviderUserID(ctx context.Context, userProviderID int64, newProviderUserID string) error

	// CreateUserWithProvider creates a new User row and an associated UserProvider row
	// in a single transaction. Timestamps are managed by the repository.
	// Returns ErrConflict if the userName is already taken.
	CreateUserWithProvider(ctx context.Context, userName string, up *models.UserProvider) (*models.User, error)

	// FindProvidersByUserID returns all UserProvider records linked to the given user.
	// Returns ErrNotFound if the user has no associated providers (or does not exist).
	FindProvidersByUserID(ctx context.Context, userID int64) ([]*models.UserProvider, error)

	// DeleteUser removes the Users row with the given ID.
	// All child records are removed automatically via ON DELETE CASCADE.
	// Returns ErrNotFound if no such user exists.
	DeleteUser(ctx context.Context, userID int64) error
}
