package testutil

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/AlexL70/linkshortener/backend/infrastructure/pg"
)

// OpenTestDB creates a fresh, isolated PostgreSQL database for the duration of
// a single test. The database name is "test_<16 random hex chars>".
//
// The schema is brought up to date by running all registered migrations. A
// t.Cleanup function is registered that drops the database and closes the
// connection automatically, so callers need no teardown code of their own.
//
// The test is skipped (rather than failed) when DATABASE_URL is not set, which
// lets the test suite degrade gracefully in CI environments that do not provide
// a PostgreSQL server.
func OpenTestDB(t *testing.T) *bun.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration tests")
	}

	dbName, err := randomTestDBName()
	if err != nil {
		t.Fatalf("OpenTestDB: generate db name: %v", err)
	}

	testDSN, err := replaceDBName(dsn, dbName)
	if err != nil {
		t.Fatalf("OpenTestDB: build test DSN: %v", err)
	}

	ctx := context.Background()

	if err := createDatabase(ctx, dsn, dbName); err != nil {
		t.Skipf("OpenTestDB: cannot reach PostgreSQL server (%v); skipping integration tests", err)
	}

	db, err := openBunDB(testDSN)
	if err != nil {
		_ = dropDatabase(context.Background(), dsn, dbName) //nolint:errcheck — best-effort, we are already failing
		t.Fatalf("OpenTestDB: open bun db: %v", err)
	}

	if err := pg.RunMigrations(ctx, db); err != nil {
		db.Close()                                          //nolint:errcheck
		_ = dropDatabase(context.Background(), dsn, dbName) //nolint:errcheck
		t.Fatalf("OpenTestDB: run migrations: %v", err)
	}

	t.Cleanup(func() {
		db.Close() //nolint:errcheck — best-effort close
		if err := dropDatabase(context.Background(), dsn, dbName); err != nil {
			t.Logf("OpenTestDB cleanup: failed to drop %q: %v", dbName, err)
		}
	})

	return db
}

// randomTestDBName generates a name in the form "test_<16 hex chars>".
func randomTestDBName() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "test_" + hex.EncodeToString(b[:]), nil
}

// replaceDBName returns a copy of dsn with the database name replaced by newDB.
// Only URI-style DSNs (postgres://...) are supported; that is the format used
// throughout this project.
func replaceDBName(dsn, newDB string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse DSN: %w", err)
	}
	if u.Path == "" || u.Path == "/" {
		return "", fmt.Errorf("DSN has no database name")
	}
	u.Path = "/" + newDB
	return u.String(), nil
}

// createDatabase connects to the "postgres" maintenance database and issues
// CREATE DATABASE for the given db name.
func createDatabase(ctx context.Context, baseDSN, dbName string) error {
	maint := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(baseDSN),
		pgdriver.WithDatabase("postgres"),
	))
	defer maint.Close() //nolint:errcheck

	quotedName := `"` + strings.ReplaceAll(dbName, `"`, `""`) + `"`
	if _, err := maint.ExecContext(ctx, "CREATE DATABASE "+quotedName); err != nil {
		return err
	}
	return nil
}

// dropDatabase connects to the "postgres" maintenance database and issues
// DROP DATABASE for the given db name. Errors are returned but not fatal by
// design — this is always used in a best-effort cleanup path.
func dropDatabase(ctx context.Context, baseDSN, dbName string) error {
	maint := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(baseDSN),
		pgdriver.WithDatabase("postgres"),
	))
	defer maint.Close() //nolint:errcheck

	quotedName := `"` + strings.ReplaceAll(dbName, `"`, `""`) + `"`
	if _, err := maint.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// openBunDB opens a *bun.DB pointed at the given DSN (no pool tuning needed
// for ephemeral test databases).
func openBunDB(dsn string) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	return bun.NewDB(sqldb, pgdialect.New()), nil
}
