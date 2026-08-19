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

// RegisterWhatsAppWebhook registers the public webhook routes (verification
// handshake + inbound messages/status updates).
func RegisterWhatsAppWebhook(public *gin.RouterGroup, chatSvc *services.ChatService, whatsappClient *clients.WhatsAppClient) {
	h := &whatsappHandler{svc: chatSvc, client: whatsappClient}
	public.GET("/whatsapp/webhook", h.verifyWebhook)
	public.POST("/whatsapp/webhook", h.webhook)
}

// RegisterWhatsAppMessages registers the protected per-lead message endpoints.
// leadsGroup must carry the seller/manager middleware.
func RegisterWhatsAppMessages(leadsGroup *gin.RouterGroup, chatSvc *services.ChatService) {
	h := &whatsappHandler{svc: chatSvc}
	leadsGroup.GET("/:id/messages", h.listMessages)
	leadsGroup.POST("/:id/messages", h.sendMessage)
}

type whatsappHandler struct {
	svc    *services.ChatService
	client *clients.WhatsAppClient
}

// verifyWebhook answers the Meta webhook verification handshake.
func (h *whatsappHandler) verifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token != "" && token == h.client.GetVerifyToken() {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verification token"})
}

// webhook receives inbound messages and status updates from the Cloud API.
func (h *whatsappHandler) webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	if !h.client.VerifySignature(body, c.GetHeader("X-Hub-Signature-256")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	var payload domain.WhatsAppWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := h.svc.HandleInbound(c.Request.Context(), payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *whatsappHandler) listMessages(c *gin.Context) {
	messages, err := h.svc.ListByLead(c.Request.Context(), c.Param("id"))
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if messages == nil {
		messages = []*domain.ChatMessage{}
	}
	c.JSON(http.StatusOK, messages)
}

func (h *whatsappHandler) sendMessage(c *gin.Context) {
	var req domain.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var msg *domain.ChatMessage
	var err error
	if req.Template != "" {
		msg, err = h.svc.SendTemplate(c.Request.Context(), c.Param("id"), req.Template, req.Lang, req.Vars)
	} else {
		msg, err = h.svc.SendMessage(c.Request.Context(), c.Param("id"), req.Body)
	}
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "mensagem vazia" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}
