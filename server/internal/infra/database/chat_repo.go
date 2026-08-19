package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// ChatMessageRepository implements domain.ChatMessageRepository using SQLite.
type ChatMessageRepository struct {
	db *sql.DB
}

// NewChatMessageRepository creates a new ChatMessageRepository.
func NewChatMessageRepository(db *sql.DB) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

const chatColumns = `id, lead_id, wa_message_id, direction, body, status, sent_at, created_at`

func (r *ChatMessageRepository) scan(msg *domain.ChatMessage, row *sql.Row) error {
	var sentAt, createdAt string
	err := row.Scan(&msg.ID, &msg.LeadID, &msg.WAMessageID, &msg.Direction, &msg.Body,
		&msg.Status, &sentAt, &createdAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("message not found")
	}
	if err != nil {
		return fmt.Errorf("scan message: %w", err)
	}
	if t, err := time.Parse(time.RFC3339, sentAt); err == nil {
		msg.SentAt = t
	}
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		msg.CreatedAt = t
	}
	return nil
}

func (r *ChatMessageRepository) Store(ctx context.Context, msg *domain.ChatMessage) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO chat_messages (lead_id, wa_message_id, direction, body, status, sent_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.LeadID, msg.WAMessageID, msg.Direction, msg.Body, msg.Status, now, now)
	if err != nil {
		return fmt.Errorf("store message: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("message id: %w", err)
	}
	msg.ID = id
	msg.SentAt = time.Now().UTC()
	msg.CreatedAt = msg.SentAt
	return nil
}

func (r *ChatMessageRepository) ListByLead(ctx context.Context, leadID string) ([]*domain.ChatMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+chatColumns+` FROM chat_messages WHERE lead_id = ? ORDER BY sent_at ASC`, leadID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []*domain.ChatMessage
	for rows.Next() {
		msg := &domain.ChatMessage{}
		var sentAt, createdAt string
		if err := rows.Scan(&msg.ID, &msg.LeadID, &msg.WAMessageID, &msg.Direction, &msg.Body,
			&msg.Status, &sentAt, &createdAt); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		if t, err := time.Parse(time.RFC3339, sentAt); err == nil {
			msg.SentAt = t
		}
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			msg.CreatedAt = t
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (r *ChatMessageRepository) FindByWAMessageID(ctx context.Context, waMessageID string) (*domain.ChatMessage, error) {
	if waMessageID == "" {
		return nil, fmt.Errorf("message not found")
	}
	msg := &domain.ChatMessage{}
	err := r.scan(msg, r.db.QueryRowContext(ctx,
		`SELECT `+chatColumns+` FROM chat_messages WHERE wa_message_id = ?`, waMessageID))
	return msg, err
}

func (r *ChatMessageRepository) UpdateStatus(ctx context.Context, waMessageID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE chat_messages SET status = ? WHERE wa_message_id = ?`, status, waMessageID)
	if err != nil {
		return fmt.Errorf("update message status: %w", err)
	}
	return nil
}
