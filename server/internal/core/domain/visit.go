package domain

import "context"

// Visit represents a scheduled visit.
type Visit struct {
	ID        string         `json:"id"`
	LeadID    string         `json:"lead_id"`
	LeadName  string         `json:"lead_name"`
	GuideID   string         `json:"guide_id"`
	GuideName string         `json:"guide_name,omitempty"`
	Date      string         `json:"date"`
	TimeSlot  string         `json:"time_slot"`
	Status    string         `json:"status"`
	Product   string         `json:"product,omitempty"`
	WhatsApp  string         `json:"whatsapp,omitempty"`
	Notes     string         `json:"notes,omitempty"`
	Feedback  *VisitFeedback `json:"feedback,omitempty"`
	CreatedAt string         `json:"created_at"`
	CreatedBy string         `json:"created_by"`
}

// VisitFeedback is recorded by the guide after the visit.
type VisitFeedback struct {
	Result    string `json:"result"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

// GuideSlot represents a guide's available time slot.
type GuideSlot struct {
	GuideID   string `json:"guide_id"`
	GuideName string `json:"guide_name,omitempty"`
	Date      string `json:"date"`
	TimeSlot  string `json:"time_slot"`
	Booked    int    `json:"booked"`
	MaxSlots  int    `json:"max_slots"`
}

// Guide represents a visit guide with availability schedule.
type Guide struct {
	ID               string              `json:"id,omitempty"`
	Name             string              `json:"name"`
	Email            string              `json:"email,omitempty"`
	Active           bool                `json:"active"`
	WeeklySchedule   map[string][]string `json:"weekly_schedule,omitempty"`
	UnavailableDates []DateRange         `json:"unavailable_dates,omitempty"`
	MaxPerSlot       int                 `json:"max_per_slot"`
	CreatedAt        string              `json:"created_at,omitempty"`
}

// DateRange is an inclusive date range for guide unavailability.
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CreateVisitRequest is the payload for POST /api/visits.
type CreateVisitRequest struct {
	LeadID    string `json:"lead_id"`
	LeadName  string `json:"lead_name"`
	GuideID   string `json:"guide_id"`
	GuideName string `json:"guide_name,omitempty"`
	Date      string `json:"date"`
	TimeSlot  string `json:"time_slot"`
	Product   string `json:"product,omitempty"`
	WhatsApp  string `json:"whatsapp,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// UpdateVisitRequest is the payload for PATCH /api/visits/:id.
type UpdateVisitRequest struct {
	Status    string         `json:"status,omitempty"`
	Feedback  *VisitFeedback `json:"feedback,omitempty"`
	GuideID   string         `json:"guide_id,omitempty"`
	GuideName string         `json:"guide_name,omitempty"`
	Date      string         `json:"date,omitempty"`
	TimeSlot  string         `json:"time_slot,omitempty"`
}

// UpdateAvailabilityRequest is the payload for PATCH /api/guides/:id/availability.
type UpdateAvailabilityRequest struct {
	WeeklySchedule   map[string][]string `json:"weekly_schedule,omitempty"`
	UnavailableDates []DateRange         `json:"unavailable_dates,omitempty"`
	MaxPerSlot       *int                `json:"max_per_slot,omitempty"`
}

// VisitRepository defines persistence for visits and guides.
type VisitRepository interface {
	Create(ctx context.Context, v *Visit) error
	Get(ctx context.Context, id string) (*Visit, error)
	Update(ctx context.Context, id string, update UpdateVisitRequest) error
	ListByGuide(ctx context.Context, guideID string) ([]*Visit, error)
	ListByLead(ctx context.Context, leadID string) ([]*Visit, error)
	ListAll(ctx context.Context) ([]*Visit, error)
	CountBySlot(ctx context.Context, guideID, date, timeSlot string) (int, error)
	// Guide operations
	ListGuides(ctx context.Context) ([]*Guide, error)
	GetGuide(ctx context.Context, id string) (*Guide, error)
	CreateGuide(ctx context.Context, g *Guide) error
	UpdateGuide(ctx context.Context, id string, g *Guide) error
	UpdateGuideAvailability(ctx context.Context, id string, req UpdateAvailabilityRequest) error
	ListAvailableSlots(ctx context.Context, date string) ([]*GuideSlot, error)
	ListAvailableSlotsForWeek(ctx context.Context, dates []string) ([]*GuideSlot, error)
}
