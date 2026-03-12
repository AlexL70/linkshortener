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
		ID:        42,
		UserID:    7,
		Shortcode: "abc123",
		LongUrl:   "https://example.com",
		ExpiresAt: &exp,
		CreatedAt: now,
		UpdatedAt: now,
	}

	vm := mappers.UrlToViewModel(m)

	require.NotNil(t, vm)
	assert.Equal(t, int64(42), vm.ID)
	assert.Equal(t, "abc123", vm.Shortcode)
	assert.Equal(t, "https://example.com", vm.LongUrl)
	require.NotNil(t, vm.ExpiresAt)
	assert.Equal(t, exp, *vm.ExpiresAt)
	assert.Equal(t, now, vm.CreatedAt)
	assert.Equal(t, now, vm.UpdatedAt)
}

func TestUrlToViewModel_NoExpiry(t *testing.T) {
	now := time.Now()
	m := &bizmodels.ShortenedUrl{
		ID:        1,
		Shortcode: "xyz789",
		LongUrl:   "https://no-expiry.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	vm := mappers.UrlToViewModel(m)

	require.NotNil(t, vm)
	assert.Nil(t, vm.ExpiresAt)
}

func TestListUrlsToResponse_MultipleItems(t *testing.T) {
	now := time.Now()
	urls := []*bizmodels.ShortenedUrl{
		{ID: 1, Shortcode: "aaa111", LongUrl: "https://a.com", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Shortcode: "bbb222", LongUrl: "https://b.com", CreatedAt: now, UpdatedAt: now},
	}

	body := mappers.ListUrlsToResponse(urls, 10, 2, 5)

	require.NotNil(t, body)
	assert.Equal(t, 10, body.Total)
	assert.Equal(t, 2, body.Page)
	assert.Equal(t, 5, body.PageSize)
	assert.Len(t, body.Items, 2)
	assert.Equal(t, "aaa111", body.Items[0].Shortcode)
	assert.Equal(t, "bbb222", body.Items[1].Shortcode)
}

func TestListUrlsToResponse_EmptySlice(t *testing.T) {
	body := mappers.ListUrlsToResponse([]*bizmodels.ShortenedUrl{}, 0, 1, 20)

	require.NotNil(t, body)
	assert.Equal(t, 0, body.Total)
	assert.Empty(t, body.Items)
}
