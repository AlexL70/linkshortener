package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/repositories"
	"github.com/AlexL70/linkshortener/backend/testutil"
)

func TestFindByProviderID_Found(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "test-find-by-provider-id",
		ProviderEmail:  "test-fbp@example.com",
	}
	created, err := repo.CreateUserWithProvider(ctx, "testfindprovider", up)
	require.NoError(t, err)

	user, gotUP, err := repo.FindByProviderID(ctx, bizmodels.ProviderGoogle, "test-find-by-provider-id")
	require.NoError(t, err)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, "testfindprovider", user.UserName)
	assert.Equal(t, "test-find-by-provider-id", gotUP.ProviderUserID)
}

func TestFindByProviderID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	_, _, err := repo.FindByProviderID(ctx, bizmodels.ProviderGoogle, "nonexistent-sub")
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

func TestFindByProviderEmailWithSeedID_Found(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: bizmodels.DevSeedProviderUserID,
		ProviderEmail:  "seed-admin@gmail.com",
	}
	created, err := repo.CreateUserWithProvider(ctx, "seedfinduser", up)
	require.NoError(t, err)

	user, gotUP, err := repo.FindByProviderEmailWithSeedID(ctx, bizmodels.ProviderGoogle, "seed-admin@gmail.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, bizmodels.DevSeedProviderUserID, gotUP.ProviderUserID)
}

func TestFindByProviderEmailWithSeedID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	_, _, err := repo.FindByProviderEmailWithSeedID(ctx, bizmodels.ProviderGoogle, "no-such-seed@gmail.com")
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

func TestUpdateProviderUserID_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: bizmodels.DevSeedProviderUserID,
		ProviderEmail:  "update-test@gmail.com",
	}
	created, err := repo.CreateUserWithProvider(ctx, "updateprovideruser", up)
	require.NoError(t, err)

	_, gotUP, err := repo.FindByProviderEmailWithSeedID(ctx, bizmodels.ProviderGoogle, "update-test@gmail.com")
	require.NoError(t, err)

	err = repo.UpdateProviderUserID(ctx, gotUP.ID, "real-google-sub-updated")
	require.NoError(t, err)

	// Verify the update took effect.
	user2, gotUP2, err := repo.FindByProviderID(ctx, bizmodels.ProviderGoogle, "real-google-sub-updated")
	require.NoError(t, err)
	assert.Equal(t, created.ID, user2.ID)
	assert.Equal(t, "real-google-sub-updated", gotUP2.ProviderUserID)
}

func TestUpdateProviderUserID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	err := repo.UpdateProviderUserID(ctx, 999999999, "some-sub")
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

func TestCreateUserWithProvider_DuplicateUserName(t *testing.T) {
	db := testutil.OpenTestDB(t)

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	up1 := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "dup-user-1",
		ProviderEmail:  "dup1@example.com",
	}
	_, err := repo.CreateUserWithProvider(ctx, "dupuser", up1)
	require.NoError(t, err)

	up2 := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "dup-user-2",
		ProviderEmail:  "dup2@example.com",
	}
	_, err = repo.CreateUserWithProvider(ctx, "dupuser", up2)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}
