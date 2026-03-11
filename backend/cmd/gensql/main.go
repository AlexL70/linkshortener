// gensql generates a deployable PostgreSQL SQL script for each registered
// migration and a combined schema.sql containing all of them in order.
//
// It is invoked automatically by "go generate ./..." (via the directive in
// backend/main.go) and should be re-run whenever a new migration is added.
//
// Output is written to the sql/ directory at the repository root.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/migrations"
)

func main() {
	outDir := flag.String("out", "../sql", "output directory for generated SQL files")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "gensql: create output dir: %v\n", err)
		os.Exit(1)
	}

	date := time.Now().UTC().Format("2006-01-02")

	for _, script := range migrations.SQLScripts {
		if err := writeMigration(*outDir, script, date); err != nil {
			fmt.Fprintf(os.Stderr, "gensql: %v\n", err)
			os.Exit(1)
		}
	}

	if err := writeFullSchema(*outDir, date); err != nil {
		fmt.Fprintf(os.Stderr, "gensql: write schema.sql: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("gensql: wrote %d migration file(s) to %s/\n", len(migrations.SQLScripts), *outDir)
}

// writeMigration writes a single migration's SQL to <outDir>/<name>.sql.
func writeMigration(outDir string, script migrations.SQLScript, date string) error {
	path := filepath.Join(outDir, script.Name+".sql")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck -- write-only file, errors are surfaced above

	fmt.Fprintf(f, "-- Migration: %s\n", script.Name)
	fmt.Fprintf(f, "-- Generated: %s\n", date)
	fmt.Fprintf(f, "-- Apply with: psql \"$DATABASE_URL\" -f sql/%s.sql\n", script.Name)
	fmt.Fprintln(f)
	fmt.Fprintln(f, "BEGIN;")
	for _, stmt := range script.Up {
		fmt.Fprintf(f, "\n%s;\n", strings.TrimRight(strings.TrimSpace(stmt), ";"))
	}
	fmt.Fprintln(f)
	fmt.Fprintln(f, "COMMIT;")
	return nil
}

// writeFullSchema writes all migrations in order to <outDir>/schema.sql.
// Use this to initialise a brand-new database in one step.
func writeFullSchema(outDir string, date string) error {
	path := filepath.Join(outDir, "schema.sql")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck -- write-only file, errors are surfaced above

	fmt.Fprintf(f, "-- Full schema: all migrations applied in order\n")
	fmt.Fprintf(f, "-- Generated: %s\n", date)
	fmt.Fprintf(f, "-- Apply with: psql \"$DATABASE_URL\" -f sql/schema.sql\n")

	for _, script := range migrations.SQLScripts {
		fmt.Fprintf(f, "\n-- ====== Migration: %s ======\n", script.Name)
		fmt.Fprintln(f)
		fmt.Fprintln(f, "BEGIN;")
		for _, stmt := range script.Up {
			fmt.Fprintf(f, "\n%s;\n", strings.TrimRight(strings.TrimSpace(stmt), ";"))
		}
		fmt.Fprintln(f)
		fmt.Fprintln(f, "COMMIT;")
	}
	return nil
}
