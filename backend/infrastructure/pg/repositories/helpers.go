package repositories

import (
	"database/sql"
	"errors"
	"strings"
)

// isUniqueViolation reports whether err originates from a PostgreSQL
// unique-constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint")
}

// isNotFound reports whether err is a "no rows" / not-found error from the database driver.
func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
