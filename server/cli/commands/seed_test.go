package commands

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rmf87/divoene/internal/infra/database"
)

func TestSeedDemoLeads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed.sqlite3")
	db, err := database.NewDB(path)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	if err := database.MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	repo := database.NewLeadRepository(db)
	ctx := context.Background()

	if err := seedDemoLeads(db); err != nil {
		t.Fatalf("seedDemoLeads: %v", err)
	}
	leads, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(leads) != 9 {
		t.Fatalf("expected 9 demo leads, got %d", len(leads))
	}

	// Idempotent: running again must not duplicate.
	if err := seedDemoLeads(db); err != nil {
		t.Fatalf("seedDemoLeads (2nd): %v", err)
	}
	leads2, _ := repo.List(ctx, nil)
	if len(leads2) != 9 {
		t.Fatalf("expected 9 leads after re-seed, got %d", len(leads2))
	}

	// L001 must be queryable (chat + pipeline use it by id).
	if _, err := repo.Get(ctx, "L001"); err != nil {
		t.Fatalf("L001 not found: %v", err)
	}
}
