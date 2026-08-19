package database

import (
	"database/sql"
	"testing"
)

// setupDB creates an in-memory SQLite database with schema for testing.
func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB(:memory:): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}
	return db
}

// setupUserDB creates a UserRepository for testing.
func setupUserDB(t *testing.T) *UserRepository {
	t.Helper()
	db := setupDB(t)
	return NewUserRepository(db)
}
