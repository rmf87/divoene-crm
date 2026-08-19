package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB opens a SQLite database with WAL mode and production-friendly settings.
func NewDB(dbPath string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=30000&_foreign_keys=ON&_cache_size=-8000&_temp_store=MEMORY",
		dbPath,
	)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

// RunMigrations applies all pending Goose migrations.
func RunMigrations(db *sql.DB) error {
	if err := MigrateUp(db); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Printf("[migrations] all migrations applied")
	return nil
}
