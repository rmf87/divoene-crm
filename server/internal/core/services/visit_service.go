package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// VisitService handles visit booking and feedback business logic.
type VisitService struct {
	repo domain.VisitRepository
}

// NewVisitService creates a new VisitService.
func NewVisitService(repo domain.VisitRepository) *VisitService {
	return &VisitService{repo: repo}
}

// HasPersistence reports whether the service has a working database connection.
func (s *VisitService) HasPersistence() bool { return s.repo != nil }

// Create books a new visit.
func (s *VisitService) Create(ctx context.Context, actor domain.Actor, req domain.CreateVisitRequest) (*domain.Visit, error) {
	if err := validateVisitCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	visit := &domain.Visit{
		ID:        fmt.Sprintf("VIS-%d", time.Now().UnixNano()),
		LeadID:    req.LeadID,
		LeadName:  req.LeadName,
		GuideID:   req.GuideID,
		GuideName: req.GuideName,
		Date:      req.Date,
		TimeSlot:  req.TimeSlot,
		Status:    "scheduled",
		Product:   req.Product,
		WhatsApp:  req.WhatsApp,
		Notes:     req.Notes,
		CreatedAt: now,
		CreatedBy: actor.UID,
	}

	if s.repo != nil {
		if err := s.repo.Create(ctx, visit); err != nil {
			return nil, fmt.Errorf("persist: %w", err)
		}
	}

	return visit, nil
}

// Get retrieves a visit by ID.
func (s *VisitService) Get(ctx context.Context, actor domain.Actor, id string) (*domain.Visit, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("visita não encontrada")
	}
	v, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.HasRole("manager") && !actor.HasRole("seller") && v.GuideID != actor.UID {
		return nil, fmt.Errorf("visita não encontrada")
	}
	return v, nil
}

// Update changes visit status or records feedback.
func (s *VisitService) Update(ctx context.Context, actor domain.Actor, id string, req domain.UpdateVisitRequest) (*domain.Visit, error) {
	if s.repo == nil {
		return &domain.Visit{ID: id, Status: req.Status, Feedback: req.Feedback}, nil
	}

	v, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Feedback != nil && !actor.HasRole("manager") && v.GuideID != actor.UID {
		return nil, fmt.Errorf("apenas o guia designado pode registrar feedback")
	}

	if err := validateVisitUpdateRequest(req, v); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}

	return s.repo.Get(ctx, id)
}

// ListSlots returns available guide slots for a date.
func (s *VisitService) ListSlots(ctx context.Context, date string) ([]*domain.GuideSlot, error) {
	if s.repo == nil {
		return defaultVisitSlots(date), nil
	}
	return s.repo.ListAvailableSlots(ctx, date)
}

// ListSlotsForWeek returns all available slots for an ISO week (Mon-Sun).
func (s *VisitService) ListSlotsForWeek(ctx context.Context, year, week int) ([]*domain.GuideSlot, error) {
	dates := isoWeekDates(year, week)
	if s.repo == nil {
		var all []*domain.GuideSlot
		for _, d := range dates {
			all = append(all, defaultVisitSlots(d)...)
		}
		return all, nil
	}
	return s.repo.ListAvailableSlotsForWeek(ctx, dates)
}

func isoWeekDates(year, week int) []string {
	t := time.Date(year, 1, 4, 12, 0, 0, 0, time.UTC)
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	t = t.AddDate(0, 0, (week-1)*7)
	dates := make([]string, 7)
	for i := 0; i < 7; i++ {
		dates[i] = t.AddDate(0, 0, i).Format("2006-01-02")
	}
	return dates
}

// Guide operations

func (s *VisitService) ListGuides(ctx context.Context) ([]*domain.Guide, error) {
	if s.repo == nil {
		return defaultVisitGuides(), nil
	}
	return s.repo.ListGuides(ctx)
}

func (s *VisitService) GetGuide(ctx context.Context, id string) (*domain.Guide, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("guia não encontrado")
	}
	return s.repo.GetGuide(ctx, id)
}

func (s *VisitService) CreateGuide(ctx context.Context, g *domain.Guide) error {
	if s.repo == nil {
		return fmt.Errorf("sem conexão com banco")
	}
	return s.repo.CreateGuide(ctx, g)
}

