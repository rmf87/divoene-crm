package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// VisitRepository implements domain.VisitRepository using SQLite.
type VisitRepository struct {
	db *sql.DB
}

// NewVisitRepository creates a new VisitRepository.
func NewVisitRepository(db *sql.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

// Visit operations

func (r *VisitRepository) Create(ctx context.Context, v *domain.Visit) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO visits (id, lead_id, lead_name, guide_id, guide_name, date, time_slot,
		 status, product, whatsapp, notes, created_at, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.LeadID, v.LeadName, v.GuideID, v.GuideName, v.Date, v.TimeSlot,
		v.Status, v.Product, v.WhatsApp, v.Notes, v.CreatedAt, v.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert visit: %w", err)
	}
	return nil
}

func (r *VisitRepository) Get(ctx context.Context, id string) (*domain.Visit, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lead_id, lead_name, guide_id, guide_name, date, time_slot,
		 status, product, whatsapp, notes, feedback, created_at, created_by
		 FROM visits WHERE id = ?`, id)

	v := &domain.Visit{}
	var feedbackRaw string
	err := row.Scan(&v.ID, &v.LeadID, &v.LeadName, &v.GuideID, &v.GuideName, &v.Date, &v.TimeSlot,
		&v.Status, &v.Product, &v.WhatsApp, &v.Notes, &feedbackRaw, &v.CreatedAt, &v.CreatedBy)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("visita não encontrada")
	}
	if err != nil {
		return nil, fmt.Errorf("scan visit: %w", err)
	}
	if feedbackRaw != "" {
		unmarshalJSON(feedbackRaw, &v.Feedback)
	}
	return v, nil
}

func (r *VisitRepository) Update(ctx context.Context, id string, update domain.UpdateVisitRequest) error {
	query := `UPDATE visits SET `
	args := []interface{}{}
	parts := []string{}

	if update.Status != "" {
		parts = append(parts, `status = ?`)
		args = append(args, update.Status)
	}
	if update.Feedback != nil {
		parts = append(parts, `feedback = ?`)
		args = append(args, marshalJSON(update.Feedback))
	}
	if update.GuideID != "" {
		parts = append(parts, `guide_id = ?`)
		args = append(args, update.GuideID)
	}
	if update.GuideName != "" {
		parts = append(parts, `guide_name = ?`)
		args = append(args, update.GuideName)
	}
	if update.Date != "" {
		parts = append(parts, `date = ?`)
		args = append(args, update.Date)
	}
	if update.TimeSlot != "" {
		parts = append(parts, `time_slot = ?`)
		args = append(args, update.TimeSlot)
	}

	if len(parts) == 0 {
		return nil
	}

	for i, p := range parts {
		if i > 0 {
			query += ", "
		}
		query += p
	}
	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update visit: %w", err)
	}
	return nil
}

func (r *VisitRepository) ListByGuide(ctx context.Context, guideID string) ([]*domain.Visit, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, lead_name, guide_id, guide_name, date, time_slot,
		 status, product, whatsapp, notes, feedback, created_at, created_by
		 FROM visits WHERE guide_id = ? ORDER BY date DESC, time_slot DESC`, guideID)
	if err != nil {
		return nil, fmt.Errorf("list visits by guide: %w", err)
	}
	defer rows.Close()
	return scanVisits(rows)
}

func (r *VisitRepository) ListByLead(ctx context.Context, leadID string) ([]*domain.Visit, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, lead_name, guide_id, guide_name, date, time_slot,
		 status, product, whatsapp, notes, feedback, created_at, created_by
		 FROM visits WHERE lead_id = ? ORDER BY date DESC`, leadID)
	if err != nil {
		return nil, fmt.Errorf("list visits by lead: %w", err)
	}
	defer rows.Close()
	return scanVisits(rows)
}

func (r *VisitRepository) ListAll(ctx context.Context) ([]*domain.Visit, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, lead_name, guide_id, guide_name, date, time_slot,
		 status, product, whatsapp, notes, feedback, created_at, created_by
		 FROM visits ORDER BY date DESC, time_slot DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all visits: %w", err)
	}
	defer rows.Close()
	return scanVisits(rows)
}

func (r *VisitRepository) ListSlots(ctx context.Context, date string) ([]*domain.GuideSlot, error) {
	return r.ListAvailableSlots(ctx, date)
}

func (r *VisitRepository) CountBySlot(ctx context.Context, guideID, date, timeSlot string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM visits WHERE guide_id = ? AND date = ? AND time_slot = ? AND status != 'cancelled'`,
		guideID, date, timeSlot).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by slot: %w", err)
	}
	return count, nil
}

func (r *VisitRepository) ListAvailableSlots(ctx context.Context, date string) ([]*domain.GuideSlot, error) {
	guides, err := r.ListGuides(ctx)
	if err != nil {
		return nil, err
	}

	var slots []*domain.GuideSlot
	for _, g := range guides {
		if !g.Active {
			continue
		}
		if isDateUnavailable(g.UnavailableDates, date) {
			continue
		}
		dayName := weekdayName(date)
		weekSlots, ok := g.WeeklySchedule[dayName]
		if !ok {
			continue
		}
		for _, ts := range weekSlots {
			booked, _ := r.CountBySlot(ctx, g.ID, date, ts)
			slots = append(slots, &domain.GuideSlot{
				GuideID:   g.ID,
				GuideName: g.Name,
				Date:      date,
				TimeSlot:  ts,
				Booked:    booked,
				MaxSlots:  g.MaxPerSlot,
			})
		}
	}
	if slots == nil {
		slots = []*domain.GuideSlot{}
	}
	return slots, nil
}

