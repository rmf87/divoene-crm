package domain

import (
	"context"
	"errors"
)

// ErrConfigNotFound is returned when a config key does not exist.
var ErrConfigNotFound = errors.New("config not found")

// ConfigSetting represents a single configuration setting.
// For secret-type settings, Value is empty in list responses (use MaskedValue).
type ConfigSetting struct {
	Key         string `json:"key"`
	Value       string `json:"value,omitempty"`
	MaskedValue string `json:"masked_value,omitempty"`
	ValueType   string `json:"value_type"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
}

// ConfigRepository defines persistence for configuration settings.
type ConfigRepository interface {
	GetAll(ctx context.Context) ([]ConfigSetting, error)
	GetByKey(ctx context.Context, key string) (*ConfigSetting, error)
	Upsert(ctx context.Context, setting *ConfigSetting) error
}
