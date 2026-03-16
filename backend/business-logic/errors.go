package businesslogic

import "errors"

// Sentinel errors used by business logic handlers.
// The web layer maps these to appropriate HTTP responses.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned when a creation or update would violate a uniqueness constraint.
	ErrConflict = errors.New("conflict")

	// ErrValidation is returned when input fails domain validation rules.
	ErrValidation = errors.New("validation error")

	// ErrExpired is returned when a resource (e.g. a shortened URL) has passed its expiry date.
	ErrExpired = errors.New("expired")

	// ErrUnauthorized is returned when the caller lacks permission for the requested operation.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrNewUser is returned by ResolveUserByProvider when the provider identity is not
	// yet associated with any user account. The caller should redirect to the registration flow.
	ErrNewUser = errors.New("new user")

	// ErrNotImplemented is returned when a feature or provider is not yet supported.
	ErrNotImplemented = errors.New("not implemented")

	// ErrVersionConflict is returned when an entity was modified by another user
	// since it was last read. The caller should ask the user to refresh and retry.
	ErrVersionConflict = errors.New("version conflict")
)
