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
	"time"

	"context"

	"github.com/rmf87/divoene/internal/core/domain"
)

type OpenPixClient struct {
	appID      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenPixClient(baseURL, appID string) *OpenPixClient {
	if baseURL == "" {
		baseURL = "https://api.woovi-sandbox.com/api"
	}
	return &OpenPixClient{
		appID:      appID,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenPixClient) IsMock() bool {
	return c.appID == ""
}

func (c *OpenPixClient) CreateCharge(amount int64, description, chargeType, payerName, payerEmail, payerTaxID, payerPhone string, addr *domain.ChargeAddress) (*domain.PIXChargeResult, error) {
	if c.IsMock() {
		return &domain.PIXChargeResult{
			TransactionID: "mock-tx-" + hex.EncodeToString([]byte(payerEmail))[:8],
			BrCode:        "00020126580014br.gov.bcb.pix0136mock-qr-code-123456520400005303986540510.005802BR5912DIVOENE6009SAO PAULO62070503***6304A1B2",
			PaymentLink:   "https://pay.openpix.com.br/sandbox/mock/" + hex.EncodeToString([]byte(payerEmail))[:8],
			ExpiresIn:     86400,
		}, nil
	}

	customer := map[string]interface{}{
		"name":  payerName,
		"email": payerEmail,
	}
	if payerTaxID != "" {
		customer["taxID"] = payerTaxID
	}
	if payerPhone != "" {
		customer["phone"] = payerPhone
	}
	if addr != nil {
		customer["address"] = map[string]interface{}{
			"zipcode": addr.ZipCode, "street": addr.Street, "number": addr.Number,
			"neighborhood": addr.Neighborhood, "city": addr.City, "state": addr.State,
		}
	}

	body := map[string]interface{}{
		"correlationID": fmt.Sprintf("divoene-%s-%d", payerEmail, time.Now().UnixNano()),
		"value":         int(amount),
		"comment":       description,
		"type":          chargeType,
		"customer":      customer,
	}

	var result struct {
		Charge struct {
			TransactionID string `json:"transactionID"`
			BrCode        string `json:"brCode"`
			PaymentLink   string `json:"paymentLinkUrl"`
			Status        string `json:"status"`
			ExpiresIn     int    `json:"expiresIn"`
		} `json:"charge"`
	}
	if err := c.post("/v1/charge", body, &result); err != nil {
		return nil, err
	}

	return &domain.PIXChargeResult{
		TransactionID: result.Charge.TransactionID,
		BrCode:        result.Charge.BrCode,
		PaymentLink:   result.Charge.PaymentLink,
		ExpiresIn:     result.Charge.ExpiresIn,
	}, nil
}

func (c *OpenPixClient) GetCharge(transactionID string) (string, error) {
	if c.IsMock() {
		return "COMPLETED", nil
	}
	var result struct {
		Charge struct {
			Status string `json:"status"`
		} `json:"charge"`
	}
	if err := c.get("/v1/charge/"+transactionID, &result); err != nil {
		return "", err
	}
	return result.Charge.Status, nil
}

func (c *OpenPixClient) VerifyWebhookSignature(body []byte, xWebhookSignature string) bool {
	if c.appID == "" {
		return true
	}
	if xWebhookSignature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(c.appID))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(xWebhookSignature))
}

func (c *OpenPixClient) post(path string, body interface{}, result interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.baseURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.appID)
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
		lastErr = fmt.Errorf("openpix API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return lastErr
}

func (c *OpenPixClient) get(path string, result interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.appID)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("openpix API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return json.Unmarshal(respBody, result)
}

// Reload re-reads OpenPix configuration from the provider.
// Called after config changes to take effect without restart.
func (c *OpenPixClient) Reload(ctx context.Context, p ConfigProvider) {
	if v, err := p.GetDecryptedValue(ctx, "openpix_app_id"); err == nil && v != "" {
		c.appID = v
	}
	if v, err := p.GetDecryptedValue(ctx, "openpix_base_url"); err == nil && v != "" {
		c.baseURL = v
	}
}
