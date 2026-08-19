package domain

import (
	"context"
	"time"
)

// Message directions and statuses for chat_messages.
const (
	DirectionSeller = "seller"
	DirectionLead   = "lead"
	DirectionSystem = "system"

	MessageStatusSent     = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
	MessageStatusReceived  = "received"
)

// ChatMessage is a single message in a lead's WhatsApp conversation.
type ChatMessage struct {
	ID          int64     `json:"id"`
	LeadID      string    `json:"lead_id"`
	WAMessageID string    `json:"wa_message_id,omitempty"`
	Direction   string    `json:"direction"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	SentAt      time.Time `json:"sent_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChatMessageRepository defines persistence for chat messages.
type ChatMessageRepository interface {
	Store(ctx context.Context, msg *ChatMessage) error
	ListByLead(ctx context.Context, leadID string) ([]*ChatMessage, error)
	FindByWAMessageID(ctx context.Context, waMessageID string) (*ChatMessage, error)
	UpdateStatus(ctx context.Context, waMessageID, status string) error
}

// WhatsAppGateway is the port for the WhatsApp Business Cloud API.
type WhatsAppGateway interface {
	// IsMock reports whether the gateway runs without real credentials.
	IsMock() bool
	// SendText sends a free-form message (24h customer-service window).
	// Returns the WhatsApp message id.
	SendText(ctx context.Context, to, body string) (string, error)
	// SendTemplate sends an approved template message.
	SendTemplate(ctx context.Context, to, templateName, lang string, vars map[string]string) (string, error)
	// VerifySignature validates the X-Hub-Signature-256 webhook header.
	VerifySignature(payload []byte, signature string) bool
}

// WhatsAppWebhookPayload is the shape of the Cloud API messages webhook.
type WhatsAppWebhookPayload struct {
	Entry []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
				Statuses []struct {
					ID        string `json:"id"`
					Status    string `json:"status"`
					Timestamp string `json:"timestamp"`
				} `json:"statuses"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}

// SendMessageRequest is the payload for POST /api/leads/:id/messages.
type SendMessageRequest struct {
	Body     string `json:"body"`
	Template string `json:"template,omitempty"`
	Lang     string `json:"lang,omitempty"`
	Vars     map[string]string `json:"vars,omitempty"`
}
