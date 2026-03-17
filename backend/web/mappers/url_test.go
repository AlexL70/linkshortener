package mappers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/mappers"
)

func TestUrlToViewModel_AllFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	exp := now.Add(24 * time.Hour)
	m := &bizmodels.ShortenedUrl{
		ID:          42,
		UserID:      7,
		Shortcode:   "abc123",
		LongUrl:     "https://example.com",
		ExpiresAt:   &exp,
		LastUpdated: now,
	}

	vm := mappers.UrlToViewModel(m, "https://s.example.com")

	require.NotNil(t, vm)
	assert.Equal(t, int64(42), vm.ID)
	assert.Equal(t, "abc123", vm.Shortcode)
	assert.Equal(t, "https://example.com", vm.LongUrl)
	assert.Equal(t, "https://s.example.com/r/abc123", vm.ShortUrl)
	require.NotNil(t, vm.ExpiresAt)
	assert.Equal(t, exp, *vm.ExpiresAt)
	assert.Equal(t, now, vm.LastUpdated)
}

func TestUrlToViewModel_NoExpiry(t *testing.T) {
	now := time.Now()
	m := &bizmodels.ShortenedUrl{
		ID:          1,
		Shortcode:   "xyz789",
		LongUrl:     "https://no-expiry.com",
		LastUpdated: now,
	}

	vm := mappers.UrlToViewModel(m, "https://s.example.com")

	require.NotNil(t, vm)
	assert.Equal(t, "https://s.example.com/r/xyz789", vm.ShortUrl)
	assert.Nil(t, vm.ExpiresAt)
}

func TestListUrlsToResponse_MultipleItems(t *testing.T) {
	now := time.Now()
	urls := []*bizmodels.ShortenedUrl{
		{ID: 1, Shortcode: "aaa111", LongUrl: "https://a.com", LastUpdated: now},
		{ID: 2, Shortcode: "bbb222", LongUrl: "https://b.com", LastUpdated: now},
	}

	body := mappers.ListUrlsToResponse(urls, 10, 2, 5, "https://s.example.com")

	require.NotNil(t, body)
	assert.Equal(t, 10, body.Total)
	assert.Equal(t, 2, body.Page)
	assert.Equal(t, 5, body.PageSize)
	assert.Len(t, body.Items, 2)
	assert.Equal(t, "aaa111", body.Items[0].Shortcode)
	assert.Equal(t, "https://s.example.com/r/aaa111", body.Items[0].ShortUrl)
	assert.Equal(t, "bbb222", body.Items[1].Shortcode)
	assert.Equal(t, "https://s.example.com/r/bbb222", body.Items[1].ShortUrl)
}

func TestListUrlsToResponse_EmptySlice(t *testing.T) {
	body := mappers.ListUrlsToResponse([]*bizmodels.ShortenedUrl{}, 0, 1, 20, "")

	require.NotNil(t, body)
	assert.Equal(t, 0, body.Total)
	assert.Empty(t, body.Items)
}

// ── CreateUrlToResponse ───────────────────────────────────────────────────────

func TestCreateUrlToResponse_AllFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	exp := now.Add(48 * time.Hour)
	m := &bizmodels.ShortenedUrl{
		ID:          99,
		UserID:      5,
		Shortcode:   "cr0001",
		LongUrl:     "https://long.example.com/path",
		ExpiresAt:   &exp,
		LastUpdated: now,
	}

	body := mappers.CreateUrlToResponse(m, "https://short.example.com")

	require.NotNil(t, body)
	assert.Equal(t, int64(99), body.ID)
	assert.Equal(t, "cr0001", body.Shortcode)
	assert.Equal(t, "https://long.example.com/path", body.LongUrl)
	assert.Equal(t, "https://short.example.com/r/cr0001", body.ShortUrl)
	require.NotNil(t, body.ExpiresAt)
	assert.Equal(t, exp, *body.ExpiresAt)
}

func TestCreateUrlToResponse_NoExpiry(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	m := &bizmodels.ShortenedUrl{
		ID:          5,
		Shortcode:   "noexp1",
		LongUrl:     "https://example.com",
		LastUpdated: now,
	}

	body := mappers.CreateUrlToResponse(m, "https://s.example.com")

	require.NotNil(t, body)
	assert.Nil(t, body.ExpiresAt)
	assert.Equal(t, "https://s.example.com/r/noexp1", body.ShortUrl)
}

// ── UpdateUrlToResponse ───────────────────────────────────────────────────────

func TestUpdateUrlToResponse_AllFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	exp := now.Add(24 * time.Hour)
	m := &bizmodels.ShortenedUrl{
		ID:          10,
		UserID:      3,
		Shortcode:   "upd001",
		LongUrl:     "https://updated.com",
		ExpiresAt:   &exp,
		LastUpdated: now,
	}

	body := mappers.UpdateUrlToResponse(m, "https://short.example.com")

	require.NotNil(t, body)
	assert.Equal(t, int64(10), body.ID)
	assert.Equal(t, "upd001", body.Shortcode)
	assert.Equal(t, "https://updated.com", body.LongUrl)
	assert.Equal(t, "https://short.example.com/r/upd001", body.ShortUrl)
	require.NotNil(t, body.ExpiresAt)
	assert.Equal(t, exp, *body.ExpiresAt)
	assert.Equal(t, now, body.LastUpdated)
}

func TestUpdateUrlToResponse_NoExpiry(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	m := &bizmodels.ShortenedUrl{
		ID:          11,
		Shortcode:   "noexp2",
		LongUrl:     "https://no-expiry.com",
		LastUpdated: now,
	}

	body := mappers.UpdateUrlToResponse(m, "https://short.example.com")

	require.NotNil(t, body)
	assert.Nil(t, body.ExpiresAt)
	assert.Equal(t, "https://no-expiry.com", body.LongUrl)
	assert.Equal(t, "https://short.example.com/r/noexp2", body.ShortUrl)
}