func (s *VisitService) UpdateGuide(ctx context.Context, id string, g *domain.Guide) error {
	if s.repo == nil {
		return fmt.Errorf("sem conexão com banco")
	}
	return s.repo.UpdateGuide(ctx, id, g)
}

func (s *VisitService) UpdateGuideAvailability(ctx context.Context, id string, req domain.UpdateAvailabilityRequest) error {
	if s.repo == nil {
		return fmt.Errorf("sem conexão com banco")
	}
	return s.repo.UpdateGuideAvailability(ctx, id, req)
}

func (s *VisitService) ListByGuide(ctx context.Context, guideID string) ([]*domain.Visit, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.ListByGuide(ctx, guideID)
}

func (s *VisitService) ListByLead(ctx context.Context, leadID string) ([]*domain.Visit, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.ListByLead(ctx, leadID)
}

func (s *VisitService) ListAll(ctx context.Context) ([]*domain.Visit, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.ListAll(ctx)
}

var validVisitTimeSlots = map[string]bool{
	"08:00": true, "09:00": true, "10:00": true, "11:00": true,
	"12:00": true, "13:00": true, "14:00": true, "15:00": true,
	"16:00": true, "17:00": true,
}

func validateVisitCreateRequest(req domain.CreateVisitRequest) error {
	req.LeadID = strings.TrimSpace(req.LeadID)
	req.LeadName = strings.TrimSpace(req.LeadName)
	req.GuideID = strings.TrimSpace(req.GuideID)
	req.Date = strings.TrimSpace(req.Date)
	req.TimeSlot = strings.TrimSpace(req.TimeSlot)

	if req.LeadID == "" {
		return fmt.Errorf("lead_id é obrigatório")
	}
	if req.LeadName == "" {
		return fmt.Errorf("nome do lead é obrigatório")
	}
	if req.GuideID == "" {
		return fmt.Errorf("guide_id é obrigatório")
	}
	if req.Date == "" {
		return fmt.Errorf("data é obrigatória")
	}
	if req.TimeSlot == "" || !validVisitTimeSlots[req.TimeSlot] {
		return fmt.Errorf("horário inválido: %s (use HH:00, 08:00-17:00)", req.TimeSlot)
	}
	return nil
}

func validateVisitUpdateRequest(req domain.UpdateVisitRequest, current *domain.Visit) error {
	if req.Status != "" {
		validStatuses := map[string]bool{"confirmed": true, "done": true, "cancelled": true}
		if !validStatuses[req.Status] {
			return fmt.Errorf("status inválido: %s", req.Status)
		}
	}
	if req.Feedback != nil {
		if current.Feedback != nil && current.Status == "done" {
			return fmt.Errorf("feedback já registrado e não pode ser alterado")
		}
		if req.Feedback.Result == "" {
			return fmt.Errorf("resultado do feedback é obrigatório (liked, disliked, maybe)")
		}
		validResults := map[string]bool{"liked": true, "disliked": true, "maybe": true}
		if !validResults[req.Feedback.Result] {
			return fmt.Errorf("resultado inválido: %s (use liked, disliked, maybe)", req.Feedback.Result)
		}
	}
	return nil
}

func defaultVisitGuides() []*domain.Guide {
	return []*domain.Guide{
		{ID: "dev-guide-uid", Name: "Guia (dev)", Active: true, MaxPerSlot: 3, WeeklySchedule: map[string][]string{
			"mon": {"09:00", "10:00", "14:00", "15:00"},
			"tue": {"09:00", "10:00"},
			"wed": {"09:00", "14:00"},
			"thu": {"10:00", "14:00"},
			"fri": {"09:00", "10:00", "14:00"},
		}},
	}
}

func defaultVisitSlots(date string) []*domain.GuideSlot {
	return []*domain.GuideSlot{
		{GuideID: "dev-guide-uid", Date: date, TimeSlot: "09:00", Booked: 0, MaxSlots: 3},
		{GuideID: "dev-guide-uid", Date: date, TimeSlot: "10:00", Booked: 1, MaxSlots: 3},
		{GuideID: "dev-guide-uid", Date: date, TimeSlot: "14:00", Booked: 0, MaxSlots: 3},
		{GuideID: "dev-guide-uid", Date: date, TimeSlot: "15:00", Booked: 2, MaxSlots: 3},
	}
}
