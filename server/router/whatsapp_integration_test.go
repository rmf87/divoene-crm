package router_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmf87/divoene/internal/core/services"
	"github.com/rmf87/divoene/internal/infra/auth"
	"github.com/rmf87/divoene/internal/infra/clients"
	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/rmf87/divoene/internal/infra/storage"
	"github.com/rmf87/divoene/router"
)

const (
	testAppSecret   = "test-app-secret"
	testPhoneNumber = "11988887777" // stored national form
	testE164        = "5511988887777"
	testJWTSecret   = "test-jwt-secret"
)

// fakeGraphAPI records WhatsApp Cloud API sends and returns a stable wamid.
func fakeGraphAPI(t *testing.T, sent *[]map[string]interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/12345/messages" {
			t.Errorf("fake graph unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("fake graph bad auth header %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("fake graph bad body: %v", err)
		}
		*sent = append(*sent, req)
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.out-1"}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func signPayload(appSecret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func newTestAPI(t *testing.T) (string, *[]map[string]interface{}) {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}
	if err := auth.SeedUsers(db, auth.AdminCredentials{Email: "admin@test.local", Password: "test-password"}); err != nil {
		t.Fatalf("SeedUsers: %v", err)
	}

	sent := &[]map[string]interface{}{}
	graph := fakeGraphAPI(t, sent)

	leadRepo := database.NewLeadRepository(db)
	contractRepo := database.NewContractRepository(db)
	paymentRepo := database.NewPaymentRepository(db)
	visitRepo := database.NewVisitRepository(db)
	userRepo := database.NewUserRepository(db)

	configSvc := services.NewConfigService(database.NewConfigRepository(db, testJWTSecret), map[string]string{})
	clicksignClient := clients.NewClicksignClient("", "", "", "")
	openpixClient := clients.NewOpenPixClient("", "")
	whatsappClient := clients.NewWhatsAppClient(graph.URL, "test-token", "12345", testAppSecret, "verifytok")

	jwtMW, err := auth.SetupJWTAuth(db, userRepo, testJWTSecret)
	if err != nil {
		t.Fatalf("SetupJWTAuth: %v", err)
	}

	cfg := &router.Config{
		DB:              db,
		LeadService:     services.NewLeadService(leadRepo),
		ContractService: services.NewContractService(contractRepo, clicksignClient),
		PaymentService:  services.NewPaymentService(paymentRepo, openpixClient),
		VisitService:    services.NewVisitService(visitRepo),
		ConfigService:   configSvc,
		UserService:     services.NewUserService(userRepo),
		BackupService:   services.NewBackupService(storage.NewSnapshotStore(db, "/data/x.sqlite3"), "/data/x.sqlite3", nil),
		ChatService:     services.NewChatService(database.NewChatMessageRepository(db), leadRepo, whatsappClient, false),
		JWTMiddleware:   jwtMW,
		ClicksignClient: clicksignClient,
		OpenPixClient:   openpixClient,
		WhatsAppClient:  whatsappClient,
	}
	engine := router.SetupRouter(cfg)
	api := httptest.NewServer(engine)
	t.Cleanup(api.Close)
	return api.URL, sent
}

func login(t *testing.T, api string) string {
	t.Helper()
	res, err := http.Post(api+"/api/auth/login", "application/json",
		bytes.NewBufferString(`{"email":"admin@test.local","password":"test-password"}`))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("login status %d: %s", res.StatusCode, b)
	}
	var body struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	return body.Token
}

