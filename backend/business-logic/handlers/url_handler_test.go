package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces/mocks"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
)

func newUrlHandler(t *testing.T) (*handlers.UrlHandler, *mocks.MockUrlRepository, *mocks.MockShortcodeGenerator) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUrlRepository(ctrl)
	gen := mocks.NewMockShortcodeGenerator(ctrl)
	return handlers.NewUrlHandler(repo, gen, 2048, 6, 6, 10), repo, gen
}

func TestListUrls_Success(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	want := []*bizmodels.ShortenedUrl{
		{ID: 1, UserID: 42, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now},
		{ID: 2, UserID: 42, Shortcode: "def456", LongUrl: "https://other.com", LastUpdated: now},
	}
	repo.EXPECT().FindByUserID(ctx, int64(42), 1, 20).Return(want, 2, nil)

	got, total, err := h.ListUrls(ctx, 42, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, total)
}

func TestListUrls_EmptyList(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	repo.EXPECT().FindByUserID(ctx, int64(99), 1, 20).Return([]*bizmodels.ShortenedUrl{}, 0, nil)

	got, total, err := h.ListUrls(ctx, 99, 1, 20)
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, 0, total)
}

func TestListUrls_RepoError(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
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
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	url := &bizmodels.ShortenedUrl{ID: 3, UserID: 7, Shortcode: "xyz789", LongUrl: "https://paginated.com"}
	repo.EXPECT().FindByUserID(ctx, int64(7), 2, 5).Return([]*bizmodels.ShortenedUrl{url}, 6, nil)

	got, total, err := h.ListUrls(ctx, 7, 2, 5)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 6, total)
}

// ── CreateUrl ─────────────────────────────────────────────────────────────────

