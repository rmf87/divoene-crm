package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rmf87/divoene/internal/core/domain"
)

// ConfigService manages configuration settings with env-var fallback.
type ConfigService struct {
	repo   domain.ConfigRepository
	envMap map[string]string // config key → env var name
}

// NewConfigService creates a ConfigService.
// envOverrides maps config keys to env var names for fallback when DB is empty.
func NewConfigService(repo domain.ConfigRepository, envOverrides map[string]string) *ConfigService {
	return &ConfigService{
		repo:   repo,
		envMap: envOverrides,
	}
}

// GetAll returns all config settings. Secrets are masked.
func (s *ConfigService) GetAll(ctx context.Context) ([]domain.ConfigSetting, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("config repository not initialized")
	}
	settings, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	// If DB value is empty and env fallback exists, show it as masked
	for i, setting := range settings {
		if setting.Value == "" && setting.MaskedValue == "" && setting.ValueType == "secret" {
			if envVar, ok := s.envMap[setting.Key]; ok {
				if v := os.Getenv(envVar); v != "" {
					settings[i].MaskedValue = maskValue(v)
				}
			}
		}
	}
	return settings, nil
}

// GetByKey returns a single config setting. Secrets are decrypted.
func (s *ConfigService) GetByKey(ctx context.Context, key string) (*domain.ConfigSetting, error) {
	return s.repo.GetByKey(ctx, key)
}

// Upsert saves a config setting and returns the result with masked value.
func (s *ConfigService) Upsert(ctx context.Context, key, value, updatedBy string) (*domain.ConfigSetting, error) {
	existing, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("config not found: %s", key)
	}

	setting := &domain.ConfigSetting{
		Key:         key,
		Value:       value,
		ValueType:   existing.ValueType,
		Description: existing.Description,
		UpdatedBy:   updatedBy,
	}

	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}

	// Return masked version — never send raw secret back
	setting.Value = ""
	if existing.ValueType == "secret" {
		setting.MaskedValue = maskValue(value)
	}
	return setting, nil
}

// GetDecryptedValue returns the plaintext value for internal use.
// DB takes priority; env var is the fallback when DB is empty.
func (s *ConfigService) GetDecryptedValue(ctx context.Context, key string) (string, error) {
	// Try DB first
	setting, err := s.repo.GetByKey(ctx, key)
	if err == nil && setting.Value != "" {
		return setting.Value, nil
	}
	// Fall back to env var
	if envVar, ok := s.envMap[key]; ok {
		if v := os.Getenv(envVar); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("config %s not found in DB or env", key)
}

// maskValue shows first 3 and last 2 characters.
func maskValue(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return s[:3] + strings.Repeat("*", len(s)-5) + s[len(s)-2:]
}
