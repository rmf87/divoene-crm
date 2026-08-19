package clients

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const defaultWhatsAppBaseURL = "https://graph.facebook.com/v21.0"

// WhatsAppClient talks to the Meta WhatsApp Business Cloud API.
type WhatsAppClient struct {
	baseURL       string
	token         string
	phoneNumberID string
	appSecret     string
	verifyToken   string
	httpClient    *http.Client
}

// NewWhatsAppClient creates a WhatsAppClient. With an empty token the client
// runs in mock mode and never calls the Meta API.
func NewWhatsAppClient(baseURL, token, phoneNumberID, appSecret, verifyToken string) *WhatsAppClient {
	if baseURL == "" {
		baseURL = defaultWhatsAppBaseURL
	}
	return &WhatsAppClient{
		baseURL:       baseURL,
		token:         token,
		phoneNumberID: phoneNumberID,
		appSecret:     appSecret,
		verifyToken:   verifyToken,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// IsMock reports whether the client runs without real credentials.
func (c *WhatsAppClient) IsMock() bool {
	return c.token == "" || c.phoneNumberID == ""
}

// GetVerifyToken returns the webhook verification token.
func (c *WhatsAppClient) GetVerifyToken() string {
	return c.verifyToken
}

// SendText sends a free-form text message. Returns the WhatsApp message id.
func (c *WhatsAppClient) SendText(ctx context.Context, to, body string) (string, error) {
	if c.IsMock() {
		return mockWamid(to, body), nil
	}
	return c.post(ctx, map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text":              map[string]interface{}{"body": body},
	})
}

// SendTemplate sends an approved template message with body parameters.
func (c *WhatsAppClient) SendTemplate(ctx context.Context, to, templateName, lang string, vars map[string]string) (string, error) {
	params := templateParams(vars)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]interface{}{
				"code": lang,
			},
		},
	}
	if len(params) > 0 {
		payload["template"].(map[string]interface{})["components"] = []map[string]interface{}{
			{"type": "body", "parameters": params},
		}
	}

	if c.IsMock() {
		return mockWamid(to, templateName), nil
	}
	return c.post(ctx, payload)
}

// mockWamid builds a unique mock message id (deterministic hash + timestamp so
// repeated sends never collide with the chat_messages unique index).
func mockWamid(seed, suffix string) string {
	return "wamid-mock-" + hashToken(seed+suffix) + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
}

// templateParams converts a map into an ordered list of body parameters.
func templateParams(vars map[string]string) []map[string]string {
	if len(vars) == 0 {
		return nil
	}
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]map[string]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, map[string]string{"type": "text", "text": vars[k]})
	}
	return out
}

func (c *WhatsAppClient) post(ctx context.Context, payload map[string]interface{}) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("whatsapp marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/"+c.phoneNumberID+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("whatsapp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("whatsapp send: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("whatsapp API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("whatsapp response: %w", err)
	}
	if len(result.Messages) == 0 {
		return "", fmt.Errorf("whatsapp response missing message id: %s", string(respBody))
	}
	return result.Messages[0].ID, nil
}

// VerifySignature validates the X-Hub-Signature-256 header against the raw
// request body using the app secret.
func (c *WhatsAppClient) VerifySignature(payload []byte, signature string) bool {
	if c.appSecret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(c.appSecret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

// Reload re-reads WhatsApp configuration from the provider.
// Called after config changes to take effect without restart.
func (c *WhatsAppClient) Reload(ctx context.Context, p ConfigProvider) {
	if v, err := p.GetDecryptedValue(ctx, "whatsapp_token"); err == nil && v != "" {
		c.token = v
	}
	if v, err := p.GetDecryptedValue(ctx, "whatsapp_phone_number_id"); err == nil && v != "" {
		c.phoneNumberID = v
	}
	if v, err := p.GetDecryptedValue(ctx, "whatsapp_app_secret"); err == nil && v != "" {
		c.appSecret = v
	}
	if v, err := p.GetDecryptedValue(ctx, "whatsapp_webhook_verify_token"); err == nil && v != "" {
		c.verifyToken = v
	}
	if v, err := p.GetDecryptedValue(ctx, "whatsapp_base_url"); err == nil && v != "" {
		c.baseURL = v
	}
}

func hashToken(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}
