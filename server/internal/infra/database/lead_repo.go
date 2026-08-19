package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// LeadRepository implements domain.LeadRepository using SQLite.
type LeadRepository struct {
	db *sql.DB
}

// NewLeadRepository creates a new LeadRepository.
func NewLeadRepository(db *sql.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// leadColumns is the shared projection for lead reads.
const leadColumns = `id, name, whatsapp, product, desired_date, source, stage, stage_history,
	event, contact_person, add_ons, notes, assigned_seller, created_at, last_stage_change`

// scanLead maps one scanned lead row into a domain.Lead.
func scanLead(l *domain.Lead, s interface{ Scan(...interface{}) error }) error {
	var stageHistoryRaw, eventRaw, contactPersonRaw, addOnsRaw, notesRaw string
	err := s.Scan(&l.ID, &l.Name, &l.WhatsApp, &l.Product, &l.DesiredDate, &l.Source, &l.Stage,
		&stageHistoryRaw, &eventRaw, &contactPersonRaw, &addOnsRaw, &notesRaw,
		&l.AssignedSeller, &l.CreatedAt, &l.LastStageChange)
	if err != nil {
		return err
	}
	unmarshalJSON(stageHistoryRaw, &l.StageHistory)
	unmarshalJSON(eventRaw, &l.Event)
	unmarshalJSON(contactPersonRaw, &l.ContactPerson)
	unmarshalJSON(addOnsRaw, &l.AddOns)
	unmarshalJSON(notesRaw, &l.Notes)
	return nil
}

func (r *LeadRepository) Store(ctx context.Context, lead *domain.Lead) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO leads (id, name, whatsapp, product, desired_date, source, stage, stage_history, event, contact_person, add_ons)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		lead.ID, lead.Name, lead.WhatsApp, lead.Product, lead.DesiredDate, lead.Source,
		lead.Stage, "[]", "", "", "[]",
	)
	if err != nil {
		return fmt.Errorf("insert lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) Get(ctx context.Context, id string) (*domain.Lead, error) {
	l := &domain.Lead{}
	err := scanLead(l, r.db.QueryRowContext(ctx,
		`SELECT `+leadColumns+` FROM leads WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lead não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan lead: %w", err)
	}
	return l, nil
}

func (r *LeadRepository) List(ctx context.Context, filters map[string]string) ([]*domain.Lead, error) {
	query := `SELECT ` + leadColumns + ` FROM leads`
	var args []interface{}
	var clauses []string
	if v, ok := filters["stage"]; ok {
		clauses = append(clauses, "stage = ?")
		args = append(args, v)
	}
	if v, ok := filters["assigned_seller"]; ok {
		clauses = append(clauses, "assigned_seller = ?")
		args = append(args, v)
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list leads: %w", err)
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		l := &domain.Lead{}
		if err := scanLead(l, rows); err != nil {
			return nil, fmt.Errorf("scan lead row: %w", err)
		}
		leads = append(leads, l)
	}
	return leads, rows.Err()
}

// FindByWhatsApp returns the first lead whose stored number matches any of the
// candidate forms (E.164, national, or bare suffix).
func (r *LeadRepository) FindByWhatsApp(ctx context.Context, numbers []string) (*domain.Lead, error) {
	if len(numbers) == 0 {
		return nil, fmt.Errorf("lead não encontrado")
	}
	placeholders := make([]string, len(numbers))
	args := make([]interface{}, len(numbers))
	for i, n := range numbers {
		placeholders[i] = "?"
		args[i] = n
	}

	query := `SELECT ` + leadColumns + ` FROM leads WHERE whatsapp IN (` + strings.Join(placeholders, ",") + `)
	          ORDER BY created_at ASC LIMIT 1`

	l := &domain.Lead{}
	err := scanLead(l, r.db.QueryRowContext(ctx, query, args...))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lead não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan lead: %w", err)
	}
	return l, nil
}

func (r *LeadRepository) UpdateStage(ctx context.Context, id, stage, changedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Append to stage_history
	_, err := r.db.ExecContext(ctx,
		`UPDATE leads SET stage = ?, last_stage_change = ?,
		 stage_history = json_insert(coalesce(nullif(stage_history, ''), '[]'), '$[#]', json_object('stage', ?, 'changed_at', ?, 'changed_by', ?))
		 WHERE id = ?`,
		stage, now, stage, now, changedBy, id)
	if err != nil {
		return fmt.Errorf("update stage: %w", err)
	}
	return nil
}

func (r *LeadRepository) UpdateValidation(ctx context.Context, id string, event *domain.EventInfo, contactPerson *domain.ContactPersonInfo, addOns []domain.AddOnItem) error {
	var eventJSON, contactJSON, addOnsJSON string
	if event != nil {
		eventJSON = marshalJSON(event)
	}
	if contactPerson != nil {
		contactJSON = marshalJSON(contactPerson)
	}
	if addOns != nil {
		addOnsJSON = marshalJSON(addOns)
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE leads SET event = CASE WHEN ? = '' THEN event ELSE ? END,
		 contact_person = CASE WHEN ? = '' THEN contact_person ELSE ? END,
		 add_ons = CASE WHEN ? = '' THEN add_ons ELSE ? END
		 WHERE id = ?`,
		eventJSON, eventJSON,
		contactJSON, contactJSON,
		addOnsJSON, addOnsJSON,
		id)
	if err != nil {
		return fmt.Errorf("update validation: %w", err)
	}
	return nil
}

// AddNote appends a quick note to a lead's notes JSON array.
func (r *LeadRepository) AddNote(ctx context.Context, id string, note domain.Note) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE leads SET notes = json_insert(coalesce(nullif(notes, ''), '[]'), '$[#]',
		   json_object('text', ?, 'created_by', ?, 'created_at', ?))
		 WHERE id = ?`,
		note.Text, note.CreatedBy, note.CreatedAt, id)
	if err != nil {
		return fmt.Errorf("add note: %w", err)
	}
	return nil
}
