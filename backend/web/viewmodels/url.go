package viewmodels

import "time"

// ListUrlsInput is the Huma input for GET /user/urls.
// Page defaults to 1. PageSize defaults to 0 (server-side default applies).
type ListUrlsInput struct {
	Page     int `query:"page" minimum:"1" default:"1"`
	PageSize int `query:"page_size" maximum:"100"`
}

// ListUrlsResponse is the Huma output for GET /user/urls.
type ListUrlsResponse struct {
	Body *ListUrlsResponseBody
}

// ListUrlsResponseBody is the JSON payload returned by GET /user/urls.
type ListUrlsResponseBody struct {
	Items    []*UrlItem `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// UrlItem represents a single shortened URL in the list response.
type UrlItem struct {
	ID          int64      `json:"id"`
	Shortcode   string     `json:"shortcode"`
	LongUrl     string     `json:"long_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUpdated time.Time  `json:"last_updated"`
}

// CreateUrlInput is the Huma input for POST /user/urls.
type CreateUrlInput struct {
	Body *CreateUrlRequestBody
}

// CreateUrlRequestBody is the JSON request body for creating a new shortened URL.
type CreateUrlRequestBody struct {
	LongUrl   string     `json:"long_url" required:"true" minLength:"1"`
	Shortcode *string    `json:"shortcode,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreateUrlResponse is the Huma output for POST /user/urls (201 Created).
type CreateUrlResponse struct {
	Body *CreateUrlResponseBody
}

// CreateUrlResponseBody is the JSON payload returned when a URL is successfully created.
type CreateUrlResponseBody struct {
	ID        int64      `json:"id"`
	Shortcode string     `json:"shortcode"`
	LongUrl   string     `json:"long_url"`
	ShortUrl  string     `json:"short_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// UpdateUrlInput is the Huma input for PATCH /user/urls/{id}.
type UpdateUrlInput struct {
	ID   int64 `path:"id"`
	Body *UpdateUrlRequestBody
}

// UpdateUrlRequestBody is the JSON request body for updating an existing shortened URL.
// LongUrl is required. Shortcode, when provided, replaces the current one.
// ExpiresAt, when null, clears the current expiry; when provided, sets a new one.
// LastUpdated is required for optimistic concurrency control.
type UpdateUrlRequestBody struct {
	LongUrl     string     `json:"long_url" required:"true" minLength:"1"`
	Shortcode   *string    `json:"shortcode,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUpdated time.Time  `json:"last_updated" validate:"required"`
}

// UpdateUrlResponse is the Huma output for PATCH /user/urls/{id}.
type UpdateUrlResponse struct {
	Body *UpdateUrlResponseBody
}

// UpdateUrlResponseBody is the JSON payload returned when a URL is successfully updated.
type UpdateUrlResponseBody struct {
	ID          int64      `json:"id"`
	Shortcode   string     `json:"shortcode"`
	LongUrl     string     `json:"long_url"`
	ShortUrl    string     `json:"short_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUpdated time.Time  `json:"last_updated"`
}
