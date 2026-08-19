package clients

import "context"

// ConfigProvider abstracts access to configuration values.
// Implemented by services.ConfigService (avoids circular import).
type ConfigProvider interface {
	GetDecryptedValue(ctx context.Context, key string) (string, error)
}
