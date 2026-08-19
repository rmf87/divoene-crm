package clients

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestWhatsAppClientSendText(t *testing.T) {
	base := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/12345/messages" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer tok" {
			t.Errorf("auth header: %q", auth)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req["to"] != "5511999998888" {
			t.Errorf("to: %v", req["to"])
		}
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.abc123"}]}`))
	})

	c := NewWhatsAppClient(base, "tok", "12345", "secret", "vt")
	id, err := c.SendText(context.Background(), "5511999998888", "Olá!")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if id != "wamid.abc123" {
		t.Errorf("got id %q", id)
	}
}

func TestWhatsAppClientSendTemplate(t *testing.T) {
	base := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req["type"] != "template" {
			t.Errorf("type: %v", req["type"])
		}
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.tpl1"}]}`))
	})

	c := NewWhatsAppClient(base, "tok", "12345", "secret", "vt")
	id, err := c.SendTemplate(context.Background(), "5511999998888", "visit_confirmation", "pt_BR",
		map[string]string{"1": "João"})
	if err != nil {
		t.Fatalf("send template: %v", err)
	}
	if id != "wamid.tpl1" {
		t.Errorf("got id %q", id)
	}
}

func TestWhatsAppClientMockMode(t *testing.T) {
	c := NewWhatsAppClient("", "", "", "", "")
	if !c.IsMock() {
		t.Error("expected mock mode with empty token")
	}
	id, err := c.SendText(context.Background(), "5511999998888", "oi")
	if err != nil || !strings.HasPrefix(id, "wamid-mock-") {
		t.Errorf("mock send: id=%q err=%v", id, err)
	}
}

func TestWhatsAppClientVerifySignature(t *testing.T) {
	c := NewWhatsAppClient("", "tok", "12345", "appsecret", "vt")

	payload := []byte(`{"entry":[]}`)
	mac := hmac.New(sha256.New, []byte("appsecret"))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !c.VerifySignature(payload, sig) {
		t.Error("valid signature rejected")
	}
	if c.VerifySignature(payload, "sha256=deadbeef") {
		t.Error("invalid signature accepted")
	}
	if c.VerifySignature(payload, "") {
		t.Error("empty signature accepted")
	}
}

func TestWhatsAppClientReload(t *testing.T) {
	c := NewWhatsAppClient("", "old", "oldid", "oldsecret", "oldvt")
	c.Reload(context.Background(), fakeProvider{
		"whatsapp_token":                "newtok",
		"whatsapp_phone_number_id":      "999",
		"whatsapp_app_secret":           "newsecret",
		"whatsapp_webhook_verify_token": "newvt",
		"whatsapp_base_url":             "http://example.test",
	})
	if c.token != "newtok" || c.phoneNumberID != "999" || c.baseURL != "http://example.test" {
		t.Errorf("reload failed: %+v", c)
	}
	if c.GetVerifyToken() != "newvt" {
		t.Errorf("verify token not reloaded: %q", c.GetVerifyToken())
	}
}

type fakeProvider map[string]string

func (f fakeProvider) GetDecryptedValue(_ context.Context, key string) (string, error) {
	return f[key], nil
}
