package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// LeadService implements the business logic for leads.
type LeadService struct {
	repo domain.LeadRepository
}

// NewLeadService creates a new LeadService.
func NewLeadService(repo domain.LeadRepository) *LeadService {
	return &LeadService{repo: repo}
}

// HasPersistence reports whether the service has a working database connection.
func (s *LeadService) HasPersistence() bool { return s.repo != nil }

// Create persists a validated lead request.
func (s *LeadService) Create(req domain.CreateLeadRequest) (*domain.Lead, error) {
	lead := &domain.Lead{
		ID:          fmt.Sprintf("LEAD-%d", time.Now().UnixNano()),
		Name:        req.Name,
		WhatsApp:    req.WhatsApp,
		Product:     req.Product,
		DesiredDate: req.DesiredDate,
		Source:      req.Source,
		Stage:       "lead",
	}

	if s.repo != nil {
		if err := s.repo.Store(context.Background(), lead); err != nil {
			return nil, fmt.Errorf("persist: %w", err)
		}
	}

	return lead, nil
}

// Get retrieves a lead by ID.
func (s *LeadService) Get(ctx context.Context, id string) (*domain.Lead, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("sem conexão com banco")
	}
	return s.repo.Get(ctx, id)
}

// List returns all leads, optionally filtered.
func (s *LeadService) List(ctx context.Context, filters map[string]string) ([]*domain.Lead, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("sem conexão com banco")
	}
	return s.repo.List(ctx, filters)
}

// UpdateValidation persists event, contact person, and add-ons for a lead.
func (s *LeadService) UpdateValidation(ctx context.Context, id string, event *domain.EventInfo, contactPerson *domain.ContactPersonInfo, addOns []domain.AddOnItem) error {
	if s.repo == nil {
		return fmt.Errorf("sem conexão com banco")
	}
	return s.repo.UpdateValidation(ctx, id, event, contactPerson, addOns)
}

// AddNote appends a quick seller note to a lead and returns the updated lead.
func (s *LeadService) AddNote(ctx context.Context, id, text, createdBy string) (*domain.Lead, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("nota vazia")
	}
	if s.repo == nil {
		return nil, fmt.Errorf("sem conexão com banco")
	}
	note := domain.Note{Text: text, CreatedBy: createdBy, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := s.repo.AddNote(ctx, id, note); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

// UpdateStage advances the lead pipeline stage with validation and history.
func (s *LeadService) UpdateStage(ctx context.Context, id, stage, changedBy string) (*domain.Lead, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("sem conexão com banco")
	}

	lead, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanTransition(lead.Stage, stage) {
		return nil, fmt.Errorf("transição inválida: %s → %s", lead.Stage, stage)
	}

	if err := s.repo.UpdateStage(ctx, id, stage, changedBy); err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, id)
}
