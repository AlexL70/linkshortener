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
	ID        int64      `json:"id"`
	Shortcode string     `json:"shortcode"`
	LongUrl   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
