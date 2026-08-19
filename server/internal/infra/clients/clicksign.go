package clients

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"context"

	"github.com/rmf87/divoene/internal/core/domain"
)

type ClicksignClient struct {
	baseURL       string
	apiKey        string
	webhookSecret string
	templateKey   string
	httpClient    *http.Client
}

func NewClicksignClient(baseURL, apiKey, webhookSecret, templateKey string) *ClicksignClient {
	if baseURL == "" {
		baseURL = "https://sandbox.clicksign.com/api/v3"
	}
	return &ClicksignClient{
		baseURL:       baseURL,
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		templateKey:   templateKey,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ClicksignClient) IsMock() bool {
	return c.apiKey == ""
}

func (c *ClicksignClient) GetWebhookSecret() string {
	return c.webhookSecret
}

// CreateAndSend creates a signer, links to document, and sends notification.
// Returns the full Contract with doc keys populated (mock or real).
func (c *ClicksignClient) CreateAndSend(contractReq *domain.CreateContractRequest) (*domain.Contract, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	if c.IsMock() {
		return &domain.Contract{
			LeadID:          contractReq.LeadID,
			ClicksignDocKey: "mock-doc-key-" + hex.EncodeToString([]byte(contractReq.LeadEmail))[:8],
			SignerKey:       "mock-signer-key-" + hex.EncodeToString([]byte(contractReq.LeadEmail))[:8],
			RequestSignKey:  "mock-request-key-" + hex.EncodeToString([]byte(contractReq.LeadEmail))[:8],
			Amount:          contractReq.Amount,
			Product:         contractReq.Product,
			EventType:       contractReq.EventType,
			EventDate:       contractReq.EventDate,
			EventDuration:   contractReq.EventDuration,
			EstPeople:       contractReq.EstPeople,
			ContactName:     contractReq.ContactName,
			ContactWpp:      contractReq.ContactWpp,
			ContactRole:     contractReq.ContactRole,
			PaymentCond:     contractReq.PaymentCond,
			Notes:           contractReq.Notes,
			Status:          "sent",
			SentAt:          now,
		}, nil
	}

	signer, err := c.createSigner(contractReq.LeadName, contractReq.LeadEmail, contractReq.LeadWhatsApp)
	if err != nil {
		return nil, fmt.Errorf("create signer: %w", err)
	}

	docKey, err := c.createDocument(contractReq.LeadName, contractReq.LeadEmail)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	reqKey, err := c.linkSigner(docKey, signer.Key)
	if err != nil {
		return nil, fmt.Errorf("link signer: %w", err)
	}

	message := buildClicksignMessage(contractReq)
	if err := c.sendNotification(reqKey, message); err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}

	return &domain.Contract{
		LeadID:          contractReq.LeadID,
		ClicksignDocKey: docKey,
		SignerKey:       signer.Key,
		RequestSignKey:  reqKey,
		Amount:          contractReq.Amount,
		Product:         contractReq.Product,
		EventType:       contractReq.EventType,
		EventDate:       contractReq.EventDate,
		EventDuration:   contractReq.EventDuration,
		EstPeople:       contractReq.EstPeople,
		ContactName:     contractReq.ContactName,
		ContactWpp:      contractReq.ContactWpp,
		ContactRole:     contractReq.ContactRole,
		PaymentCond:     contractReq.PaymentCond,
		Notes:           contractReq.Notes,
		Status:          "sent",
		SentAt:          now,
	}, nil
}

func buildClicksignMessage(req *domain.CreateContractRequest) string {
	amountReais := float64(req.Amount) / 100.0
	msg := fmt.Sprintf(
		"Olá %s,\n\nSegue o contrato para o seu evento na Chácara Divoene.\n\n"+
			"📅 Evento: %s\n📦 Produto: %s\n⏱ Duração: %d horas\n"+
			"👥 Público estimado: %d pessoas\n💰 Valor total: R$ %.2f\n💳 Condições: %s\n\n",
		req.ContactName, req.EventDate, productLabel(req.Product),
		req.EventDuration, req.EstPeople, amountReais, req.PaymentCond,
	)
	if req.Notes != "" {
		msg += fmt.Sprintf("📝 Observações: %s\n\n", req.Notes)
	}
	msg += "Por favor, assine digitalmente para confirmar sua reserva.\n\nAtenciosamente,\nEquipe Chácara Divoene"
	return msg
}

func productLabel(id string) string {
	labels := map[string]string{
		"ensaio_fotografico": "Ensaio Fotográfico",
		"locacao_eventos":    "Locação para Eventos",
		"corporativo":        "Corporativo",
		"casamentos":         "Casamentos",
		"buffet_infantil":    "Buffet Infantil",
		"passeios_escolares": "Passeios Escolares",
	}
	if l, ok := labels[id]; ok {
		return l
	}
	return id
}

func (c *ClicksignClient) PollStatus(documentKey string) (string, error) {
	return "sent", nil
}

