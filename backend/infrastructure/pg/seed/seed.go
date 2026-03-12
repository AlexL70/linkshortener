package seed

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	pgmodels "github.com/AlexL70/linkshortener/backend/infrastructure/pg/models"
	"github.com/uptrace/bun"
)

//go:embed urls.json
var urlsJSON []byte

type seedURL struct {
	Shortcode string `json:"shortcode"`
	LongURL   string `json:"long_url"`
}

// RunIfEmpty seeds the database with a dev user and shortened URLs when the
// Users table is empty. Errors are logged as WARN and never abort startup.
// Must only be called in dev mode.
func RunIfEmpty(ctx context.Context, db *bun.DB) {
	count, err := db.NewSelect().Model((*pgmodels.User)(nil)).Count(ctx)
	if err != nil {
		slog.Warn("seed: failed to count users, skipping seed", "error", err)
		return
	}
	if count > 0 {
		return
	}

	if err := run(ctx, db); err != nil {
		slog.Warn("seed: seeding failed, application will still start", "error", err)
	}
}

func run(ctx context.Context, db *bun.DB) error {
	email := os.Getenv("SUPER_ADMIN_EMAIL")
	if email == "" {
		return fmt.Errorf("SUPER_ADMIN_EMAIL is not set")
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid SUPER_ADMIN_EMAIL: %q", email)
	}
	userName := parts[0]
	domain := strings.ToLower(parts[1])
	provider := inferProvider(domain) // returns bizmodels.Provider

	now := time.Now()

	// Insert seed user.
	user := &pgmodels.User{
		UserName:  userName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(user).Returning("id").Exec(ctx); err != nil {
		return fmt.Errorf("seed: insert user: %w", err)
	}
	slog.Info("seed: created user", "user_name", user.UserName, "id", user.ID)

	// Insert user provider record (no real OAuth involved — dev only).
	up := &pgmodels.UserProvider{
		UserID:         user.ID,
		Provider:       string(provider),
		ProviderUserID: bizmodels.DevSeedProviderUserID,
		ProviderEmail:  email,
		CreatedAt:      now,
	}
	if _, err := db.NewInsert().Model(up).Returning("id").Exec(ctx); err != nil {
		return fmt.Errorf("seed: insert user provider: %w", err)
	}
	slog.Info("seed: created user provider", "provider", up.Provider, "email", up.ProviderEmail, "id", up.ID)

	// Parse seed URL list from embedded JSON.
	var allURLs []seedURL
	if err := json.Unmarshal(urlsJSON, &allURLs); err != nil {
		return fmt.Errorf("seed: parse urls.json: %w", err)
	}

	// Pick between 5 and 10 URLs without replacement using a random permutation.
	pickCount := 5 + rand.Intn(6) // [5, 10]
	if pickCount > len(allURLs) {
		pickCount = len(allURLs)
	}
	indices := rand.Perm(len(allURLs))

	for i := 0; i < pickCount; i++ {
		u := allURLs[indices[i]]
		row := &pgmodels.ShortenedUrl{
			UserID:    user.ID,
			Shortcode: u.Shortcode,
			LongUrl:   u.LongURL,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := db.NewInsert().Model(row).Returning("id").Exec(ctx); err != nil {
			return fmt.Errorf("seed: insert shortened URL %q: %w", u.Shortcode, err)
		}
		slog.Info("seed: created shortened URL", "shortcode", row.Shortcode, "long_url", row.LongUrl, "id", row.ID)
	}

	return nil
}

// inferProvider maps an email domain to the matching OAuth provider name.
func inferProvider(domain string) bizmodels.Provider {
	switch domain {
	case "gmail.com":
		return bizmodels.ProviderGoogle
	case "outlook.com", "hotmail.com", "live.com", "msn.com":
		return bizmodels.ProviderMicrosoft
	case "facebook.com":
		return bizmodels.ProviderFacebook
	default:
		// Google is the most commonly configured provider in dev.
		return bizmodels.ProviderGoogle
	}
}
