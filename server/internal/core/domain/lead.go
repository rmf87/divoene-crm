package domain

import "context"

// StageHistoryEntry records a stage change for audit trail.
type StageHistoryEntry struct {
	Stage     string `json:"stage"`
	ChangedAt string `json:"changed_at"`
	ChangedBy string `json:"changed_by"`
}

// EventInfo holds validated event details collected at the "validated" stage.
type EventInfo struct {
	PossibleDates      []string `json:"possible_dates,omitempty"`
	EventType          string   `json:"event_type,omitempty"`
	DesiredDurationHrs int      `json:"desired_duration_hours,omitempty"`
	DesiredDayOfWeek   string   `json:"desired_day_of_week,omitempty"`
	EstimatedPeople    int      `json:"estimated_people,omitempty"`
}

// ContactPersonInfo holds the decision-maker contact details.
type ContactPersonInfo struct {
	Name     string `json:"name"`
	WhatsApp string `json:"whatsapp"`
	Role     string `json:"role,omitempty"`
}

// AddOnItem is an extra service contracted for the event.
type AddOnItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// Note is a quick seller note attached to a lead.
type Note struct {
	Text      string `json:"text"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// Lead represents a lead in the pipeline.
type Lead struct {
	ID              string              `json:"id,omitempty"`
	Name            string              `json:"name"`
	WhatsApp        string              `json:"whatsapp"`
	Product         string              `json:"product"`
	DesiredDate     string              `json:"desired_date,omitempty"`
	Source          string              `json:"source,omitempty"`
	Stage           string              `json:"stage,omitempty"`
	StageHistory    []StageHistoryEntry `json:"stage_history,omitempty"`
	Event           *EventInfo          `json:"event,omitempty"`
	ContactPerson   *ContactPersonInfo  `json:"contact_person,omitempty"`
	AddOns          []AddOnItem         `json:"add_ons,omitempty"`
	Notes           []Note              `json:"notes,omitempty"`
	AssignedSeller  string              `json:"assigned_seller,omitempty"`
	CreatedAt       string              `json:"created_at,omitempty"`
	LastStageChange string              `json:"last_stage_change,omitempty"`
}

// CreateLeadRequest is the payload for POST /api/leads.
type CreateLeadRequest struct {
	Name        string `json:"name"`
	WhatsApp    string `json:"whatsapp"`
	Product     string `json:"product"`
	DesiredDate string `json:"desired_date,omitempty"`
	Source      string `json:"source,omitempty"`
}

// UpdateStageRequest is the payload for PATCH /api/leads/:id.
type UpdateStageRequest struct {
	Stage     string `json:"stage"`
	ChangedBy string `json:"changed_by,omitempty"`
}

// UpdateValidationRequest is the payload for PATCH /api/leads/:id/validation.
type UpdateValidationRequest struct {
	Event         *EventInfo         `json:"event,omitempty"`
	ContactPerson *ContactPersonInfo `json:"contact_person,omitempty"`
	AddOns        []AddOnItem        `json:"add_ons,omitempty"`
}

// LeadRepository defines persistence for leads.
type LeadRepository interface {
	Store(ctx context.Context, lead *Lead) error
	Get(ctx context.Context, id string) (*Lead, error)
	List(ctx context.Context, filters map[string]string) ([]*Lead, error)
	// FindByWhatsApp matches a lead by one of the candidate phone number forms
	// (E.164, national, or suffix). Used to route inbound WhatsApp messages.
	FindByWhatsApp(ctx context.Context, numbers []string) (*Lead, error)
	UpdateStage(ctx context.Context, id, stage, changedBy string) error
	UpdateValidation(ctx context.Context, id string, event *EventInfo, contactPerson *ContactPersonInfo, addOns []AddOnItem) error
	AddNote(ctx context.Context, id string, note Note) error
}
