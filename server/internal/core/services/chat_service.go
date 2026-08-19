package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// ChatService orchestrates WhatsApp conversations with leads.
type ChatService struct {
	repo          domain.ChatMessageRepository
	leads         domain.LeadRepository
	whatsapp      domain.WhatsAppGateway
	mockAutoreply bool
}

// NewChatService wires the message repo, lead repo and WhatsApp gateway.
// mockAutoreply enables synthetic lead replies in mock mode (dev only).
func NewChatService(repo domain.ChatMessageRepository, leads domain.LeadRepository, whatsapp domain.WhatsAppGateway, mockAutoreply bool) *ChatService {
	return &ChatService{repo: repo, leads: leads, whatsapp: whatsapp, mockAutoreply: mockAutoreply}
}

// ListByLead returns all messages of a lead, oldest first. It validates the
// lead exists so the API can surface a 404 instead of a silent empty list.
func (s *ChatService) ListByLead(ctx context.Context, leadID string) ([]*domain.ChatMessage, error) {
	if _, err := s.leads.Get(ctx, leadID); err != nil {
		return nil, err
	}
	return s.repo.ListByLead(ctx, leadID)
}

// SendMessage sends a free-form text to the lead's WhatsApp and persists it.
func (s *ChatService) SendMessage(ctx context.Context, leadID, body string) (*domain.ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("mensagem vazia")
	}
	lead, err := s.leads.Get(ctx, leadID)
	if err != nil {
		return nil, err
	}

	waID, err := s.whatsapp.SendText(ctx, toE164(lead.WhatsApp), body)
	if err != nil {
		return nil, err
	}

	msg := &domain.ChatMessage{
		LeadID:      lead.ID,
		WAMessageID: waID,
		Direction:   domain.DirectionSeller,
		Body:        body,
		Status:      domain.MessageStatusSent,
	}
	if err := s.repo.Store(ctx, msg); err != nil {
		return nil, err
	}
	s.scheduleMockReply(ctx, lead)
	return msg, nil
}

// SendTemplate sends an approved template message to the lead and persists it.
func (s *ChatService) SendTemplate(ctx context.Context, leadID, templateName, lang string, vars map[string]string) (*domain.ChatMessage, error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return nil, errors.New("template inválido")
	}
	if lang == "" {
		lang = "pt_BR"
	}
	lead, err := s.leads.Get(ctx, leadID)
	if err != nil {
		return nil, err
	}

	waID, err := s.whatsapp.SendTemplate(ctx, toE164(lead.WhatsApp), templateName, lang, vars)
	if err != nil {
		return nil, err
	}

	msg := &domain.ChatMessage{
		LeadID:      lead.ID,
		WAMessageID: waID,
		Direction:   domain.DirectionSeller,
		Body:        templateName,
		Status:      domain.MessageStatusSent,
	}
	if err := s.repo.Store(ctx, msg); err != nil {
		return nil, err
	}
	s.scheduleMockReply(ctx, lead)
	return msg, nil
}

// HandleInbound processes a WhatsApp webhook: delivery/read status updates and
// inbound text messages routed to the matching lead.
func (s *ChatService) HandleInbound(ctx context.Context, payload domain.WhatsAppWebhookPayload) error {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			value := change.Value

			// Delivery/read updates for outgoing messages.
			for _, st := range value.Statuses {
				if st.ID != "" {
					_ = s.repo.UpdateStatus(ctx, st.ID, normalizeStatus(st.Status))
				}
			}

			// Inbound text messages from leads.
			for _, m := range value.Messages {
				if m.Type != "text" {
					continue
				}
				lead, err := s.leads.FindByWhatsApp(ctx, phoneCandidates(m.From))
				if err != nil {
					// Unknown number — no lead to attach to; skip silently.
					continue
				}
				msg := &domain.ChatMessage{
					LeadID:      lead.ID,
					WAMessageID: m.ID,
					Direction:   domain.DirectionLead,
					Body:        m.Text.Body,
					Status:      domain.MessageStatusReceived,
				}
				_ = s.repo.Store(ctx, msg)
			}
		}
	}
	return nil
}

// scheduleMockReply injects a synthetic lead reply ~2s after a send, but only
// when running in mock mode with autoreply enabled (dev only).
func (s *ChatService) scheduleMockReply(ctx context.Context, lead *domain.Lead) {
	if !s.mockAutoreply || !s.whatsapp.IsMock() || lead == nil {
		return
	}
	from := toE164(lead.WhatsApp)
	if from == "" {
		return
	}
	go func() {
		time.Sleep(2 * time.Second)
		payload := MockInboundPayload(from, "Perfeito, obrigado! Vou conferir e te retorno.")
		_ = s.HandleInbound(context.Background(), payload)
	}()
}

// MockInboundPayload builds a webhook payload that looks like a lead message.
// Exported so the CLI simulator can inject inbound messages in dev.
func MockInboundPayload(from, text string) domain.WhatsAppWebhookPayload {
	raw := map[string]interface{}{
		"entry": []interface{}{
			map[string]interface{}{
				"id": "mock-entry",
				"changes": []interface{}{
					map[string]interface{}{
						"field": "messages",
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"metadata": map[string]interface{}{
								"display_phone_number": "5500000000000",
								"phone_number_id":      "12345",
							},
							"contacts": []interface{}{
								map[string]interface{}{"wa_id": from, "profile": map[string]interface{}{"name": "Lead"}},
							},
							"messages": []interface{}{
								map[string]interface{}{
									"from":      from,
									"id":        fmt.Sprintf("wamid-mock-inbound-%d", time.Now().UnixNano()),
									"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
									"type":      "text",
									"text":      map[string]interface{}{"body": text},
								},
							},
						},
					},
				},
			},
		},
	}

	var payload domain.WhatsAppWebhookPayload
	if b, err := json.Marshal(raw); err == nil {
		_ = json.Unmarshal(b, &payload)
	}
	return payload
}

// normalizeStatus maps Meta statuses onto the message model.
func normalizeStatus(s string) string {
	switch s {
	case "delivered":
		return domain.MessageStatusDelivered
	case "read":
		return domain.MessageStatusRead
	default:
		return domain.MessageStatusSent
	}
}

// toE164 converts a stored number (digits) to the E.164 form expected by the
// WhatsApp API (country code included, no "+").
func toE164(number string) string {
	d := digitsOnly(number)
	switch {
	case len(d) >= 12:
		return d
	case len(d) == 10 || len(d) == 11:
		return "55" + d
	default:
		return d
	}
}

// phoneCandidates builds the forms used to match an inbound wa_id to a stored
// lead number (E.164, national and trailing suffixes).
func phoneCandidates(number string) []string {
	d := digitsOnly(number)
	if d == "" {
		return nil
	}

	var out []string
	if len(d) >= 12 {
		out = append(out, d)
	}
	if len(d) >= 10 {
		national := d
		if strings.HasPrefix(d, "55") && (len(d) == 12 || len(d) == 13) {
			national = d[2:]
		}
		out = append(out, national)
		if len(d) > 11 {
			out = append(out, d[len(d)-11:])
		}
		if len(d) > 10 {
			out = append(out, d[len(d)-10:])
		}
	}
	return dedupe(out)
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
