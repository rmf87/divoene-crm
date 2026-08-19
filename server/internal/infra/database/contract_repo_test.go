package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

func TestContractRepository_Create(t *testing.T) {
	db := setupDB(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	contract := &domain.Contract{
		ID:      "CT-001",
		LeadID:  "LEAD-001",
		Amount:  150000,
		Product: "buffet_infantil",
		Status:  "sent",
	}
	if err := repo.Create(ctx, contract); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, "CT-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Amount != 150000 || got.Status != "sent" {
		t.Errorf("got %+v", got)
	}
}

func TestContractRepository_GetByDocKey(t *testing.T) {
	db := setupDB(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	contract := &domain.Contract{
		ID:              "CT-DOCKEY",
		LeadID:          "LEAD-002",
		ClicksignDocKey: "doc-key-123",
		Amount:          200000,
		Product:         "ensaio_fotografico",
		Status:          "sent",
	}
	_ = repo.Create(ctx, contract)

	got, err := repo.GetByDocKey(ctx, "doc-key-123")
	if err != nil {
		t.Fatalf("GetByDocKey: %v", err)
	}
	if got.ID != "CT-DOCKEY" {
		t.Errorf("got %q, want %q", got.ID, "CT-DOCKEY")
	}
}

func TestContractRepository_UpdateStatus(t *testing.T) {
	db := setupDB(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Contract{
		ID: "CT-STATUS", LeadID: "LEAD-003",
		Amount: 100000, Product: "corporativo", Status: "sent",
	})

	now := time.Now().UTC().Format(time.RFC3339)
	extra := map[string]string{"signed_at": now}
	if err := repo.UpdateStatus(ctx, "CT-STATUS", "signed", extra); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.Get(ctx, "CT-STATUS")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != "signed" || got.SignedAt != now {
		t.Errorf("update failed: status=%s signed_at=%s", got.Status, got.SignedAt)
	}
}

func TestContractRepository_ListByLead(t *testing.T) {
	db := setupDB(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = repo.Create(ctx, &domain.Contract{
			ID: fmt.Sprintf("CT-LIST-%d", i), LeadID: "LEAD-010",
			Amount: int64(100000 * (i + 1)), Product: "casamentos",
			Status: "sent",
		})
	}

	contracts, err := repo.ListByLead(ctx, "LEAD-010")
	if err != nil {
		t.Fatalf("ListByLead: %v", err)
	}
	if len(contracts) != 3 {
		t.Errorf("got %d contracts, want 3", len(contracts))
	}
}

func TestContractRepository_ListAll(t *testing.T) {
	db := setupDB(t)
	repo := NewContractRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Contract{
		ID: "CT-A", LeadID: "L1", Amount: 100, Product: "site", Status: "sent",
	})
	_ = repo.Create(ctx, &domain.Contract{
		ID: "CT-B", LeadID: "L2", Amount: 200, Product: "site", Status: "signed",
	})

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("got %d, want 2", len(all))
	}
}
