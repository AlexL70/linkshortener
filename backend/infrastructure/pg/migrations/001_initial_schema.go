package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// initialSchemaUp contains every DDL statement that brings the schema from
// zero to v1. Raw SQL is required because Bun's query builder does not support
// multi-column UNIQUE constraints, composite index declarations, or complex DDL
// across multiple statements in a single call.
var initialSchemaUp = []string{
	`CREATE TABLE IF NOT EXISTS "Users" (
		"id"         BIGSERIAL PRIMARY KEY,
		"user_name"  VARCHAR NOT NULL,
		"created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT "Users_user_name_key" UNIQUE ("user_name")
	)`,

	`CREATE TABLE IF NOT EXISTS "UserProviders" (
		"id"               BIGSERIAL PRIMARY KEY,
		"user_id"          BIGINT NOT NULL REFERENCES "Users" ("id") ON DELETE CASCADE,
		"provider"         VARCHAR NOT NULL,
		"provider_user_id" VARCHAR NOT NULL,
		"provider_email"   VARCHAR,
		"created_at"       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT "UserProviders_provider_user_id_key" UNIQUE ("provider", "provider_user_id")
	)`,
	`CREATE INDEX IF NOT EXISTS "UserProviders_user_id_idx" ON "UserProviders" ("user_id")`,

	`CREATE TABLE IF NOT EXISTS "ShortenedUrls" (
		"id"         BIGSERIAL PRIMARY KEY,
		"user_id"    BIGINT NOT NULL REFERENCES "Users" ("id") ON DELETE CASCADE,
		"shortcode"  VARCHAR NOT NULL,
		"long_url"   TEXT NOT NULL,
		"expires_at" TIMESTAMPTZ,
		"created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT "ShortenedUrls_shortcode_key" UNIQUE ("shortcode")
	)`,
	`CREATE INDEX IF NOT EXISTS "ShortenedUrls_shortcode_idx" ON "ShortenedUrls" ("shortcode")`,
	`CREATE INDEX IF NOT EXISTS "ShortenedUrls_user_id_idx" ON "ShortenedUrls" ("user_id")`,

	`CREATE TABLE IF NOT EXISTS "UrlClicks" (
		"id"         BIGSERIAL PRIMARY KEY,
		"url_id"     BIGINT NOT NULL REFERENCES "ShortenedUrls" ("id") ON DELETE CASCADE,
		"clicked_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"ip_address" VARCHAR,
		"user_agent" TEXT,
		"referer"    TEXT
	)`,
	`CREATE INDEX IF NOT EXISTS "UrlClicks_url_id_idx"    ON "UrlClicks" ("url_id")`,
	`CREATE INDEX IF NOT EXISTS "UrlClicks_clicked_at_idx" ON "UrlClicks" ("clicked_at")`,
	`CREATE INDEX IF NOT EXISTS "UrlClicks_ip_address_idx" ON "UrlClicks" ("ip_address")`,
}

// initialSchemaDown contains every statement that undoes initialSchemaUp.
var initialSchemaDown = []string{
	`DROP TABLE IF EXISTS "UrlClicks" CASCADE`,
	`DROP TABLE IF EXISTS "ShortenedUrls" CASCADE`,
	`DROP TABLE IF EXISTS "UserProviders" CASCADE`,
	`DROP TABLE IF EXISTS "Users" CASCADE`,
}

func init() {
	Migrations.MustRegister(upInitialSchema, downInitialSchema)
	SQLScripts = append(SQLScripts, SQLScript{
		Name: "001_initial_schema",
		Up:   initialSchemaUp,
		Down: initialSchemaDown,
	})
}

func upInitialSchema(ctx context.Context, db *bun.DB) error {
	for _, stmt := range initialSchemaUp {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("001_initial_schema up: %w", err)
		}
	}
	return nil
}

func downInitialSchema(ctx context.Context, db *bun.DB) error {
	for _, stmt := range initialSchemaDown {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("001_initial_schema down: %w", err)
		}
	}
	return nil
}

// Compile-time check that the functions satisfy the expected signature.
var _ migrate.MigrationFunc = upInitialSchema
var _ migrate.MigrationFunc = downInitialSchema