func (r *VisitRepository) ListAvailableSlotsForWeek(ctx context.Context, dates []string) ([]*domain.GuideSlot, error) {
	var all []*domain.GuideSlot
	for _, d := range dates {
		slots, err := r.ListAvailableSlots(ctx, d)
		if err != nil {
			return nil, err
		}
		all = append(all, slots...)
	}
	return all, nil
}

// Guide operations

func (r *VisitRepository) ListGuides(ctx context.Context) ([]*domain.Guide, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, active, weekly_schedule, unavailable_dates, max_per_slot
		 FROM guides ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list guides: %w", err)
	}
	defer rows.Close()

	var guides []*domain.Guide
	for rows.Next() {
		g := &domain.Guide{}
		var weeklyRaw, unavailRaw string
		var active int
		if err := rows.Scan(&g.ID, &g.Name, &g.Email, &active, &weeklyRaw, &unavailRaw, &g.MaxPerSlot); err != nil {
			return nil, fmt.Errorf("scan guide: %w", err)
		}
		g.Active = active == 1
		unmarshalJSON(weeklyRaw, &g.WeeklySchedule)
		unmarshalJSON(unavailRaw, &g.UnavailableDates)
		guides = append(guides, g)
	}
	return guides, rows.Err()
}

func (r *VisitRepository) GetGuide(ctx context.Context, id string) (*domain.Guide, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, active, weekly_schedule, unavailable_dates, max_per_slot
		 FROM guides WHERE id = ?`, id)

	g := &domain.Guide{}
	var weeklyRaw, unavailRaw string
	var active int
	err := row.Scan(&g.ID, &g.Name, &g.Email, &active, &weeklyRaw, &unavailRaw, &g.MaxPerSlot)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("guia não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan guide: %w", err)
	}
	g.Active = active == 1
	unmarshalJSON(weeklyRaw, &g.WeeklySchedule)
	unmarshalJSON(unavailRaw, &g.UnavailableDates)
	return g, nil
}

func (r *VisitRepository) CreateGuide(ctx context.Context, g *domain.Guide) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO guides (id, name, email, active, weekly_schedule, unavailable_dates, max_per_slot)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		g.ID, g.Name, g.Email, boolToInt(g.Active),
		marshalJSON(g.WeeklySchedule), marshalJSON(g.UnavailableDates), g.MaxPerSlot)
	if err != nil {
		return fmt.Errorf("insert guide: %w", err)
	}
	return nil
}

func (r *VisitRepository) UpdateGuide(ctx context.Context, id string, g *domain.Guide) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE guides SET name = ?, email = ?, active = ?,
		 weekly_schedule = ?, unavailable_dates = ?, max_per_slot = ?
		 WHERE id = ?`,
		g.Name, g.Email, boolToInt(g.Active),
		marshalJSON(g.WeeklySchedule), marshalJSON(g.UnavailableDates), g.MaxPerSlot, id)
	if err != nil {
		return fmt.Errorf("update guide: %w", err)
	}
	return nil
}

func (r *VisitRepository) UpdateGuideAvailability(ctx context.Context, id string, req domain.UpdateAvailabilityRequest) error {
	query := `UPDATE guides SET `
	args := []interface{}{}
	parts := []string{}

	if req.WeeklySchedule != nil {
		parts = append(parts, `weekly_schedule = ?`)
		args = append(args, marshalJSON(req.WeeklySchedule))
	}
	if req.UnavailableDates != nil {
		parts = append(parts, `unavailable_dates = ?`)
		args = append(args, marshalJSON(req.UnavailableDates))
	}
	if req.MaxPerSlot != nil {
		parts = append(parts, `max_per_slot = ?`)
		args = append(args, *req.MaxPerSlot)
	}

	if len(parts) == 0 {
		return nil
	}

	for i, p := range parts {
		if i > 0 {
			query += ", "
		}
		query += p
	}
	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update guide availability: %w", err)
	}
	return nil
}

// helpers

func scanVisits(rows *sql.Rows) ([]*domain.Visit, error) {
	var visits []*domain.Visit
	for rows.Next() {
		v := &domain.Visit{}
		var feedbackRaw string
		if err := rows.Scan(&v.ID, &v.LeadID, &v.LeadName, &v.GuideID, &v.GuideName, &v.Date, &v.TimeSlot,
			&v.Status, &v.Product, &v.WhatsApp, &v.Notes, &feedbackRaw, &v.CreatedAt, &v.CreatedBy); err != nil {
			return nil, fmt.Errorf("scan visit row: %w", err)
		}
		if feedbackRaw != "" {
			unmarshalJSON(feedbackRaw, &v.Feedback)
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

func isDateUnavailable(ranges []domain.DateRange, date string) bool {
	for _, dr := range ranges {
		if date >= dr.From && date <= dr.To {
			return true
		}
	}
	return false
}

func weekdayName(dateStr string) string {
	// Parse YYYY-MM-DD and return abbreviated weekday
	layouts := []string{"2006-01-02", "2006-1-2", "2006-01-2", "2006-1-02"}
	var t interface{}
	_ = t // placeholder — handled by map approach
	days := map[string]string{
		"Monday": "mon", "Tuesday": "tue", "Wednesday": "wed",
		"Thursday": "thu", "Friday": "fri", "Saturday": "sat", "Sunday": "sun",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, dateStr)
		if err == nil {
			return days[parsed.Weekday().String()]
		}
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
