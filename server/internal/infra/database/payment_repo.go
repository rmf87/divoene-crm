package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rmf87/divoene/internal/core/domain"
)

// PaymentRepository implements domain.PaymentRepository using SQLite.
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments (id, lead_id, contract_id, openpix_transaction_id, amount,
		 description, type, charge_type, status, br_code, payment_link_url, created_at, created_by, confirmed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.LeadID, p.ContractID, p.OpenPixTransactionID, p.Amount,
		p.Description, p.Type, p.ChargeType, p.Status, p.BrCode, p.PaymentLinkURL,
		p.CreatedAt, p.CreatedBy, p.ConfirmedAt,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) Get(ctx context.Context, id string) (*domain.Payment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lead_id, contract_id, openpix_transaction_id, amount,
		 description, type, charge_type, status, br_code, payment_link_url,
		 created_at, created_by, confirmed_at
		 FROM payments WHERE id = ?`, id)

	p := &domain.Payment{}
	err := row.Scan(&p.ID, &p.LeadID, &p.ContractID, &p.OpenPixTransactionID, &p.Amount,
		&p.Description, &p.Type, &p.ChargeType, &p.Status, &p.BrCode, &p.PaymentLinkURL,
		&p.CreatedAt, &p.CreatedBy, &p.ConfirmedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pagamento não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) GetByTransactionID(ctx context.Context, txID string) (*domain.Payment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lead_id, contract_id, openpix_transaction_id, amount,
		 description, type, charge_type, status, br_code, payment_link_url,
		 created_at, created_by, confirmed_at
		 FROM payments WHERE openpix_transaction_id = ?`, txID)

	p := &domain.Payment{}
	err := row.Scan(&p.ID, &p.LeadID, &p.ContractID, &p.OpenPixTransactionID, &p.Amount,
		&p.Description, &p.Type, &p.ChargeType, &p.Status, &p.BrCode, &p.PaymentLinkURL,
		&p.CreatedAt, &p.CreatedBy, &p.ConfirmedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment by tx not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan payment by tx: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status string, extra map[string]string) error {
	query := `UPDATE payments SET status = ?`
	args := []interface{}{status}

	if v, ok := extra["confirmed_at"]; ok {
		query += `, confirmed_at = ?`
		args = append(args, v)
	}
	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}

func (r *PaymentRepository) ListByLead(ctx context.Context, leadID string) ([]*domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, lead_id, contract_id, openpix_transaction_id, amount,
		 description, type, charge_type, status, br_code, payment_link_url,
		 created_at, created_by, confirmed_at
		 FROM payments WHERE lead_id = ? ORDER BY created_at DESC`, leadID)
	if err != nil {
		return nil, fmt.Errorf("list payments by lead: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.LeadID, &p.ContractID, &p.OpenPixTransactionID, &p.Amount,
			&p.Description, &p.Type, &p.ChargeType, &p.Status, &p.BrCode, &p.PaymentLinkURL,
			&p.CreatedAt, &p.CreatedBy, &p.ConfirmedAt); err != nil {
			return nil, fmt.Errorf("scan payment row: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
