package database

import (
	"context"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

func setupTestDB(t *testing.T) *LeadRepository {
	t.Helper()
	db := setupDB(t)
	repo := NewLeadRepository(db)
	return repo
}

func TestLeadRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	lead := &domain.Lead{
		ID:       "TEST-001",
		Name:     "Test Lead",
		WhatsApp: "11999998888",
		Product:  "buffet_infantil",
		Source:   "site",
		Stage:    "lead",
	}
	if err := repo.Store(ctx, lead); err != nil {
		t.Fatalf("Store: %v", err)
	}

	got, err := repo.Get(ctx, "TEST-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Test Lead" {
		t.Errorf("got name %q, want %q", got.Name, "Test Lead")
	}
	if got.Stage != "lead" {
		t.Errorf("got stage %q, want %q", got.Stage, "lead")
	}
}

func TestLeadRepository_Get_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	_, err := repo.Get(context.Background(), "DOES-NOT-EXIST")
	if err == nil {
		t.Error("expected error for non-existent lead")
	}
}

func TestLeadRepository_List(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	leads := []*domain.Lead{
		{ID: "L1", Name: "Alice", Product: "buffet_infantil", Stage: "lead", WhatsApp: "11999998881"},
		{ID: "L2", Name: "Bob", Product: "ensaio_fotografico", Stage: "validated", WhatsApp: "11999998882"},
		{ID: "L3", Name: "Charlie", Product: "buffet_infantil", Stage: "contract", WhatsApp: "11999998883"},
	}
	for _, l := range leads {
		if err := repo.Store(ctx, l); err != nil {
			t.Fatalf("Store %s: %v", l.ID, err)
		}
	}

	all, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("got %d leads, want 3", len(all))
	}

	filtered, err := repo.List(ctx, map[string]string{"stage": "lead"})
	if err != nil {
		t.Fatalf("List filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "Alice" {
		t.Errorf("got %d leads for stage=lead, want 1 (Alice)", len(filtered))
	}
}

func TestLeadRepository_UpdateStage(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	lead := &domain.Lead{
		ID: "UPDATE-001", Name: "Update Test", Product: "corporativo",
		Stage: "lead", WhatsApp: "11999998884",
	}
	if err := repo.Store(ctx, lead); err != nil {
		t.Fatalf("Store: %v", err)
	}

	if err := repo.UpdateStage(ctx, "UPDATE-001", "validated", "seller-1"); err != nil {
		t.Fatalf("UpdateStage: %v", err)
	}

	got, err := repo.Get(ctx, "UPDATE-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Stage != "validated" {
		t.Errorf("got stage %q, want %q", got.Stage, "validated")
	}
}

func TestLeadRepository_UpdateValidation(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	lead := &domain.Lead{
		ID: "VAL-001", Name: "Val Test", Product: "casamentos",
		Stage: "validated", WhatsApp: "11999998885",
	}
	if err := repo.Store(ctx, lead); err != nil {
		t.Fatalf("Store: %v", err)
	}

	event := &domain.EventInfo{
		EventType:          "casamento",
		PossibleDates:      []string{"2026-08-15"},
		EstimatedPeople:    100,
		DesiredDurationHrs: 6,
	}
	contact := &domain.ContactPersonInfo{
		Name:     "Noiva",
		WhatsApp: "11999998886",
		Role:     "noiva",
	}
	addOns := []domain.AddOnItem{
		{Name: "DJ", Quantity: 1},
	}

	if err := repo.UpdateValidation(ctx, "VAL-001", event, contact, addOns); err != nil {
		t.Fatalf("UpdateValidation: %v", err)
	}

	got, err := repo.Get(ctx, "VAL-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Event == nil || got.Event.EventType != "casamento" {
		t.Error("event not updated correctly")
	}
	if got.ContactPerson == nil || got.ContactPerson.Name != "Noiva" {
		t.Error("contact not updated correctly")
	}
	if len(got.AddOns) != 1 || got.AddOns[0].Name != "DJ" {
		t.Error("add-ons not updated correctly")
	}
}

func TestLeadRepository_FindByWhatsApp(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	if err := repo.Store(ctx, &domain.Lead{
		ID: "WA-001", Name: "Maria", WhatsApp: "11999998888", Product: "buffet_infantil", Stage: "lead",
	}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	got, err := repo.FindByWhatsApp(ctx, []string{"5511999998888", "11999998888"})
	if err != nil {
		t.Fatalf("FindByWhatsApp: %v", err)
	}
	if got.ID != "WA-001" {
		t.Errorf("got lead %q", got.ID)
	}

	if _, err := repo.FindByWhatsApp(ctx, []string{"55115550000"}); err == nil {
		t.Error("expected not found for unknown number")
	}
}

func TestLeadRepository_AddNote(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	if err := repo.Store(ctx, &domain.Lead{
		ID: "NOTE-001", Name: "Maria", WhatsApp: "11999998888", Product: "buffet_infantil", Stage: "lead",
	}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	if err := repo.AddNote(ctx, "NOTE-001", domain.Note{Text: "Cliente pediu orçamento", CreatedBy: "Vendedor", CreatedAt: "2026-08-15T00:00:00Z"}); err != nil {
		t.Fatalf("AddNote: %v", err)
	}

	got, err := repo.Get(ctx, "NOTE-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Notes) != 1 || got.Notes[0].Text != "Cliente pediu orçamento" {
		t.Fatalf("notes not persisted: %+v", got.Notes)
	}

	// Second note appends.
	if err := repo.AddNote(ctx, "NOTE-001", domain.Note{Text: "Enviar contrato", CreatedBy: "Vendedor", CreatedAt: "2026-08-15T01:00:00Z"}); err != nil {
		t.Fatalf("AddNote 2nd: %v", err)
	}
	got, _ = repo.Get(ctx, "NOTE-001")
	if len(got.Notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(got.Notes))
	}
}
