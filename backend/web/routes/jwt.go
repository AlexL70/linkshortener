package routes

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

const preRegTokenDuration = 10 * time.Minute

// JWTClaims holds the claims embedded in every full session JWT.
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// preRegClaims are the JWT claims for a short-lived pre-registration token.
type preRegClaims struct {
	Provider       bizmodels.Provider `json:"provider"`
	ProviderUserID string             `json:"provider_user_id"`
	Email          string             `json:"email"`
	DisplayName    string             `json:"display_name"`
	jwt.RegisteredClaims
}

// generateJTI generates a cryptographically random 16-byte hex string to use as
// a JWT ID (jti). This allows individual tokens to be blacklisted on logout.
func generateJTI() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generateJTI: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// CreateJWT creates a signed JWT for the given authenticated user.
// Claims include user_id, user_name, email, sub, jti, iat, and exp (24-hour expiry).
func CreateJWT(user *bizmodels.User, email string) (string, error) {
	jti, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("CreateJWT: %w", err)
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		UserName: user.UserName,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("CreateJWT: %w", err)
	}
	return signed, nil
}

// ParseJWT validates and parses a full session JWT.
// Returns ErrValidation if the token is invalid, expired, or uses an unexpected signing method.
func ParseJWT(tokenStr string) (*JWTClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", businesslogic.ErrValidation, err.Error())
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: invalid token claims", businesslogic.ErrValidation)
	}

	return claims, nil
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