type ClicksignSigner struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone_number"`
}

func (c *ClicksignClient) VerifyWebhookSignature(body []byte, contentHmacHeader string) bool {
	if c.webhookSecret == "" {
		return true
	}
	if !strings.HasPrefix(contentHmacHeader, "sha256=") {
		return false
	}
	expected := strings.TrimPrefix(contentHmacHeader, "sha256=")
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(body)
	computed := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(expected))
}

func (c *ClicksignClient) createSigner(name, email, phone string) (*ClicksignSigner, error) {
	auths := []string{"email"}
	if phone != "" {
		auths = append(auths, "whatsapp")
	}
	body := map[string]interface{}{
		"signer": map[string]interface{}{
			"name":         name,
			"email":        email,
			"auths":        auths,
			"phone_number": phone,
		},
	}
	var resp struct {
		Signer ClicksignSigner `json:"signer"`
	}
	if err := c.post("/api/v1/signers", body, &resp); err != nil {
		return nil, err
	}
	return &resp.Signer, nil
}

func (c *ClicksignClient) createDocument(clientName, clientEmail string) (string, error) {
	if c.templateKey == "" {
		return c.createBlankDocument(clientName)
	}
	body := map[string]interface{}{
		"document": map[string]interface{}{
			"path": fmt.Sprintf("/Contratos/%s-%d.docx", sanitizePath(clientName), time.Now().Unix()),
			"template": map[string]interface{}{
				"data": map[string]string{
					"Client Name":  clientName,
					"Client Email": clientEmail,
				},
			},
		},
	}
	var resp struct {
		Document struct {
			Key string `json:"key"`
		} `json:"document"`
	}
	if err := c.post("/api/v2/templates/"+c.templateKey+"/documents", body, &resp); err != nil {
		return "", err
	}
	return resp.Document.Key, nil
}

func (c *ClicksignClient) createBlankDocument(clientName string) (string, error) {
	body := map[string]interface{}{
		"document": map[string]interface{}{
			"path":               fmt.Sprintf("/Contratos/%s-%d.docx", sanitizePath(clientName), time.Now().Unix()),
			"auto_close":         true,
			"deadline_at":        time.Now().AddDate(0, 1, 0).Format("2006-01-02T15:04:05-03:00"),
			"remind_interval":    3,
			"sequence_enabled":   false,
			"block_after_refusal": true,
		},
	}
	var resp struct {
		Document struct {
			Key string `json:"key"`
		} `json:"document"`
	}
	if err := c.post("/api/v1/documents", body, &resp); err != nil {
		return "", err
	}
	return resp.Document.Key, nil
}

func (c *ClicksignClient) linkSigner(docKey, signerKey string) (string, error) {
	body := map[string]interface{}{
		"list": map[string]interface{}{
			"document_key": docKey,
			"signer_key":   signerKey,
			"sign_as":      "sign",
			"refusable":    true,
		},
	}
	var resp struct {
		List struct {
			RequestSignatureKey string `json:"request_signature_key"`
		} `json:"list"`
	}
	if err := c.post("/api/v1/lists", body, &resp); err != nil {
		return "", err
	}
	return resp.List.RequestSignatureKey, nil
}

func (c *ClicksignClient) sendNotification(requestSignKey, message string) error {
	body := map[string]interface{}{
		"request_signature_key": requestSignKey,
		"message":               message,
	}
	var resp interface{}
	return c.post("/api/v1/notifications", body, &resp)
}

func (c *ClicksignClient) post(path string, body interface{}, result interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	url := c.baseURL
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	url += path
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("access-token", c.apiKey)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil {
				return json.Unmarshal(respBody, result)
			}
			return nil
		}
		lastErr = fmt.Errorf("clicksign API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return lastErr
}

func sanitizePath(name string) string {
	r := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-",
		"*", "", "?", "", "\"", "", "<", "", ">", "", "|", "")
	s := r.Replace(name)
	s = strings.Map(func(r rune) rune {
		if r > 127 {
			return -1
		}
		return r
	}, s)
	if s == "" {
		s = "contrato"
	}
	return s
}

// Reload re-reads Clicksign configuration from the provider.
// Called after config changes to take effect without restart.
func (c *ClicksignClient) Reload(ctx context.Context, p ConfigProvider) {
	if v, err := p.GetDecryptedValue(ctx, "clicksign_api_key"); err == nil && v != "" {
		c.apiKey = v
	}
	if v, err := p.GetDecryptedValue(ctx, "clicksign_base_url"); err == nil && v != "" {
		c.baseURL = v
	}
	if v, err := p.GetDecryptedValue(ctx, "clicksign_webhook_secret"); err == nil && v != "" {
		c.webhookSecret = v
	}
	if v, err := p.GetDecryptedValue(ctx, "clicksign_template_key"); err == nil && v != "" {
		c.templateKey = v
	}
}
