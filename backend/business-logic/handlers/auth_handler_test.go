package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces/mocks"
	"github.com/AlexL70/linkshortener/backend/business-logic/models"
)

const adminEmail = "admin@gmail.com"

func newHandler(t *testing.T, isDevMode bool) (*handlers.AuthHandler, *mocks.MockUserRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	return handlers.NewAuthHandler(repo, isDevMode, adminEmail), repo
}

func TestResolveUserByProvider_ExistingUser(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-123", Email: "user@example.com"}
	want := &models.User{ID: 1, UserName: "user"}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(want, &models.UserProvider{}, nil)
	got, err := h.ResolveUserByProvider(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestResolveUserByProvider_NewUser(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-new", Email: "new@example.com"}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	_, err := h.ResolveUserByProvider(ctx, input)
	assert.ErrorIs(t, err, businesslogic.ErrNewUser)
}

func TestResolveUserByProvider_RepoError(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-err", Email: "err@example.com"}
	dbErr := errors.New("db connection lost")
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, dbErr)
	_, err := h.ResolveUserByProvider(ctx, input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "db connection lost")
}

func TestResolveUserByProvider_DevMode_AdminSeedUpdate(t *testing.T) {
	h, repo := newHandler(t, true)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "real-sub", Email: adminEmail}
	seedUser := &models.User{ID: 7, UserName: "admin"}
	seedUP := &models.UserProvider{ID: 42, ProviderUserID: models.DevSeedProviderUserID}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	repo.EXPECT().FindByProviderEmailWithSeedID(ctx, input.Provider, adminEmail).Return(seedUser, seedUP, nil)
	repo.EXPECT().UpdateProviderUserID(ctx, seedUP.ID, input.ProviderUserID).Return(nil)
	got, err := h.ResolveUserByProvider(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, seedUser, got)
}

func TestResolveUserByProvider_DevMode_NonAdminNoSeedPath(t *testing.T) {
	h, repo := newHandler(t, true)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-other", Email: "other@example.com"}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	_, err := h.ResolveUserByProvider(ctx, input)
	assert.ErrorIs(t, err, businesslogic.ErrNewUser)
}

func TestResolveUserByProvider_ProdMode_NeverUpdatesSeed(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "real-sub", Email: adminEmail}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	_, err := h.ResolveUserByProvider(ctx, input)
	assert.ErrorIs(t, err, businesslogic.ErrNewUser)
}

func TestResolveUserByProvider_DevMode_AdminSeedNotFound_FallsThrough(t *testing.T) {
	h, repo := newHandler(t, true)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "real-sub", Email: adminEmail}
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	repo.EXPECT().FindByProviderEmailWithSeedID(ctx, input.Provider, adminEmail).Return(nil, nil, businesslogic.ErrNotFound)
	_, err := h.ResolveUserByProvider(ctx, input)
	assert.ErrorIs(t, err, businesslogic.ErrNewUser)
}

func TestResolveUserByProvider_DevMode_AdminSeedFindError(t *testing.T) {
	h, repo := newHandler(t, true)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "real-sub", Email: adminEmail}
	dbErr := errors.New("unexpected db error")
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	repo.EXPECT().FindByProviderEmailWithSeedID(ctx, input.Provider, adminEmail).Return(nil, nil, dbErr)
	_, err := h.ResolveUserByProvider(ctx, input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unexpected db error")
}

func TestResolveUserByProvider_DevMode_SeedUpdateError(t *testing.T) {
	h, repo := newHandler(t, true)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "real-sub", Email: adminEmail}
	seedUser := &models.User{ID: 7, UserName: "admin"}
	seedUP := &models.UserProvider{ID: 42}
	dbErr := errors.New("update failed")
	repo.EXPECT().FindByProviderID(ctx, input.Provider, input.ProviderUserID).Return(nil, nil, businesslogic.ErrNotFound)
	repo.EXPECT().FindByProviderEmailWithSeedID(ctx, input.Provider, adminEmail).Return(seedUser, seedUP, nil)
	repo.EXPECT().UpdateProviderUserID(ctx, seedUP.ID, input.ProviderUserID).Return(dbErr)
	_, err := h.ResolveUserByProvider(ctx, input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "update failed")
}

func TestCreateUser_Success(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-new", Email: "new@example.com"}
	want := &models.User{ID: 10, UserName: "newuser"}
	repo.EXPECT().CreateUserWithProvider(ctx, "newuser", gomock.Any()).Return(want, nil)
	got, err := h.CreateUser(ctx, "newuser", input)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateUser_DuplicateUserName(t *testing.T) {
	h, repo := newHandler(t, false)
	ctx := context.Background()
	input := &models.AuthInput{Provider: models.ProviderGoogle, ProviderUserID: "sub-dup", Email: "dup@example.com"}
	repo.EXPECT().CreateUserWithProvider(ctx, "taken", gomock.Any()).Return(nil, businesslogic.ErrConflict)
	_, err := h.CreateUser(ctx, "taken", input)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}
