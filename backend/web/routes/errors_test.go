package routes_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func assertHTTPStatus(t *testing.T, err error, wantStatus int) {
	t.Helper()
	require.Error(t, err)
	se, ok := err.(huma.StatusError)
	require.True(t, ok, "expected a StatusError, got %T", err)
	assert.Equal(t, wantStatus, se.GetStatus())
}

func TestMapError_KnownSentinels(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"ErrNotFound", businesslogic.ErrNotFound, http.StatusNotFound},
		{"ErrConflict", businesslogic.ErrConflict, http.StatusConflict},
		{"ErrValidation", businesslogic.ErrValidation, http.StatusBadRequest},
		{"ErrExpired", businesslogic.ErrExpired, http.StatusGone},
		{"ErrUnauthorized", businesslogic.ErrUnauthorized, http.StatusForbidden},
		{"ErrNotImplemented", businesslogic.ErrNotImplemented, http.StatusNotImplemented},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertHTTPStatus(t, routes.MapError(tc.err), tc.wantStatus)
		})
	}
}

func TestMapError_WrappedSentinel(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", businesslogic.ErrNotFound)
	assertHTTPStatus(t, routes.MapError(wrapped), http.StatusNotFound)
}

func TestMapError_UnknownError(t *testing.T) {
	assertHTTPStatus(t, routes.MapError(fmt.Errorf("unexpected internal error")), http.StatusInternalServerError)
}
