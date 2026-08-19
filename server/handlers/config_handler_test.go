package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// mockConfigRepo implements domain.ConfigRepository for testing.
type mockConfigRepo struct {
	settings map[string]domain.ConfigSetting
}

func (m *mockConfigRepo) GetAll(ctx context.Context) ([]domain.ConfigSetting, error) {
	var out []domain.ConfigSetting
	for _, s := range m.settings {
		if s.ValueType == "secret" && s.Value != "" {
			s.MaskedValue = maskSecretForTest(s.Value)
			s.Value = ""
		}
		out = append(out, s)
	}
	return out, nil
}

func maskSecretForTest(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return s[:3] + strings.Repeat("*", len(s)-5) + s[len(s)-2:]
}

func (m *mockConfigRepo) GetByKey(ctx context.Context, key string) (*domain.ConfigSetting, error) {
	if s, ok := m.settings[key]; ok {
		return &s, nil
	}
	return nil, domain.ErrConfigNotFound
}

func (m *mockConfigRepo) Upsert(ctx context.Context, setting *domain.ConfigSetting) error {
	s := *setting
	m.settings[setting.Key] = s
	return nil
}

func newMockService() *services.ConfigService {
	repo := &mockConfigRepo{
		settings: map[string]domain.ConfigSetting{
			"clicksign_api_key": {
				Key:         "clicksign_api_key",
				Value:       "sk_test_abc",
				ValueType:   "secret",
				Description: "Clicksign API key",
			},
			"openpix_base_url": {
				Key:         "openpix_base_url",
				Value:       "https://api.woovi-sandbox.com/api",
				ValueType:   "string",
				Description: "OpenPix base URL",
			},
		},
	}
	return services.NewConfigService(repo, nil)
}

func setupConfigTestRouter(svc *services.ConfigService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rg := r.Group("/api/config")
	rg.Use(func(c *gin.Context) {
		c.Set("actor", &domain.Actor{UID: "mgr-1", Roles: []string{"manager"}, Name: "Admin"})
		c.Next()
	})
	var reloadedKeys []string
	RegisterConfigRoutes(rg, svc, func(key string) {
		reloadedKeys = append(reloadedKeys, key)
	})
	return r
}

func TestListConfigs(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings []domain.ConfigSetting
	if err := json.Unmarshal(w.Body.Bytes(), &settings); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(settings) < 2 {
		t.Errorf("expected at least 2 settings, got %d", len(settings))
	}

	// Secret should have empty Value and set MaskedValue
	for _, s := range settings {
		if s.ValueType == "secret" && s.Value != "" {
			t.Errorf("secret %q should have empty Value, got %q", s.Key, s.Value)
		}
	}
}

func TestGetConfig(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/openpix_base_url", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var s domain.ConfigSetting
	json.Unmarshal(w.Body.Bytes(), &s)
	if s.Key != "openpix_base_url" {
		t.Errorf("expected openpix_base_url, got %s", s.Key)
	}
}

func TestGetConfigNotFound(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateConfig(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	body := `{"value": "new-api-key-value"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/config/clicksign_api_key", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var s domain.ConfigSetting
	json.Unmarshal(w.Body.Bytes(), &s)
	if s.Value != "" {
		t.Errorf("response should not contain raw secret value, got %q", s.Value)
	}
	if s.MaskedValue == "" {
		t.Error("response should contain masked_value for secret")
	}
	if s.UpdatedBy != "Admin" {
		t.Errorf("expected UpdatedBy=Admin, got %q", s.UpdatedBy)
	}
}

func TestUpdateConfigNotFound(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	body := `{"value": "some-value"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/config/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateConfigInvalidJSON(t *testing.T) {
	svc := newMockService()
	r := setupConfigTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/config/clicksign_api_key", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
