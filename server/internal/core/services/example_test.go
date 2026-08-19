package services

import (
	"context"
	"errors"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

// mockLeadRepo implements domain.LeadRepository for testing.
type mockLeadRepo struct {
	leads map[string]*domain.Lead
}

func (m *mockLeadRepo) Store(ctx context.Context, lead *domain.Lead) error {
	if m.leads == nil {
		m.leads = make(map[string]*domain.Lead)
	}
	m.leads[lead.ID] = lead
	return nil
}
func (m *mockLeadRepo) Get(ctx context.Context, id string) (*domain.Lead, error) {
	if m.leads == nil {
		return nil, errors.New("not found")
	}
	l, ok := m.leads[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return l, nil
}
func (m *mockLeadRepo) List(ctx context.Context, filters map[string]string) ([]*domain.Lead, error) {
	if m.leads == nil {
		return nil, nil
	}
	var result []*domain.Lead
	for _, l := range m.leads {
		if stage, ok := filters["stage"]; ok && l.Stage != stage {
			continue
		}
		result = append(result, l)
	}
	return result, nil
}
func (m *mockLeadRepo) FindByWhatsApp(ctx context.Context, numbers []string) (*domain.Lead, error) {
	if m.leads == nil {
		return nil, errors.New("not found")
	}
	for _, l := range m.leads {
		for _, n := range numbers {
			if l.WhatsApp == n {
				return l, nil
			}
		}
	}
	return nil, errors.New("not found")
}
func (m *mockLeadRepo) UpdateStage(ctx context.Context, id, stage, changedBy string) error {
	l, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	if !domain.CanTransition(l.Stage, stage) {
		return errors.New("invalid transition")
	}
	l.Stage = stage
	return nil
}
func (m *mockLeadRepo) UpdateValidation(ctx context.Context, id string, event *domain.EventInfo, contact *domain.ContactPersonInfo, addOns []domain.AddOnItem) error {
	l, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	l.Event = event
	l.ContactPerson = contact
	l.AddOns = addOns
	return nil
}
func (m *mockLeadRepo) AddNote(ctx context.Context, id string, note domain.Note) error {
	l, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	l.Notes = append(l.Notes, note)
	return nil
}

func TestLeadService_Create(t *testing.T) {
	svc := NewLeadService(&mockLeadRepo{})
	lead, err := svc.Create(domain.CreateLeadRequest{
		Name: "Alice", WhatsApp: "11999998888", Product: "buffet_infantil",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if lead.ID == "" {
		t.Error("expected ID to be generated")
	}
	if lead.Stage != "lead" {
		t.Errorf("stage = %s, want lead", lead.Stage)
	}
}

func TestLeadService_HasPersistence(t *testing.T) {
	svc := NewLeadService(&mockLeadRepo{})
	if !svc.HasPersistence() {
		t.Error("HasPersistence should be true with non-nil repo")
	}
}

// mockContractRepo implements domain.ContractRepository for testing.
type mockContractRepo struct {
	contracts map[string]*domain.Contract
}

func (m *mockContractRepo) Create(ctx context.Context, c *domain.Contract) error {
	if m.contracts == nil {
		m.contracts = make(map[string]*domain.Contract)
	}
	m.contracts[c.ID] = c
	return nil
}
func (m *mockContractRepo) Get(ctx context.Context, id string) (*domain.Contract, error) {
	c, ok := m.contracts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}
func (m *mockContractRepo) GetByDocKey(ctx context.Context, docKey string) (*domain.Contract, error) {
	for _, c := range m.contracts {
		if c.ClicksignDocKey == docKey {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockContractRepo) ListByLead(ctx context.Context, leadID string) ([]*domain.Contract, error) {
	return nil, nil
}
func (m *mockContractRepo) ListAll(ctx context.Context) ([]*domain.Contract, error) {
	return nil, nil
}
func (m *mockContractRepo) UpdateStatus(ctx context.Context, id, status string, extra map[string]string) error {
	c, ok := m.contracts[id]
	if !ok {
		return errors.New("not found")
	}
	c.Status = status
	return nil
}

// mockClicksign implements domain.ClicksignSender for testing.
type mockClicksign struct{}

func (m *mockClicksign) CreateAndSend(contract *domain.CreateContractRequest) (*domain.Contract, error) {
	return &domain.Contract{
		ID: "CT-MOCK", LeadID: contract.LeadID, Amount: contract.Amount,
		Status: "sent", ClicksignDocKey: "doc-001",
	}, nil
}
func (m *mockClicksign) VerifyWebhookSignature(body []byte, sig string) bool { return true }
func (m *mockClicksign) IsMock() bool                                        { return true }
func (m *mockClicksign) GetWebhookSecret() string                            { return "" }
func (m *mockClicksign) PollStatus(documentKey string) (string, error)       { return "signed", nil }

func TestContractService_Create(t *testing.T) {
	svc := NewContractService(&mockContractRepo{}, &mockClicksign{})
	actor := domain.Actor{UID: "seller-1", Roles: []string{"seller"}}
	contract, err := svc.Create(context.Background(), actor, domain.CreateContractRequest{
		LeadID: "L1", LeadName: "Alice", LeadEmail: "alice@test.com",
		Amount: 50000, Product: "buffet_infantil", EventDate: "2026-07-15",
		PaymentCond: "50% sinal",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if contract.Status != "sent" {
		t.Errorf("status = %s, want sent", contract.Status)
	}
	if contract.SentBy != "seller-1" {
		t.Errorf("sent_by = %s, want seller-1", contract.SentBy)
	}
}

func TestContractService_RoleGate(t *testing.T) {
	svc := NewContractService(&mockContractRepo{}, &mockClicksign{})
	// Create contract as seller-1
	actor := domain.Actor{UID: "seller-1", Roles: []string{"seller"}}
	contract, _ := svc.Create(context.Background(), actor, domain.CreateContractRequest{
		LeadID: "L1", LeadName: "Alice", LeadEmail: "alice@test.com",
		Amount: 50000, Product: "buffet_infantil", EventDate: "2026-07-15",
		PaymentCond: "50% sinal",
	})

	// Different seller trying to access
	other := domain.Actor{UID: "seller-2", Roles: []string{"seller"}}
	_, err := svc.Get(context.Background(), other, contract.ID)
	if err == nil {
		t.Error("expected error for different seller")
	}

	// Manager can access
	mgr := domain.Actor{UID: "manager-1", Roles: []string{"manager"}}
	got, err := svc.Get(context.Background(), mgr, contract.ID)
	if err != nil {
		t.Errorf("manager should be able to access: %v", err)
	}
	if got == nil {
		t.Error("got nil contract")
	}
}

func TestLeadService_AddNote(t *testing.T) {
	svc := NewLeadService(&mockLeadRepo{})
	lead, err := svc.Create(domain.CreateLeadRequest{
		Name: "Alice", WhatsApp: "11999998888", Product: "buffet_infantil",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.AddNote(context.Background(), lead.ID, "  Quer data de manhã  ", "Vendedor")
	if err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	if len(updated.Notes) != 1 || updated.Notes[0].Text != "Quer data de manhã" {
		t.Fatalf("note not added: %+v", updated.Notes)
	}
	if _, err := svc.AddNote(context.Background(), lead.ID, "   ", "Vendedor"); err == nil {
		t.Fatal("expected error for empty note")
	}
}
