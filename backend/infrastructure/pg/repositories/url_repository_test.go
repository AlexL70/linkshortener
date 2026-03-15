package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/repositories"
)

// insertTestURL inserts a ShortenedUrl row directly via Bun for test setup.
func insertTestURL(t *testing.T, db *bun.DB, userID int64, shortcode, longURL string) int64 {
	t.Helper()
	ctx := context.Background()
	now := time.Now()
	row := &pgmodels.ShortenedUrl{
		UserID:    userID,
		Shortcode: shortcode,
		LongUrl:   longURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := db.NewInsert().Model(row).Returning("id").Exec(ctx)
	require.NoError(t, err, "insertTestURL: %s", shortcode)
	return row.ID
}

func TestFindByUserID_ReturnsPaginatedURLs(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "url-list-test-sub",
		ProviderEmail:  "url-list@example.com",
	}
	user, err := userRepo.CreateUserWithProvider(ctx, "urllstuser1", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	insertTestURL(t, db, user.ID, "aaa111", "https://a.com")
	insertTestURL(t, db, user.ID, "bbb222", "https://b.com")
	insertTestURL(t, db, user.ID, "ccc333", "https://c.com")

	urlRepo := repositories.NewUrlRepository(db)
	urls, total, err := urlRepo.FindByUserID(ctx, user.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, urls, 3)
}

func TestFindByUserID_EmptyList(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "url-empty-test-sub",
		ProviderEmail:  "url-empty@example.com",
	}
	user, err := userRepo.CreateUserWithProvider(ctx, "urlemptyuser", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	urlRepo := repositories.NewUrlRepository(db)
	urls, total, err := urlRepo.FindByUserID(ctx, user.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, urls)
}

func TestFindByUserID_PaginationOffset(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{
		Provider:       bizmodels.ProviderGoogle,
		ProviderUserID: "url-page-test-sub",
		ProviderEmail:  "url-page@example.com",
	}
	user, err := userRepo.CreateUserWithProvider(ctx, "urlpageuser", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	for _, sc := range []string{"pg0001", "pg0002", "pg0003", "pg0004", "pg0005"} {
		insertTestURL(t, db, user.ID, sc, "https://paginate.com/"+sc)
		// Small sleep so created_at timestamps differ and ordering is deterministic.
		time.Sleep(2 * time.Millisecond)
	}

	urlRepo := repositories.NewUrlRepository(db)

	// Page 1 with page size 2 should return 2 records; total is 5.
	page1, total, err := urlRepo.FindByUserID(ctx, user.ID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, page1, 2)

	// Page 3 with page size 2 should return the 1 remaining record.
	page3, total2, err := urlRepo.FindByUserID(ctx, user.ID, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total2)
	assert.Len(t, page3, 1)
}

func TestFindByUserID_IsolatedToOwner(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)

	up1 := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "iso-sub-1", ProviderEmail: "iso1@example.com"}
	user1, err := userRepo.CreateUserWithProvider(ctx, "isoluser1", up1)
	require.NoError(t, err)
	defer cleanUsers(t, db, user1.ID)

	up2 := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "iso-sub-2", ProviderEmail: "iso2@example.com"}
	user2, err := userRepo.CreateUserWithProvider(ctx, "isoluser2", up2)
	require.NoError(t, err)
	defer cleanUsers(t, db, user2.ID)

	insertTestURL(t, db, user1.ID, "u1abc1", "https://user1.com")
	insertTestURL(t, db, user2.ID, "u2xyz1", "https://user2.com")

	urlRepo := repositories.NewUrlRepository(db)

	urls1, total1, err := urlRepo.FindByUserID(ctx, user1.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total1)
	assert.Len(t, urls1, 1)
	assert.Equal(t, user1.ID, urls1[0].UserID)

	urls2, total2, err := urlRepo.FindByUserID(ctx, user2.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total2)
	assert.Len(t, urls2, 1)
	assert.Equal(t, user2.ID, urls2[0].UserID)
}

