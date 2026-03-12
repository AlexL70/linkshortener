package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces/mocks"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

func newUrlHandler(t *testing.T) (*handlers.UrlHandler, *mocks.MockUrlRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUrlRepository(ctrl)
	return handlers.NewUrlHandler(repo), repo
}

func TestListUrls_Success(t *testing.T) {
	h, repo := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	want := []*bizmodels.ShortenedUrl{
		{ID: 1, UserID: 42, Shortcode: "abc123", LongUrl: "https://example.com", CreatedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 42, Shortcode: "def456", LongUrl: "https://other.com", CreatedAt: now, UpdatedAt: now},
	}
	repo.EXPECT().FindByUserID(ctx, int64(42), 1, 20).Return(want, 2, nil)

	got, total, err := h.ListUrls(ctx, 42, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, total)
}

func TestListUrls_EmptyList(t *testing.T) {
	h, repo := newUrlHandler(t)
	ctx := context.Background()
	repo.EXPECT().FindByUserID(ctx, int64(99), 1, 20).Return([]*bizmodels.ShortenedUrl{}, 0, nil)

	got, total, err := h.ListUrls(ctx, 99, 1, 20)
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, 0, total)
}

func TestListUrls_RepoError(t *testing.T) {
	h, repo := newUrlHandler(t)
	ctx := context.Background()
	dbErr := errors.New("db connection lost")
	repo.EXPECT().FindByUserID(ctx, int64(1), 1, 20).Return(nil, 0, dbErr)

	got, total, err := h.ListUrls(ctx, 1, 1, 20)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, 0, total)
	assert.ErrorContains(t, err, "db connection lost")
}

func TestListUrls_PaginationPassthrough(t *testing.T) {
	h, repo := newUrlHandler(t)
	ctx := context.Background()
	url := &bizmodels.ShortenedUrl{ID: 3, UserID: 7, Shortcode: "xyz789", LongUrl: "https://paginated.com"}
	repo.EXPECT().FindByUserID(ctx, int64(7), 2, 5).Return([]*bizmodels.ShortenedUrl{url}, 6, nil)

	got, total, err := h.ListUrls(ctx, 7, 2, 5)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 6, total)
}
