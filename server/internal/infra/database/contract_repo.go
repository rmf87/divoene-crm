package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rmf87/divoene/internal/core/domain"
)

// ContractRepository implements domain.ContractRepository using SQLite.
type ContractRepository struct {
	db *sql.DB
}

// NewContractRepository creates a new ContractRepository.
func NewContractRepository(db *sql.DB) *ContractRepository {
	return &ContractRepository{db: db}
}

func (r *ContractRepository) Create(ctx context.Context, c *domain.Contract) error {
	addOnsJSON := marshalJSON(c.AddOns)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO contracts (id, lead_id, clicksign_doc_key, signer_key, request_sign_key,
		 amount, product, event_type, event_date, event_duration, estimated_people,
		 contact_name, contact_whatsapp, contact_role, add_ons, payment_conditions, notes,
		 status, sent_at, sent_by, signed_at, declined_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.LeadID, c.ClicksignDocKey, c.SignerKey, c.RequestSignKey,
		c.Amount, c.Product, c.EventType, c.EventDate, c.EventDuration, c.EstPeople,
		c.ContactName, c.ContactWpp, c.ContactRole, addOnsJSON, c.PaymentCond, c.Notes,
		c.Status, c.SentAt, c.SentBy, c.SignedAt, c.DeclinedAt,
	)
	if err != nil {
		return fmt.Errorf("insert contract: %w", err)
	}
	return nil
}

func (r *ContractRepository) Get(ctx context.Context, id string) (*domain.Contract, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lead_id, clicksign_doc_key, signer_key, request_sign_key,
		 amount, product, event_type, event_date, event_duration, estimated_people,
		 contact_name, contact_whatsapp, contact_role, add_ons, payment_conditions, notes,
		 status, sent_at, sent_by, signed_at, declined_at, created_at
		 FROM contracts WHERE id = ?`, id)

	c := &domain.Contract{}
	var addOnsRaw string
	err := row.Scan(&c.ID, &c.LeadID, &c.ClicksignDocKey, &c.SignerKey, &c.RequestSignKey,
		&c.Amount, &c.Product, &c.EventType, &c.EventDate, &c.EventDuration, &c.EstPeople,
		&c.ContactName, &c.ContactWpp, &c.ContactRole, &addOnsRaw, &c.PaymentCond, &c.Notes,
		&c.Status, &c.SentAt, &c.SentBy, &c.SignedAt, &c.DeclinedAt, &c.Notes) // notes reused for created_at
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contrato não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan contract: %w", err)
	}
	if c.Notes == "" {
		c.Notes = ""
	}
	unmarshalJSON(addOnsRaw, &c.AddOns)
	return c, nil
}

func (r *ContractRepository) GetByDocKey(ctx context.Context, docKey string) (*domain.Contract, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lead_id, clicksign_doc_key, signer_key, request_sign_key,
		 amount, product, event_type, event_date, event_duration, estimated_people,
		 contact_name, contact_whatsapp, contact_role, add_ons, payment_conditions, notes,
		 status, sent_at, sent_by, signed_at, declined_at, created_at
		 FROM contracts WHERE clicksign_doc_key = ?`, docKey)

	c := &domain.Contract{}
	var addOnsRaw, created_at string // addOnsRaw placeholder, created_at not in domain.Contract
	err := row.Scan(&c.ID, &c.LeadID, &c.ClicksignDocKey, &c.SignerKey, &c.RequestSignKey,
		&c.Amount, &c.Product, &c.EventType, &c.EventDate, &c.EventDuration, &c.EstPeople,
		&c.ContactName, &c.ContactWpp, &c.ContactRole, &addOnsRaw, &c.PaymentCond, &c.Notes,
		&c.Status, &c.SentAt, &c.SentBy, &c.SignedAt, &c.DeclinedAt, &created_at)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contract by doc key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan contract by doc key: %w", err)
	}
	unmarshalJSON(addOnsRaw, &c.AddOns)
	return c, nil
}

func (r *ContractRepository) UpdateStatus(ctx context.Context, id string, status string, extra map[string]string) error {
	query := `UPDATE contracts SET status = ?`
	args := []interface{}{status}

	if v, ok := extra["signed_at"]; ok {
		query += `, signed_at = ?`
		args = append(args, v)
	}
	if v, ok := extra["declined_at"]; ok {
		query += `, declined_at = ?`
		args = append(args, v)
	}
	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update contract status: %w", err)
	}
	return nil
}

func (r *ContractRepository) ListAll(ctx context.Context) ([]*domain.Contract, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, clicksign_doc_key, signer_key, request_sign_key,
		 amount, product, event_type, event_date, event_duration, estimated_people,
		 contact_name, contact_whatsapp, contact_role, add_ons, payment_conditions, notes,
		 status, sent_at, sent_by, signed_at, declined_at
		 FROM contracts ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all contracts: %w", err)
	}
	defer rows.Close()

	var contracts []*domain.Contract
	for rows.Next() {
		c := &domain.Contract{}
		var addOnsRaw string
		if err := rows.Scan(&c.ID, &c.LeadID, &c.ClicksignDocKey, &c.SignerKey, &c.RequestSignKey,
			&c.Amount, &c.Product, &c.EventType, &c.EventDate, &c.EventDuration, &c.EstPeople,
			&c.ContactName, &c.ContactWpp, &c.ContactRole, &addOnsRaw, &c.PaymentCond, &c.Notes,
			&c.Status, &c.SentAt, &c.SentBy, &c.SignedAt, &c.DeclinedAt); err != nil {
			return nil, fmt.Errorf("scan contract row: %w", err)
		}
		unmarshalJSON(addOnsRaw, &c.AddOns)
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

func (r *ContractRepository) ListByLead(ctx context.Context, leadID string) ([]*domain.Contract, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, clicksign_doc_key, signer_key, request_sign_key,
		 amount, product, event_type, event_date, event_duration, estimated_people,
		 contact_name, contact_whatsapp, contact_role, add_ons, payment_conditions, notes,
		 status, sent_at, sent_by, signed_at, declined_at
		 FROM contracts WHERE lead_id = ? ORDER BY created_at DESC`, leadID)
	if err != nil {
		return nil, fmt.Errorf("list contracts by lead: %w", err)
	}
	defer rows.Close()

	var contracts []*domain.Contract
	for rows.Next() {
		c := &domain.Contract{}
		var addOnsRaw string
		if err := rows.Scan(&c.ID, &c.LeadID, &c.ClicksignDocKey, &c.SignerKey, &c.RequestSignKey,
			&c.Amount, &c.Product, &c.EventType, &c.EventDate, &c.EventDuration, &c.EstPeople,
			&c.ContactName, &c.ContactWpp, &c.ContactRole, &addOnsRaw, &c.PaymentCond, &c.Notes,
			&c.Status, &c.SentAt, &c.SentBy, &c.SignedAt, &c.DeclinedAt); err != nil {
			return nil, fmt.Errorf("scan contract row: %w", err)
		}
		unmarshalJSON(addOnsRaw, &c.AddOns)
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}
