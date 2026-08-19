package database

import (
	"context"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

func TestPaymentRepository_CreateAndGet(t *testing.T) {
	db := setupDB(t)
	repo := NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{
		ID: "PAY-001", LeadID: "L1", Amount: 50000, Type: "sinal",
		Status: "pending", Description: "Sinal casamento",
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, "PAY-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Amount != 50000 {
		t.Errorf("amount = %d, want 50000", got.Amount)
	}
	if got.Status != "pending" {
		t.Errorf("status = %s, want pending", got.Status)
	}
}

func TestPaymentRepository_UpdateStatus(t *testing.T) {
	db := setupDB(t)
	repo := NewPaymentRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.Payment{ID: "PAY-002", LeadID: "L1", Amount: 30000}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.UpdateStatus(ctx, "PAY-002", "confirmed", map[string]string{"confirmed_at": "2026-06-28"}); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, _ := repo.Get(ctx, "PAY-002")
	if got.Status != "confirmed" {
		t.Errorf("status = %s, want confirmed", got.Status)
	}
}

func TestPaymentRepository_GetByTransactionID(t *testing.T) {
	db := setupDB(t)
	repo := NewPaymentRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.Payment{ID: "PAY-003", LeadID: "L1", Amount: 10000, OpenPixTransactionID: "tx-abc"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByTransactionID(ctx, "tx-abc")
	if err != nil {
		t.Fatalf("GetByTransactionID: %v", err)
	}
	if got.ID != "PAY-003" {
		t.Errorf("id = %s, want PAY-003", got.ID)
	}
}

func TestPaymentRepository_ListByLead(t *testing.T) {
	db := setupDB(t)
	repo := NewPaymentRepository(db)
	ctx := context.Background()

	for _, id := range []string{"PAY-004", "PAY-005"} {
		if err := repo.Create(ctx, &domain.Payment{ID: id, LeadID: "L2", Amount: 10000}); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	payments, err := repo.ListByLead(ctx, "L2")
	if err != nil {
		t.Fatalf("ListByLead: %v", err)
	}
	if len(payments) != 2 {
		t.Errorf("got %d payments, want 2", len(payments))
	}
}
