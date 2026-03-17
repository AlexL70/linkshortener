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
	token, err := routes.CreateJWT(user, "testuser@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCreateJWT_JTIPresent(t *testing.T) {
	// Every full JWT must carry a non-empty jti claim so it can be blacklisted.
	user := &bizmodels.User{ID: 99, UserName: "jtiuser"}
	token, err := routes.CreateJWT(user, "jtiuser@example.com")
	require.NoError(t, err)

	claims, err := routes.ParseJWT(token)
	require.NoError(t, err)
	assert.NotEmpty(t, claims.ID, "jti (RegisteredClaims.ID) must be non-empty")
}

func TestCreateJWT_TwoTokensHaveDifferentJTI(t *testing.T) {
	user := &bizmodels.User{ID: 7, UserName: "uniquejti"}
	t1, err := routes.CreateJWT(user, "uniquejti@example.com")
	require.NoError(t, err)
	t2, err := routes.CreateJWT(user, "uniquejti@example.com")
	require.NoError(t, err)

	c1, _ := routes.ParseJWT(t1)
	c2, _ := routes.ParseJWT(t2)
	assert.NotEqual(t, c1.ID, c2.ID, "each issued JWT must have a unique jti")
}

func TestParseJWT_RoundTrip(t *testing.T) {
	user := &bizmodels.User{ID: 5, UserName: "parsetest"}
	tokenStr, err := routes.CreateJWT(user, "parsetest@example.com")
	require.NoError(t, err)

	claims, err := routes.ParseJWT(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, int64(5), claims.UserID)
	assert.Equal(t, "parsetest", claims.UserName)
	assert.Equal(t, "parsetest@example.com", claims.Email)
	assert.Equal(t, "5", claims.Subject)
	assert.NotEmpty(t, claims.ID)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := routes.ParseJWT("not.a.valid.jwt")
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestParseJWT_WrongSecret(t *testing.T) {
	user := &bizmodels.User{ID: 3, UserName: "secrettest"}
	tokenStr, err := routes.CreateJWT(user, "secrettest@example.com")
	require.NoError(t, err)

	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "completely-different-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	_, err = routes.ParseJWT(tokenStr)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
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
