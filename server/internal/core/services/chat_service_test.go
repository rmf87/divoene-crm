package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

type stubChatRepo struct {
	msgs []*domain.ChatMessage
}

func (r *stubChatRepo) Store(_ context.Context, msg *domain.ChatMessage) error {
	r.msgs = append(r.msgs, msg)
	return nil
}

func (r *stubChatRepo) ListByLead(_ context.Context, leadID string) ([]*domain.ChatMessage, error) {
	var out []*domain.ChatMessage
	for _, m := range r.msgs {
		if m.LeadID == leadID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *stubChatRepo) FindByWAMessageID(_ context.Context, waID string) (*domain.ChatMessage, error) {
	for _, m := range r.msgs {
		if m.WAMessageID == waID {
			return m, nil
		}
	}
	return nil, errors.New("message not found")
}

func (r *stubChatRepo) UpdateStatus(_ context.Context, waID, status string) error {
	for _, m := range r.msgs {
		if m.WAMessageID == waID {
			m.Status = status
		}
	}
	return nil
}

type stubLeadRepo struct {
	leads map[string]*domain.Lead
}

func (r *stubLeadRepo) Store(_ context.Context, l *domain.Lead) error { return nil }
func (r *stubLeadRepo) Get(_ context.Context, id string) (*domain.Lead, error) {
	if l, ok := r.leads[id]; ok {
		return l, nil
	}
	return nil, errors.New("lead não encontrado")
}
func (r *stubLeadRepo) List(_ context.Context, _ map[string]string) ([]*domain.Lead, error) {
	var out []*domain.Lead
	for _, l := range r.leads {
		out = append(out, l)
	}
	return out, nil
}
func (r *stubLeadRepo) FindByWhatsApp(_ context.Context, numbers []string) (*domain.Lead, error) {
	for _, l := range r.leads {
		for _, n := range numbers {
			if strings.Contains(l.WhatsApp, n) || strings.Contains(n, l.WhatsApp) {
				return l, nil
			}
		}
	}
	return nil, errors.New("lead não encontrado")
}
func (r *stubLeadRepo) UpdateStage(_ context.Context, _, _, _ string) error { return nil }
func (r *stubLeadRepo) UpdateValidation(_ context.Context, _ string, _ *domain.EventInfo, _ *domain.ContactPersonInfo, _ []domain.AddOnItem) error {
	return nil
}
func (r *stubLeadRepo) AddNote(_ context.Context, _ string, _ domain.Note) error { return nil }

type stubWhatsApp struct{}

func (w *stubWhatsApp) IsMock() bool { return true }
func (w *stubWhatsApp) SendText(_ context.Context, to, body string) (string, error) {
	return "wamid-send-" + to, nil
}
func (w *stubWhatsApp) SendTemplate(_ context.Context, to, name, lang string, _ map[string]string) (string, error) {
	return "wamid-tpl-" + to + "-" + name, nil
}
func (w *stubWhatsApp) VerifySignature(_ []byte, _ string) bool { return true }

func newChatHarness() (*ChatService, *stubChatRepo, *stubLeadRepo) {
	leads := &stubLeadRepo{leads: map[string]*domain.Lead{
		"L1": {ID: "L1", Name: "Maria", WhatsApp: "11999998888"},
	}}
	repo := &stubChatRepo{}
	return NewChatService(repo, leads, &stubWhatsApp{}, false), repo, leads
}

func TestChatServiceSendMessage(t *testing.T) {
	svc, repo, _ := newChatHarness()
	msg, err := svc.SendMessage(context.Background(), "L1", "Olá!")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if msg.Direction != domain.DirectionSeller || msg.Status != domain.MessageStatusSent {
		t.Errorf("unexpected msg: %+v", msg)
	}
	if msg.WAMessageID != "wamid-send-5511999998888" {
		t.Errorf("wa id: %q", msg.WAMessageID)
	}
	if len(repo.msgs) != 1 {
		t.Errorf("expected 1 stored, got %d", len(repo.msgs))
	}
}

func TestChatServiceSendMessageEmpty(t *testing.T) {
	svc, _, _ := newChatHarness()
	if _, err := svc.SendMessage(context.Background(), "L1", "  "); err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestChatServiceSendMessageUnknownLead(t *testing.T) {
	svc, _, _ := newChatHarness()
	if _, err := svc.SendMessage(context.Background(), "NOPE", "oi"); err == nil {
		t.Fatal("expected error for unknown lead")
	}
}

func TestChatServiceSendTemplate(t *testing.T) {
	svc, _, _ := newChatHarness()
	msg, err := svc.SendTemplate(context.Background(), "L1", "visit_confirmation", "pt_BR", nil)
	if err != nil {
		t.Fatalf("send template: %v", err)
	}
	if msg.WAMessageID != "wamid-tpl-5511999998888-visit_confirmation" {
		t.Errorf("wa id: %q", msg.WAMessageID)
	}
	if _, err := svc.SendTemplate(context.Background(), "L1", "", "pt_BR", nil); err == nil {
		t.Fatal("expected error for empty template")
	}
}

func TestChatServiceHandleInbound(t *testing.T) {
	svc, repo, _ := newChatHarness()

	payload := domain.WhatsAppWebhookPayload{
		Entry: []struct {
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
		}{
			{Changes: []struct {
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
			}{
				{Value: struct {
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
				}{
					Messages: []struct {
						From      string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
					}{
						{From: "5511999998888", ID: "wamid.in1", Type: "text", Text: struct {
							Body string `json:"body"`
						}{Body: "Quero agendar uma visita"}},
					},
				},
				}},
			}},
	}

	if err := svc.HandleInbound(context.Background(), payload); err != nil {
		t.Fatalf("handle inbound: %v", err)
	}
	if len(repo.msgs) != 1 {
		t.Fatalf("expected 1 stored, got %d", len(repo.msgs))
	}
	if repo.msgs[0].Direction != domain.DirectionLead {
		t.Errorf("direction: %q", repo.msgs[0].Direction)
	}
	if repo.msgs[0].LeadID != "L1" {
		t.Errorf("routed to wrong lead: %q", repo.msgs[0].LeadID)
	}
}

func TestChatServiceHandleInboundUnknown(t *testing.T) {
	svc, repo, _ := newChatHarness()
	payload := domain.WhatsAppWebhookPayload{
		Entry: []struct {
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
		}{
			{Changes: []struct {
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
			}{
				{Value: struct {
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
				}{
					Messages: []struct {
						From      string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
					}{
						{From: "55115550000", ID: "wamid.x", Type: "text", Text: struct {
							Body string `json:"body"`
						}{Body: "oi"}},
					},
				},
				}},
			}},
	}

	if err := svc.HandleInbound(context.Background(), payload); err != nil {
		t.Fatalf("handle inbound: %v", err)
	}
	if len(repo.msgs) != 0 {
		t.Errorf("unknown number should not store messages, got %d", len(repo.msgs))
	}
}

func TestPhoneCandidates(t *testing.T) {
	cases := map[string]int{
		"5511999998888": 3,
		"11999998888":   2,
		"99999":         0,
	}
	for in, want := range cases {
		got := phoneCandidates(in)
		if len(got) != want {
			t.Errorf("phoneCandidates(%q) = %d entries, want %d", in, len(got), want)
		}
	}
}

func TestToE164(t *testing.T) {
	cases := map[string]string{
		"11999998888":   "5511999998888",
		"5511999998888": "5511999998888",
		"62 99999-8888": "5562999998888",
	}
	for in, want := range cases {
		if got := toE164(in); got != want {
			t.Errorf("toE164(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChatServiceMockAutoreply(t *testing.T) {
	leads := &stubLeadRepo{leads: map[string]*domain.Lead{
		"L1": {ID: "L1", Name: "Maria", WhatsApp: "11999998888"},
	}}
	repo := &stubChatRepo{}
	svc := NewChatService(repo, leads, &stubWhatsApp{}, true)

	if _, err := svc.SendMessage(context.Background(), "L1", "Olá!"); err != nil {
		t.Fatalf("send: %v", err)
	}
	// Autoreply fires after ~2s. Give it time.
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		list, _ := repo.ListByLead(context.Background(), "L1")
		if len(list) >= 2 {
			if list[1].Direction != domain.DirectionLead {
				t.Fatalf("expected lead reply, got %+v", list[1])
			}
			if list[1].Body == "" {
				t.Fatalf("expected reply body, got empty")
			}
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("autoreply never arrived")
}

func TestChatServiceMockAutoreplyDisabled(t *testing.T) {
	leads := &stubLeadRepo{leads: map[string]*domain.Lead{
		"L1": {ID: "L1", Name: "Maria", WhatsApp: "11999998888"},
	}}
	repo := &stubChatRepo{}
	svc := NewChatService(repo, leads, &stubWhatsApp{}, false)

	if _, err := svc.SendMessage(context.Background(), "L1", "Olá!"); err != nil {
		t.Fatalf("send: %v", err)
	}
	time.Sleep(2500 * time.Millisecond)
	list, _ := repo.ListByLead(context.Background(), "L1")
	if len(list) != 1 {
		t.Fatalf("expected only the outbound message, got %d", len(list))
	}
}
