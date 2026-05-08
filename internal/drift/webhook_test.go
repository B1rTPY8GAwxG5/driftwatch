package drift

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func webhookDriftResult() DriftResult {
	return DriftResult{
		Service: "api",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		},
	}
}

func TestNewWebhookSender_NotNil(t *testing.T) {
	s := NewWebhookSender(WebhookConfig{URL: "http://example.com"})
	if s == nil {
		t.Fatal("expected non-nil WebhookSender")
	}
}

func TestNewWebhookSender_DefaultTimeout(t *testing.T) {
	s := NewWebhookSender(WebhookConfig{URL: "http://example.com"})
	if s.client.Timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", s.client.Timeout)
	}
}

func TestWebhookSender_Send_Success(t *testing.T) {
	var received WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewWebhookSender(WebhookConfig{URL: server.URL})
	result := webhookDriftResult()
	if err := sender.Send(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Service != "api" {
		t.Errorf("expected service 'api', got %q", received.Service)
	}
	if !received.Drifted {
		t.Error("expected drifted=true")
	}
	if len(received.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(received.Entries))
	}
}

func TestWebhookSender_Send_CustomHeaders(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Token")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sender := NewWebhookSender(WebhookConfig{
		URL:     server.URL,
		Headers: map[string]string{"X-Token": "secret"},
	})
	if err := sender.Send(webhookDriftResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHeader != "secret" {
		t.Errorf("expected header 'secret', got %q", gotHeader)
	}
}

func TestWebhookSender_Send_Non2xx_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewWebhookSender(WebhookConfig{URL: server.URL})
	if err := sender.Send(webhookDriftResult()); err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestWebhookSender_Send_InvalidURL_ReturnsError(t *testing.T) {
	sender := NewWebhookSender(WebhookConfig{URL: "http://127.0.0.1:0"})
	if err := sender.Send(webhookDriftResult()); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
