package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/infra/crypto"
)

// ConfigRepository implements domain.ConfigRepository using SQLite with AES-256-GCM encryption.
type ConfigRepository struct {
	db       *sql.DB
	cryptKey []byte // derived from JWT_SECRET; nil means encryption disabled
}

// NewConfigRepository creates a ConfigRepository. jwtSecret is used to derive the encryption key.
func NewConfigRepository(db *sql.DB, jwtSecret string) *ConfigRepository {
	r := &ConfigRepository{db: db}
	if jwtSecret != "" {
		r.cryptKey = crypto.DeriveKey(jwtSecret)
	}
	return r
}

// GetAll returns all config settings. Secret values are decrypted then masked.
func (r *ConfigRepository) GetAll(ctx context.Context) ([]domain.ConfigSetting, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT key, value, value_type, description, updated_at, updated_by
		 FROM config_settings ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("list configs: %w", err)
	}
	defer rows.Close()

	var settings []domain.ConfigSetting
	for rows.Next() {
		var s domain.ConfigSetting
		var updatedAt, updatedBy sql.NullString
		if err := rows.Scan(&s.Key, &s.Value, &s.ValueType, &s.Description, &updatedAt, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		s.UpdatedAt = updatedAt.String
		s.UpdatedBy = updatedBy.String

		if s.ValueType == "secret" && s.Value != "" {
			if plain, err := r.decrypt(s.Value); err == nil {
				s.MaskedValue = maskSecret(plain)
				s.Value = "" // never expose raw secret in list
			}
			// If decryption fails, leave both Value and MaskedValue empty
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// GetByKey returns a single config setting. Secret values are decrypted.
func (r *ConfigRepository) GetByKey(ctx context.Context, key string) (*domain.ConfigSetting, error) {
	var s domain.ConfigSetting
	var updatedAt, updatedBy sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT key, value, value_type, description, updated_at, updated_by
		 FROM config_settings WHERE key = ?`, key).
		Scan(&s.Key, &s.Value, &s.ValueType, &s.Description, &updatedAt, &updatedBy)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	s.UpdatedAt = updatedAt.String
	s.UpdatedBy = updatedBy.String

	if s.ValueType == "secret" && s.Value != "" {
		if plain, err := r.decrypt(s.Value); err == nil {
			s.Value = plain
		}
	}
	return &s, nil
}

// Upsert inserts or updates a config setting. Secret values are encrypted before storage.
func (r *ConfigRepository) Upsert(ctx context.Context, setting *domain.ConfigSetting) error {
	value := setting.Value

	// Encrypt secrets before storing
	if setting.ValueType == "secret" && value != "" && r.cryptKey != nil {
		enc, err := crypto.Encrypt([]byte(value), r.cryptKey)
		if err != nil {
			return fmt.Errorf("encrypt config: %w", err)
		}
		value = enc
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO config_settings (key, value, value_type, description, updated_at, updated_by)
		 VALUES (?, ?, ?, ?, datetime('now'), ?)
		 ON CONFLICT(key) DO UPDATE SET
		   value = excluded.value,
		   value_type = excluded.value_type,
		   description = excluded.description,
		   updated_at = datetime('now'),
		   updated_by = excluded.updated_by`,
		setting.Key, value, setting.ValueType, setting.Description, setting.UpdatedBy)
	if err != nil {
		return fmt.Errorf("upsert config: %w", err)
	}
	return nil
}

func (r *ConfigRepository) decrypt(encryptedHex string) (string, error) {
	if r.cryptKey == nil {
		return "", fmt.Errorf("encryption key not available")
	}
	plain, err := crypto.Decrypt(encryptedHex, r.cryptKey)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// maskSecret shows first 3 and last 2 characters: "cli****23".
// Short strings (≤5 chars) are fully replaced with asterisks.
func maskSecret(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return s[:3] + strings.Repeat("*", len(s)-5) + s[len(s)-2:]
}
