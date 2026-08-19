package database

import (
	"context"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

func TestVisitRepository_CreateAndGet(t *testing.T) {
	db := setupDB(t)
	repo := NewVisitRepository(db)
	ctx := context.Background()

	v := &domain.Visit{
		ID: "V-001", LeadID: "L1", GuideID: "G1",
		Date: "2026-07-15", TimeSlot: "10:00",
	}
	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, "V-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Date != "2026-07-15" {
		t.Errorf("date = %s, want 2026-07-15", got.Date)
	}
}

func TestVisitRepository_Update(t *testing.T) {
	db := setupDB(t)
	repo := NewVisitRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.Visit{ID: "V-002", LeadID: "L1", GuideID: "G1", Date: "2026-07-15", TimeSlot: "10:00"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Update via the repo method
	upd := domain.UpdateVisitRequest{Status: "confirmed"}
	if err := repo.Update(ctx, "V-002", upd); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.Get(ctx, "V-002")
	if got.Status != "confirmed" {
		t.Errorf("status = %s, want confirmed", got.Status)
	}
}

func TestGuideRepository_CreateAndGet(t *testing.T) {
	db := setupDB(t)
	repo := NewVisitRepository(db)
	ctx := context.Background()

	g := &domain.Guide{ID: "G-001", Name: "João Guia", Email: "joao@divoene.com.br"}
	if err := repo.CreateGuide(ctx, g); err != nil {
		t.Fatalf("CreateGuide: %v", err)
	}

	got, err := repo.GetGuide(ctx, "G-001")
	if err != nil {
		t.Fatalf("GetGuide: %v", err)
	}
	if got.Name != "João Guia" {
		t.Errorf("name = %s, want João Guia", got.Name)
	}
}
