package routes

import (
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
)

// MapError translates a business logic sentinel error to the appropriate Huma HTTP error.
// All unmapped errors are logged at ERROR level and returned as 500.
func MapError(err error) error {
	switch {
	case errors.Is(err, businesslogic.ErrNotFound):
		return huma.Error404NotFound("resource not found")
	case errors.Is(err, businesslogic.ErrConflict):
		return huma.Error409Conflict("resource already exists")
	case errors.Is(err, businesslogic.ErrValidation):
		return huma.Error400BadRequest(err.Error())
	case errors.Is(err, businesslogic.ErrExpired):
		return huma.Error410Gone("resource has expired")
	case errors.Is(err, businesslogic.ErrUnauthorized):
		return huma.Error403Forbidden("access denied")
	case errors.Is(err, businesslogic.ErrNotImplemented):
		return huma.NewError(501, "not implemented")
	default:
		slog.Error("unhandled business error", "error", err)
		return huma.Error500InternalServerError("internal server error")
	}
}
