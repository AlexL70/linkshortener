package models

// AuthInput carries the identity information returned by an OAuth2/OIDC provider
// after a successful authorization. It is the input to the auth business handler.
type AuthInput struct {
	Provider       Provider
	ProviderUserID string
	Email          string
	// DisplayName is a suggested username derived from the provider profile.
	// It may not be unique; the caller is responsible for uniqueness checks.
	DisplayName string
}
