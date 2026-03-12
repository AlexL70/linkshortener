package routes

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

const preRegTokenDuration = 10 * time.Minute

// preRegClaims are the JWT claims for a short-lived pre-registration token.
type preRegClaims struct {
	Provider       bizmodels.Provider `json:"provider"`
	ProviderUserID string             `json:"provider_user_id"`
	Email          string             `json:"email"`
	DisplayName    string             `json:"display_name"`
	jwt.RegisteredClaims
}

// CreateJWT creates a signed JWT for the given authenticated user.
// Claims include user_id, user_name, sub, iat, and exp (24-hour expiry).
func CreateJWT(user *bizmodels.User) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"user_name": user.UserName,
		"sub":       fmt.Sprintf("%d", user.ID),
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("CreateJWT: %w", err)
	}
	return signed, nil
}

// CreatePreRegToken creates a short-lived JWT encoding the provider identity returned
// by the OAuth callback for a new (not yet registered) user. Expires in 10 minutes.
func CreatePreRegToken(input *bizmodels.AuthInput) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	claims := preRegClaims{
		Provider:       input.Provider,
		ProviderUserID: input.ProviderUserID,
		Email:          input.Email,
		DisplayName:    input.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(preRegTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("CreatePreRegToken: %w", err)
	}
	return signed, nil
}

// ParsePreRegToken validates and parses a pre-registration JWT.
// Returns ErrValidation if the token is invalid or expired.
func ParsePreRegToken(tokenStr string) (*bizmodels.AuthInput, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenStr, &preRegClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", businesslogic.ErrValidation, err.Error())
	}

	claims, ok := token.Claims.(*preRegClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: invalid token claims", businesslogic.ErrValidation)
	}

	return &bizmodels.AuthInput{
		Provider:       claims.Provider,
		ProviderUserID: claims.ProviderUserID,
		Email:          claims.Email,
		DisplayName:    claims.DisplayName,
	}, nil
}
