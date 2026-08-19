package database

import (
	"context"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

func TestChatRepository_Crud(t *testing.T) {
	db := setupDB(t)
	leadRepo := NewLeadRepository(db)
	repo := NewChatMessageRepository(db)
	ctx := context.Background()

	if err := leadRepo.Store(ctx, &domain.Lead{
		ID: "L-1", Name: "Maria", WhatsApp: "11999998888", Product: "buffet_infantil", Stage: "lead",
	}); err != nil {
		t.Fatalf("seed lead: %v", err)
	}

	msg := &domain.ChatMessage{
		LeadID:      "L-1",
		WAMessageID: "wamid.1",
		Direction:   domain.DirectionSeller,
		Body:        "Olá!",
		Status:      domain.MessageStatusSent,
	}
	if err := repo.Store(ctx, msg); err != nil {
		t.Fatalf("store: %v", err)
	}
	if msg.ID == 0 {
		t.Error("expected autoincrement id")
	}

	list, err := repo.ListByLead(ctx, "L-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Body != "Olá!" {
		t.Fatalf("unexpected list: %+v", list)
	}

	got, err := repo.FindByWAMessageID(ctx, "wamid.1")
	if err != nil {
		t.Fatalf("find by wa id: %v", err)
	}
	if got.LeadID != "L-1" {
		t.Errorf("lead id: %q", got.LeadID)
	}

	if err := repo.UpdateStatus(ctx, "wamid.1", domain.MessageStatusRead); err != nil {
		t.Fatalf("update status: %v", err)
	}
	got, _ = repo.FindByWAMessageID(ctx, "wamid.1")
	if got.Status != domain.MessageStatusRead {
		t.Errorf("status: %q", got.Status)
	}
}

func TestChatRepository_ListOtherLead(t *testing.T) {
	db := setupDB(t)
	repo := NewChatMessageRepository(db)
	ctx := context.Background()

	list, err := repo.ListByLead(ctx, "NOPE")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %+v", list)
	}
}
