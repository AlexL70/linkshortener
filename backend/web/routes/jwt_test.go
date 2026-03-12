package routes_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func TestCreateJWT_RoundTrip(t *testing.T) {
	user := &bizmodels.User{ID: 42, UserName: "testuser"}
	token, err := routes.CreateJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCreateAndParsePreRegToken_RoundTrip(t *testing.T) {
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "google-sub-test",
		Email:          "test@example.com",
		DisplayName:    "Test User",
	}

	token, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsed, err := routes.ParsePreRegToken(token)
	require.NoError(t, err)
	assert.Equal(t, input.Provider, parsed.Provider)
	assert.Equal(t, input.ProviderUserID, parsed.ProviderUserID)
	assert.Equal(t, input.Email, parsed.Email)
	assert.Equal(t, input.DisplayName, parsed.DisplayName)
}

func TestParsePreRegToken_InvalidToken(t *testing.T) {
	_, err := routes.ParsePreRegToken("not.a.valid.jwt")
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestParsePreRegToken_WrongSecret(t *testing.T) {
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "sub",
		Email:          "a@b.com",
	}
	token, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)

	// Switch secret before parsing.
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "completely-different-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	_, err = routes.ParsePreRegToken(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreatePreRegToken_ValidRoundTrip(t *testing.T) {
	// Verify a freshly-issued pre-reg token can be parsed without error,
	// confirming that the expiry is set in the future (not already expired).
	input := &bizmodels.AuthInput{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "sub-expiry",
		Email:          "exp@example.com",
	}
	token, err := routes.CreatePreRegToken(input)
	require.NoError(t, err)

	parsed, err := routes.ParsePreRegToken(token)
	require.NoError(t, err)
	assert.Equal(t, input.Email, parsed.Email)
}
