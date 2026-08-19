package database

import (
	"database/sql"
	"testing"
)

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var n string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&n)
	if err != nil {
		t.Fatalf("table %s missing: %v", name, err)
	}
}

func assertTableMissing(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var n string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&n)
	if err != sql.ErrNoRows {
		t.Fatalf("table %s should be gone, got err=%v", name, err)
	}
}
