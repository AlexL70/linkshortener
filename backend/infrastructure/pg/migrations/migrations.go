package migrations

import "github.com/uptrace/bun/migrate"

// Migrations is the global migrator registry for bun/migrate.
// Each migration file registers itself via init().
var Migrations = migrate.NewMigrations()

// SQLScript holds the raw DDL statements for a single migration.
// It is used by cmd/gensql to produce deployable .sql files for production.
type SQLScript struct {
	Name string   // e.g. "001_initial_schema"
	Up   []string // DDL statements that apply the migration
	Down []string // DDL statements that roll the migration back
}

// SQLScripts is the ordered list of migration SQL scripts registered by each
// migration file's init() function. The order matches the migration order.
var SQLScripts []SQLScript
