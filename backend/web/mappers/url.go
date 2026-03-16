package mappers

import (
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/viewmodels"
)

// UrlToViewModel converts a business-layer ShortenedUrl to a UrlItem viewmodel.
func UrlToViewModel(m *bizmodels.ShortenedUrl) *viewmodels.UrlItem {
	return &viewmodels.UrlItem{
		ID:          m.ID,
		Shortcode:   m.Shortcode,
		LongUrl:     m.LongUrl,
		ExpiresAt:   m.ExpiresAt,
		LastUpdated: m.LastUpdated,
	}
}

// ListUrlsToResponse builds the paginated list response body from a slice of business models.
func ListUrlsToResponse(urls []*bizmodels.ShortenedUrl, total, page, pageSize int) *viewmodels.ListUrlsResponseBody {
	items := make([]*viewmodels.UrlItem, len(urls))
	for i, u := range urls {
		items[i] = UrlToViewModel(u)
	}
	return &viewmodels.ListUrlsResponseBody{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// CreateUrlToResponse builds a CreateUrlResponseBody from a created ShortenedUrl.
// shortUrl is constructed as baseUrl + "/r/" + shortcode.
func CreateUrlToResponse(m *bizmodels.ShortenedUrl, baseUrl string) *viewmodels.CreateUrlResponseBody {
	return &viewmodels.CreateUrlResponseBody{
		ID:        m.ID,
		Shortcode: m.Shortcode,
		LongUrl:   m.LongUrl,
		ShortUrl:  baseUrl + "/r/" + m.Shortcode,
		ExpiresAt: m.ExpiresAt,
	}
}

// UpdateUrlToResponse builds an UpdateUrlResponseBody from an updated ShortenedUrl.
// shortUrl is constructed as baseUrl + "/r/" + shortcode.
func UpdateUrlToResponse(m *bizmodels.ShortenedUrl, baseUrl string) *viewmodels.UpdateUrlResponseBody {
	return &viewmodels.UpdateUrlResponseBody{
		ID:          m.ID,
		Shortcode:   m.Shortcode,
		LongUrl:     m.LongUrl,
		ShortUrl:    baseUrl + "/r/" + m.Shortcode,
		ExpiresAt:   m.ExpiresAt,
		LastUpdated: m.LastUpdated,
	}
}