// openTestDB opens a *bun.DB from the DATABASE_URL environment variable.
// Tests are skipped when DATABASE_URL is not set (re-declared locally to keep this
// file self-contained; the function is also present in user_repository_test.go but
// that file is in the same package so we cannot re-declare it there — the shared
// helper lives in the existing file and this call will reuse it).
//
// openTestDB and cleanUsers are defined in user_repository_test.go and are
// accessible here since both files share the repositories_test package.

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "create-sub-1", ProviderEmail: "create1@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "createuser1", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{
		UserID:    user.ID,
		Shortcode: "cr0001",
		LongUrl:   "https://create-test.com",
	}
	created, err := urlRepo.Create(ctx, input)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "cr0001", created.Shortcode)
	assert.Equal(t, "https://create-test.com", created.LongUrl)
	assert.Equal(t, user.ID, created.UserID)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestCreate_WithExpiry_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "create-sub-exp", ProviderEmail: "createexp@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "createuserexp", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	exp := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{
		UserID:    user.ID,
		Shortcode: "expurl",
		LongUrl:   "https://expiry-test.com",
		ExpiresAt: &exp,
	}
	created, err := urlRepo.Create(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created.ExpiresAt)
	assert.Equal(t, exp.Unix(), created.ExpiresAt.Unix())
}

func TestCreate_DuplicateShortcode_ReturnsConflict(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "create-sub-dup", ProviderEmail: "createdup@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "createuserdup", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{UserID: user.ID, Shortcode: "dup001", LongUrl: "https://dup.com"}
	_, err = urlRepo.Create(ctx, input)
	require.NoError(t, err)

	// Second insert with the same shortcode must return ErrConflict.
	_, err = urlRepo.Create(ctx, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}

// ── FindByID ──────────────────────────────────────────────────────────────────

func TestFindByID_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "fbi-sub-1", ProviderEmail: "fbi1@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "fbiuser1", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	id := insertTestURL(t, db, user.ID, "fbi001", "https://findbyid.com")

	urlRepo := repositories.NewUrlRepository(db)
	got, err := urlRepo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "fbi001", got.Shortcode)
	assert.Equal(t, "https://findbyid.com", got.LongUrl)
	assert.Equal(t, user.ID, got.UserID)
}

func TestFindByID_NotFound_ReturnsErrNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	urlRepo := repositories.NewUrlRepository(db)

	_, err := urlRepo.FindByID(ctx, -999)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrNotFound)
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "upd-sub-1", ProviderEmail: "upd1@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "upduser1", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	id := insertTestURL(t, db, user.ID, "up0001", "https://original.com")

	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{
		ID:        id,
		UserID:    user.ID,
		Shortcode: "up0001",
		LongUrl:   "https://updated.com",
	}
	updated, err := urlRepo.Update(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, "https://updated.com", updated.LongUrl)
	assert.Nil(t, updated.ExpiresAt)
	assert.False(t, updated.UpdatedAt.IsZero())
}

func TestUpdate_ChangeShortcode_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "upd-sub-2", ProviderEmail: "upd2@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "upduser2", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	id := insertTestURL(t, db, user.ID, "old001", "https://example.com")

	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{
		ID:        id,
		UserID:    user.ID,
		Shortcode: "new001",
		LongUrl:   "https://example.com",
	}
	updated, err := urlRepo.Update(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "new001", updated.Shortcode)
}

func TestUpdate_DuplicateShortcode_ReturnsConflict(t *testing.T) {
	db := openTestDB(t)
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db)
	up := &bizmodels.UserProvider{Provider: bizmodels.ProviderGoogle, ProviderUserID: "upd-sub-3", ProviderEmail: "upd3@example.com"}
	user, err := userRepo.CreateUserWithProvider(ctx, "upduser3", up)
	require.NoError(t, err)
	defer cleanUsers(t, db, user.ID)

	insertTestURL(t, db, user.ID, "taken2", "https://taken.com")
	id := insertTestURL(t, db, user.ID, "other2", "https://other.com")

	urlRepo := repositories.NewUrlRepository(db)
	input := &bizmodels.ShortenedUrl{
		ID:        id,
		UserID:    user.ID,
		Shortcode: "taken2", // conflict with existing row
		LongUrl:   "https://other.com",
	}
	_, err = urlRepo.Update(ctx, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrConflict)
}
