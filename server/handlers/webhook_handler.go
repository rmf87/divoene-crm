package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
	"github.com/rmf87/divoene/internal/infra/clients"
)

// RegisterWebhookRoutes registers public /api/webhooks routes.
func RegisterWebhookRoutes(rg *gin.RouterGroup, contractSvc *services.ContractService, paymentSvc *services.PaymentService,
	clicksignClient *clients.ClicksignClient, openpixClient *clients.OpenPixClient) {

	h := &webhookHandler{
		contractSvc:   contractSvc,
		paymentSvc:    paymentSvc,
		clicksignClient: clicksignClient,
		openpixClient:   openpixClient,
	}
	rg.POST("/webhooks/clicksign", h.clicksign)
	rg.POST("/webhooks/openpix", h.openpix)
}

type webhookHandler struct {
	contractSvc   *services.ContractService
	paymentSvc    *services.PaymentService
	clicksignClient *clients.ClicksignClient
	openpixClient   *clients.OpenPixClient
}

func (h *webhookHandler) clicksign(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify signature
	hmacHeader := c.GetHeader("Content-Hmac")
	if !h.clicksignClient.VerifyWebhookSignature(body, hmacHeader) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	// Parse Clicksign event
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	event := extractClicksignEvent(raw)
	if event.EventName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	status, err := h.contractSvc.HandleWebhook(c.Request.Context(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (h *webhookHandler) openpix(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("x-webhook-signature")
	if !h.openpixClient.VerifyWebhookSignature(body, signature) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	// Extract transactionID from OpenPix payload
	var raw struct {
		Charge struct {
			TransactionID string `json:"transactionID"`
		} `json:"charge"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.Charge.TransactionID == "" {
		c.JSON(http.StatusOK, gin.H{"error": "ignored"})
		return
	}

	status, err := h.paymentSvc.HandleWebhook(c.Request.Context(), raw.Charge.TransactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func extractClicksignEvent(raw map[string]interface{}) domain.ClicksignWebhookEvent {
	ev := domain.ClicksignWebhookEvent{}

	if event, ok := raw["event"].(map[string]interface{}); ok {
		if name, ok := event["name"].(string); ok {
			ev.EventName = name
		}
	}

	if doc, ok := raw["document"].(map[string]interface{}); ok {
		if key, ok := doc["key"].(string); ok {
			ev.DocumentKey = key
		}
	}

	if signer, ok := raw["signer"].(map[string]interface{}); ok {
		if key, ok := signer["key"].(string); ok {
			ev.SignerKey = key
		}
		if email, ok := signer["email"].(string); ok {
			ev.SignerEmail = email
		}
		if name, ok := signer["name"].(string); ok {
			ev.SignerName = name
		}
		// Refusal reasons
		if reasons, ok := signer["refusal_reasons"].([]interface{}); ok {
			for _, r := range reasons {
				ev.RefusalReasons = append(ev.RefusalReasons, strings.TrimSpace(r.(string)))
			}
		}
	}

	return ev
}
