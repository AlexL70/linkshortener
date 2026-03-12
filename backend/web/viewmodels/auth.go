package viewmodels

// RegisterRequestBody is the payload for POST /auth/register.
type RegisterRequestBody struct {
	// PreRegistrationToken is the short-lived JWT issued during the OAuth callback
	// when a new user identity is detected. It encodes the provider info.
	PreRegistrationToken string `json:"pre_registration_token" validate:"required"`
	// UserName is the desired username chosen by the user during registration.
	UserName string `json:"user_name" validate:"required" minLength:"3" maxLength:"50"`
}

// RegisterRequest is the Huma input wrapper for POST /auth/register.
type RegisterRequest struct {
	Body *RegisterRequestBody
}

// AuthTokenBody is the response body returned after a successful login or registration.
type AuthTokenBody struct {
	Token string `json:"token"`
}

// AuthTokenResponse is the Huma output wrapper for a successful auth response.
type AuthTokenResponse struct {
	Body *AuthTokenBody
}

// PreRegistrationTokenBody is returned from the OAuth callback when a new user
// identity is detected. The frontend should present a registration form, collect
// a username, and POST both fields to /auth/register.
type PreRegistrationTokenBody struct {
	PreRegistrationToken string `json:"pre_registration_token"`
	SuggestedUserName    string `json:"suggested_user_name"`
}
