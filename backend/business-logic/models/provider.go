package models

// Provider enumerates the supported OAuth2/OIDC identity providers.
type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
	ProviderFacebook  Provider = "facebook"
)
