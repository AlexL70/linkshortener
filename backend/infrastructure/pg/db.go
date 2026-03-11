package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

// Open creates and configures a *bun.DB from the DATABASE_URL environment
// variable and pool settings read from the environment.
func Open() (*bun.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("pg.Open: DATABASE_URL is not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	maxOpen := intEnv("DB_MAX_OPEN_CONNS", 25)
	maxIdle := intEnv("DB_MAX_IDLE_CONNS", 5)
	maxLifetime := intEnv("DB_CONN_MAX_LIFETIME", 3600)
	maxIdleTime := intEnv("DB_CONN_MAX_IDLE_TIME", 600)

	sqldb.SetMaxOpenConns(maxOpen)
	sqldb.SetMaxIdleConns(maxIdle)
	sqldb.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	sqldb.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}

// RunMigrations applies any pending migrations using the registered migration set.
func RunMigrations(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("pg.RunMigrations: init migrator: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("pg.RunMigrations: migrate: %w", err)
	}

	if group.IsZero() {
		slog.Info("db migrations: already up to date")
	} else {
		slog.Info("db migrations: applied", "group", group.String())
	}

	return nil
}

// intEnv reads an integer from an environment variable. Falls back to
// defaultVal when the variable is absent or unparseable.
func intEnv(key string, defaultVal int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("intEnv: invalid value, using default", "key", key, "value", raw, "default", defaultVal)
		return defaultVal
	}
	return v
}