func TestWhatsAppRoundTrip(t *testing.T) {
	api, sent := newTestAPI(t)
	token := login(t, api)

	// 1. Create lead (public route).
	create := fmt.Sprintf(`{"name":"Maria","whatsapp":"%s","product":"buffet_infantil","source":"site"}`, testPhoneNumber)
	res, err := http.Post(api+"/api/leads", "application/json", bytes.NewBufferString(create))
	if err != nil {
		t.Fatalf("create lead: %v", err)
	}
	var lead struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&lead)
	res.Body.Close()
	if lead.ID == "" {
		t.Fatal("lead id empty")
	}

	// 2. Send a free-form message → Graph API must receive E.164 + text.
	send := fmt.Sprintf(`{"body":"Olá! Aqui é o vendedor da Chácara Divoene."}`)
	req, _ := http.NewRequest(http.MethodPost, api+"/api/leads/"+lead.ID+"/messages",
		bytes.NewBufferString(send))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	var sentMsg struct {
		WAMessageID string `json:"wa_message_id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&sentMsg)
	res.Body.Close()
	if res.StatusCode != http.StatusCreated || sentMsg.WAMessageID != "wamid.out-1" {
		t.Fatalf("send status %d wa=%q", res.StatusCode, sentMsg.WAMessageID)
	}
	if len(*sent) != 1 {
		t.Fatalf("expected 1 Graph call, got %d", len(*sent))
	}
	payload := (*sent)[0]
	if payload["to"] != testE164 {
		t.Errorf("Graph to = %v, want %s", payload["to"], testE164)
	}
	if payload["type"] != "text" {
		t.Errorf("Graph type = %v, want text", payload["type"])
	}

	// 3. Inbound webhook (signed) → routed to the lead.
	inbound := services.MockInboundPayload(testE164, "Quero agendar uma visita")
	ib, _ := json.Marshal(inbound)
	req, _ = http.NewRequest(http.MethodPost, api+"/api/whatsapp/webhook", bytes.NewReader(ib))
	req.Header.Set("X-Hub-Signature-256", signPayload(testAppSecret, ib))
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("webhook inbound: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("webhook inbound status %d", res.StatusCode)
	}

	// 4. List → 2 messages, last is from lead.
	req, _ = http.NewRequest(http.MethodGet, api+"/api/leads/"+lead.ID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var msgs []map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&msgs)
	res.Body.Close()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[1]["direction"] != "lead" || msgs[1]["body"] != "Quero agendar uma visita" {
		t.Errorf("inbound not routed: %+v", msgs[1])
	}

	// 5. Status webhook (delivered) updates the outbound message.
	status := `{"entry":[{"id":"x","changes":[{"field":"messages","value":{"statuses":[{"id":"wamid.out-1","status":"delivered"}]}}]}]}`
	req, _ = http.NewRequest(http.MethodPost, api+"/api/whatsapp/webhook", bytes.NewBufferString(status))
	req.Header.Set("X-Hub-Signature-256", signPayload(testAppSecret, []byte(status)))
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("webhook status: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("webhook status %d", res.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, api+"/api/leads/"+lead.ID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ = http.DefaultClient.Do(req)
	var msgs2 []map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&msgs2)
	res.Body.Close()
	if msgs2[0]["status"] != "delivered" {
		t.Errorf("outbound status = %v, want delivered", msgs2[0]["status"])
	}

	// 6. Webhook with invalid signature → 403.
	req, _ = http.NewRequest(http.MethodPost, api+"/api/whatsapp/webhook", bytes.NewBufferString(status))
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	res, _ = http.DefaultClient.Do(req)
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("invalid signature status %d, want 403", res.StatusCode)
	}

	// 7. List messages for unknown lead → 404 (lead not found).
	req, _ = http.NewRequest(http.MethodGet, api+"/api/leads/UNKNOWN/messages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ = http.DefaultClient.Do(req)
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("unknown lead list status %d, want 404", res.StatusCode)
	}
}

func TestWhatsAppWebhookVerification(t *testing.T) {
	api, _ := newTestAPI(t)

	// Valid verify token echoes the challenge.
	res, err := http.Get(api + "/api/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=verifytok&hub.challenge=abc123")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || string(body) != "abc123" {
		t.Errorf("verify status %d body %q", res.StatusCode, body)
	}

	// Wrong verify token → 403.
	res, err = http.Get(api + "/api/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=abc123")
	if err != nil {
		t.Fatalf("verify wrong: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("wrong verify status %d, want 403", res.StatusCode)
	}
}

func TestWhatsAppSendRequiresAuth(t *testing.T) {
	api, _ := newTestAPI(t)
	res, err := http.Post(api+"/api/leads/L001/messages", "application/json",
		bytes.NewBufferString(`{"body":"oi"}`))
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-auth send status %d, want 401", res.StatusCode)
	}
}
