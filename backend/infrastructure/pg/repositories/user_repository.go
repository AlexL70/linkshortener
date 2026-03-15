package repositories

import (
	"context"
	"fmt"
	"time"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmappings "github.com/AlexL70/linkshortener/backend/infrastructure/pg/mappings"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
	"github.com/uptrace/bun"
)

type userRepository struct {
	db *bun.DB
}

// NewUserRepository constructs a UserRepository backed by the given *bun.DB.
func NewUserRepository(db *bun.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByProviderID(ctx context.Context, provider bizmodels.Provider, providerUserID string) (*bizmodels.User, *bizmodels.UserProvider, error) {
	up := new(pgmodels.UserProvider)
	err := r.db.NewSelect().
		Model(up).
		Column("up.id", "up.user_id", "up.provider", "up.provider_user_id", "up.provider_email", "up.created_at").
		Where("up.provider = ? AND up.provider_user_id = ?", string(provider), providerUserID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, businesslogic.ErrNotFound
		}
		return nil, nil, fmt.Errorf("UserRepository.FindByProviderID: %w", err)
	}

	user, err := r.findUserByID(ctx, up.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("UserRepository.FindByProviderID: %w", err)
	}

	return pgmappings.UserToBusinessModel(user), pgmappings.UserProviderToBusinessModel(up), nil
}

func (r *userRepository) FindByProviderEmailWithSeedID(ctx context.Context, provider bizmodels.Provider, email string) (*bizmodels.User, *bizmodels.UserProvider, error) {
	up := new(pgmodels.UserProvider)
	err := r.db.NewSelect().
		Model(up).
		Column("up.id", "up.user_id", "up.provider", "up.provider_user_id", "up.provider_email", "up.created_at").
		Where("up.provider = ? AND up.provider_email = ? AND up.provider_user_id = ?", string(provider), email, bizmodels.DevSeedProviderUserID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, businesslogic.ErrNotFound
		}
		return nil, nil, fmt.Errorf("UserRepository.FindByProviderEmailWithSeedID: %w", err)
	}

	user, err := r.findUserByID(ctx, up.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("UserRepository.FindByProviderEmailWithSeedID: %w", err)
	}

	return pgmappings.UserToBusinessModel(user), pgmappings.UserProviderToBusinessModel(up), nil
}

func (r *userRepository) UpdateProviderUserID(ctx context.Context, userProviderID int64, newProviderUserID string) error {
	result, err := r.db.NewUpdate().
		Model((*pgmodels.UserProvider)(nil)).
		Set("provider_user_id = ?", newProviderUserID).
		Where("id = ?", userProviderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("UserRepository.UpdateProviderUserID: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UserRepository.UpdateProviderUserID: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("UserRepository.UpdateProviderUserID: %w", businesslogic.ErrNotFound)
	}
	return nil
}

func (r *userRepository) CreateUserWithProvider(ctx context.Context, userName string, up *bizmodels.UserProvider) (*bizmodels.User, error) {
	now := time.Now()

	dbUser := &pgmodels.User{
		UserName:  userName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, insertErr := tx.NewInsert().Model(dbUser).Returning("id").Exec(ctx); insertErr != nil {
			return fmt.Errorf("insert user: %w", insertErr)
		}

		dbUP := pgmappings.UserProviderToDbModel(up, now)
		dbUP.UserID = dbUser.ID

		if _, insertErr := tx.NewInsert().Model(dbUP).Exec(ctx); insertErr != nil {
			return fmt.Errorf("insert user_provider: %w", insertErr)
		}
		return nil
	})
	if err != nil {
		// PostgreSQL unique violation code is 23505.
		if isUniqueViolation(err) {
			return nil, businesslogic.ErrConflict
		}
		return nil, fmt.Errorf("UserRepository.CreateUserWithProvider: %w", err)
	}

	return pgmappings.UserToBusinessModel(dbUser), nil
}

// findUserByID is a shared helper to look up a single user by primary key.
func (r *userRepository) findUserByID(ctx context.Context, id int64) (*pgmodels.User, error) {
	user := new(pgmodels.User)
	err := r.db.NewSelect().
		Model(user).
		Column("u.id", "u.user_name", "u.created_at", "u.updated_at").
		Where("u.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, businesslogic.ErrNotFound
		}
		return nil, fmt.Errorf("findUserByID: %w", err)
	}
	return user, nil
}
