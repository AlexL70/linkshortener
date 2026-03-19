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

// RegisterResponse is the Huma output for POST /auth/register.
// The JWT is delivered as an HttpOnly session cookie via the Set-Cookie header;
// no token is exposed in the response body.
type RegisterResponse struct {
	SetCookie string `header:"Set-Cookie"`
}

// LogoutResponse is the Huma output for POST /auth/logout.
// It carries a Set-Cookie header that clears the session cookie.
type LogoutResponse struct {
	SetCookie string `header:"Set-Cookie"`
}

// MeBody is the response body for GET /auth/me.
type MeBody struct {
	UserID        int64  `json:"user_id"`
	UserName      string `json:"user_name"`
	ProviderEmail string `json:"provider_email"`
}

// MeResponse is the Huma output wrapper for GET /auth/me.
type MeResponse struct {
	Body *MeBody
}

// PreRegistrationTokenBody is returned from the OAuth callback when a new user
// identity is detected. The frontend should present a registration form, collect
// a username, and POST both fields to /auth/register.
type PreRegistrationTokenBody struct {
	PreRegistrationToken string `json:"pre_registration_token"`
	SuggestedUserName    string `json:"suggested_user_name"`
}