func TestCreateUrl_AutoShortcode_Success(t *testing.T) {
	h, repo, gen := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	want := &bizmodels.ShortenedUrl{ID: 10, UserID: 1, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	gen.EXPECT().GenerateShortcode().Return("abc123", nil)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(want, nil)

	got, err := h.CreateUrl(ctx, 1, "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateUrl_CustomShortcode_Success(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	sc := "my-sc1"
	want := &bizmodels.ShortenedUrl{ID: 11, UserID: 2, Shortcode: sc, LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().Create(ctx, gomock.Any()).Return(want, nil)

	got, err := h.CreateUrl(ctx, 2, "https://example.com", &sc, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateUrl_WithExpiry_Success(t *testing.T) {
	h, repo, gen := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	exp := now.Add(72 * time.Hour)
	want := &bizmodels.ShortenedUrl{ID: 12, UserID: 3, Shortcode: "exp123", LongUrl: "https://example.com", ExpiresAt: &exp, LastUpdated: now}

	gen.EXPECT().GenerateShortcode().Return("exp123", nil)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(want, nil)

	got, err := h.CreateUrl(ctx, 3, "https://example.com", nil, &exp)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateUrl_InvalidUrl_Scheme(t *testing.T) {
	h, _, _ := newUrlHandler(t)
	_, err := h.CreateUrl(context.Background(), 1, "ftp://example.com", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreateUrl_InvalidUrl_Localhost(t *testing.T) {
	h, _, _ := newUrlHandler(t)
	_, err := h.CreateUrl(context.Background(), 1, "http://localhost/admin", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreateUrl_InvalidUrl_PrivateIP(t *testing.T) {
	h, _, _ := newUrlHandler(t)
	_, err := h.CreateUrl(context.Background(), 1, "http://192.168.1.1/secret", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreateUrl_InvalidCustomShortcode_WrongLength(t *testing.T) {
	h, _, _ := newUrlHandler(t)
	sc := "ab"
	_, err := h.CreateUrl(context.Background(), 1, "https://example.com", &sc, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreateUrl_InvalidCustomShortcode_Reserved(t *testing.T) {
	h, _, _ := newUrlHandler(t)
	sc := "admin1"
	_, err := h.CreateUrl(context.Background(), 1, "https://example.com", &sc, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestCreateUrl_CollidesOnce_RetriesSuccessfully(t *testing.T) {
	h, repo, gen := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	want := &bizmodels.ShortenedUrl{ID: 20, UserID: 5, Shortcode: "second", LongUrl: "https://example.com", LastUpdated: now}

	gomock.InOrder(
		gen.EXPECT().GenerateShortcode().Return("first0", nil),
		gen.EXPECT().GenerateShortcode().Return("second", nil),
	)
	gomock.InOrder(
		repo.EXPECT().Create(ctx, gomock.Any()).Return(nil, businesslogic.ErrConflict),
		repo.EXPECT().Create(ctx, gomock.Any()).Return(want, nil),
	)

	got, err := h.CreateUrl(ctx, 5, "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateUrl_ExhaustsRetries_ReturnsConflict(t *testing.T) {
	h, repo, gen := newUrlHandler(t)
	ctx := context.Background()

	gen.EXPECT().GenerateShortcode().Return("collide", nil).Times(10)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(nil, businesslogic.ErrConflict).Times(10)

	_, err := h.CreateUrl(ctx, 6, "https://example.com", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}

func TestCreateUrl_GeneratorError(t *testing.T) {
	h, _, gen := newUrlHandler(t)
	genErr := errors.New("entropy failure")
	gen.EXPECT().GenerateShortcode().Return("", genErr)

	_, err := h.CreateUrl(context.Background(), 7, "https://example.com", nil, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "entropy failure")
}

func TestCreateUrl_RepoCreateError(t *testing.T) {
	h, repo, gen := newUrlHandler(t)
	ctx := context.Background()
	dbErr := errors.New("db failure")

	gen.EXPECT().GenerateShortcode().Return("abc456", nil)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(nil, dbErr)

	_, err := h.CreateUrl(ctx, 8, "https://example.com", nil, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "db failure")
}

// ── UpdateUrl ─────────────────────────────────────────────────────────────────

func TestUpdateUrl_Success(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	existing := &bizmodels.ShortenedUrl{ID: 1, UserID: 10, Shortcode: "old-sc", LongUrl: "https://old.com", LastUpdated: now}
	updated := &bizmodels.ShortenedUrl{ID: 1, UserID: 10, Shortcode: "old-sc", LongUrl: "https://new.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(1)).Return(existing, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(updated, nil)

	got, err := h.UpdateUrl(ctx, 1, 10, "https://new.com", nil, nil, now)
	require.NoError(t, err)
	assert.Equal(t, updated, got)
}

func TestUpdateUrl_ChangeShortcode_Success(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	sc := "new-sc"
	existing := &bizmodels.ShortenedUrl{ID: 2, UserID: 10, Shortcode: "old-sc", LongUrl: "https://example.com", LastUpdated: now}
	updated := &bizmodels.ShortenedUrl{ID: 2, UserID: 10, Shortcode: sc, LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(2)).Return(existing, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(updated, nil)

	got, err := h.UpdateUrl(ctx, 2, 10, "https://example.com", &sc, nil, now)
	require.NoError(t, err)
	assert.Equal(t, updated, got)
}

func TestUpdateUrl_NotFound_ReturnsNotFound(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()

	repo.EXPECT().FindByID(ctx, int64(99)).Return(nil, businesslogic.ErrNotFound)

	_, err := h.UpdateUrl(ctx, 99, 10, "https://example.com", nil, nil, time.Now())
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

func TestUpdateUrl_WrongOwner_ReturnsUnauthorized(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	existing := &bizmodels.ShortenedUrl{ID: 3, UserID: 99, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(3)).Return(existing, nil)

	_, err := h.UpdateUrl(ctx, 3, 10, "https://example.com", nil, nil, now)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrUnauthorized)
}

func TestUpdateUrl_InvalidUrl_ReturnsValidation(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	existing := &bizmodels.ShortenedUrl{ID: 4, UserID: 10, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(4)).Return(existing, nil)

	_, err := h.UpdateUrl(ctx, 4, 10, "ftp://bad-scheme.com", nil, nil, now)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestUpdateUrl_InvalidShortcode_ReturnsValidation(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	sc := "ab"
	existing := &bizmodels.ShortenedUrl{ID: 5, UserID: 10, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(5)).Return(existing, nil)

	_, err := h.UpdateUrl(ctx, 5, 10, "https://example.com", &sc, nil, now)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestUpdateUrl_ShortcodeConflict_ReturnsConflict(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	sc := "taken1"
	existing := &bizmodels.ShortenedUrl{ID: 6, UserID: 10, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(6)).Return(existing, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(nil, businesslogic.ErrConflict)

	_, err := h.UpdateUrl(ctx, 6, 10, "https://example.com", &sc, nil, now)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}

func TestUpdateUrl_RepoUpdateError(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	existing := &bizmodels.ShortenedUrl{ID: 7, UserID: 10, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}
	dbErr := errors.New("db failure")

	repo.EXPECT().FindByID(ctx, int64(7)).Return(existing, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(nil, dbErr)

	_, err := h.UpdateUrl(ctx, 7, 10, "https://example.com", nil, nil, now)
	require.Error(t, err)
	assert.ErrorContains(t, err, "db failure")
}

func TestUpdateUrl_VersionConflict_ReturnsVersionConflict(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()
	existing := &bizmodels.ShortenedUrl{ID: 8, UserID: 10, Shortcode: "abc123", LongUrl: "https://example.com", LastUpdated: now}

	repo.EXPECT().FindByID(ctx, int64(8)).Return(existing, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).Return(nil, businesslogic.ErrVersionConflict)

	staleVersion := now.Add(-5 * time.Minute)
	_, err := h.UpdateUrl(ctx, 8, 10, "https://example.com", nil, nil, staleVersion)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrVersionConflict)
}

// ── DeleteUrl ─────────────────────────────────────────────────────────────────

func TestDeleteUrl_Success(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()

	repo.EXPECT().Delete(ctx, int64(1), int64(10), now).Return(nil)

	err := h.DeleteUrl(ctx, 1, 10, now)
	require.NoError(t, err)
}

func TestDeleteUrl_NotFound_ReturnsNotFound(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()

	repo.EXPECT().Delete(ctx, int64(99), int64(10), gomock.Any()).Return(businesslogic.ErrNotFound)

	err := h.DeleteUrl(ctx, 99, 10, time.Now())
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

func TestDeleteUrl_VersionConflict_ReturnsVersionConflict(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	now := time.Now()

	repo.EXPECT().Delete(ctx, int64(5), int64(10), gomock.Any()).Return(businesslogic.ErrVersionConflict)

	staleVersion := now.Add(-5 * time.Minute)
	err := h.DeleteUrl(ctx, 5, 10, staleVersion)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrVersionConflict)
}

func TestDeleteUrl_RepoError_Wraps(t *testing.T) {
	h, repo, _ := newUrlHandler(t)
	ctx := context.Background()
	dbErr := errors.New("db connection lost")

	repo.EXPECT().Delete(ctx, int64(3), int64(10), gomock.Any()).Return(dbErr)

	err := h.DeleteUrl(ctx, 3, 10, time.Now())
	require.Error(t, err)
	assert.ErrorContains(t, err, "db connection lost")
}
