package models

// DevSeedProviderUserID is the placeholder ProviderUserID written by the database
// seeder in dev mode. When a seeded admin user completes a real OAuth login for the
// first time, this value is replaced with the actual provider-issued subject ID.
const DevSeedProviderUserID = "dev-seed"

// UserProvider holds the link between a User and an external OAuth2/OIDC provider.
type UserProvider struct {
	ID             int64
	UserID         int64
	Provider       Provider
	ProviderUserID string
	ProviderEmail  string
}
