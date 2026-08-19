package database

import (
	"path/filepath"
	"testing"
)

// TestMigration verifies the Goose migration chain runs cleanly and that the
// chat_messages migration (00007) can be applied and rolled back.
func TestMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migrate.sqlite3")

	db, err := NewDB(path)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	assertTableExists(t, db, "chat_messages")

	var version int64
	if err := db.QueryRow(
		`SELECT version_id FROM goose_db_version ORDER BY id DESC LIMIT 1`,
	).Scan(&version); err != nil {
		t.Fatalf("read goose version: %v", err)
	}
	if version != 9 {
		t.Errorf("expected latest migration 9, got %d", version)
	}

	// whatsapp config keys seeded
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM config_settings WHERE key LIKE 'whatsapp_%'`,
	).Scan(&n); err != nil {
		t.Fatalf("count whatsapp config: %v", err)
	}
	if n < 4 {
		t.Errorf("expected whatsapp config keys seeded, got %d", n)
	}

	// lead notes column present
	var notesCol int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('leads') WHERE name='notes'`,
	).Scan(&notesCol); err != nil {
		t.Fatalf("check notes column: %v", err)
	}
	if notesCol != 1 {
		t.Errorf("expected notes column on leads, got %d", notesCol)
	}

	// Roll back to just before chat_messages (00007): first down drops the
	// lead notes (00009), second drops the whatsapp config seed (00008),
	// third drops chat_messages (00007).
	for i := 0; i < 3; i++ {
		if err := MigrateDown(db); err != nil {
			t.Fatalf("MigrateDown: %v", err)
		}
	}
	assertTableMissing(t, db, "chat_messages")

	// Re-apply so the final DB state is fully migrated.
	if err := MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp re-apply: %v", err)
	}
	assertTableExists(t, db, "chat_messages")
}
