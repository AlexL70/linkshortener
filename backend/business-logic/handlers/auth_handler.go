package handlers

import (
	"context"
	"fmt"
	"log/slog"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	"github.com/AlexL70/linkshortener/backend/business-logic/models"
)

// AuthHandler contains the business logic for authenticating and registering users.
type AuthHandler struct {
	users      interfaces.UserRepository
	isDevMode  bool
	adminEmail string
}

// NewAuthHandler constructs an AuthHandler.
// adminEmail must be the value of the SUPER_ADMIN_EMAIL environment variable;
// it is only used in dev mode to apply the seed-update logic.
func NewAuthHandler(users interfaces.UserRepository, isDevMode bool, adminEmail string) *AuthHandler {
	return &AuthHandler{users: users, isDevMode: isDevMode, adminEmail: adminEmail}
}

// ResolveUserByProvider resolves an OAuth callback to an existing user.
//
// If the (provider, providerUserID) pair is already in the database, the
// corresponding User is returned.
//
// Dev-mode only: if the authenticated email matches the admin email and the
// stored ProviderUserID is still the dev-seed placeholder, the record is
// updated with the real provider-issued subject ID before returning.
//
// If no matching user is found, ErrNewUser is returned so the caller can
// redirect to the registration flow.
func (h *AuthHandler) ResolveUserByProvider(ctx context.Context, input *models.AuthInput) (*models.User, error) {
	// Fast path: look up by the exact (provider, providerUserID) pair.
	user, _, err := h.users.FindByProviderID(ctx, input.Provider, input.ProviderUserID)
	if err == nil {
		return user, nil
	}
	if err != businesslogic.ErrNotFound {
		return nil, fmt.Errorf("AuthHandler.ResolveUserByProvider: %w", err)
	}

	// Dev-mode seed-update path: only when the authenticated user is the admin.
	if h.isDevMode && input.Email == h.adminEmail {
		seedUser, seedUP, findErr := h.users.FindByProviderEmailWithSeedID(ctx, input.Provider, input.Email)
		if findErr == nil {
			// Replace the placeholder ProviderUserID with the real one.
			if updateErr := h.users.UpdateProviderUserID(ctx, seedUP.ID, input.ProviderUserID); updateErr != nil {
				return nil, fmt.Errorf("AuthHandler.ResolveUserByProvider: update seed provider id: %w", updateErr)
			}
			slog.Info("auth: updated seed ProviderUserID for admin user",
				"user_id", seedUser.ID,
				"user_name", seedUser.UserName,
				"provider", input.Provider,
			)
			return seedUser, nil
		}
		if findErr != businesslogic.ErrNotFound {
			return nil, fmt.Errorf("AuthHandler.ResolveUserByProvider: find seed user: %w", findErr)
		}
		// Seed record not found — fall through to ErrNewUser so the admin can register normally.
	}

	return nil, businesslogic.ErrNewUser
}

// CreateUser registers a new user with the given userName and binds the OAuth
// provider identity supplied in input.
//
// Returns ErrConflict if the userName is already taken.
func (h *AuthHandler) CreateUser(ctx context.Context, userName string, input *models.AuthInput) (*models.User, error) {
	up := &models.UserProvider{
		Provider:       input.Provider,
		ProviderUserID: input.ProviderUserID,
		ProviderEmail:  input.Email,
	}

	user, err := h.users.CreateUserWithProvider(ctx, userName, up)
	if err != nil {
		return nil, fmt.Errorf("AuthHandler.CreateUser: %w", err)
	}

	slog.Info("auth: registered new user",
		"user_id", user.ID,
		"user_name", user.UserName,
		"provider", input.Provider,
	)
	return user, nil
}
