package drift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to a webhook endpoint.
type WebhookPayload struct {
	Service   string       `json:"service"`
	Drifted   bool         `json:"drifted"`
	Timestamp time.Time    `json:"timestamp"`
	Entries   []DriftEntry `json:"entries,omitempty"`
}

// WebhookConfig holds configuration for a webhook notifier.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout time.Duration     `yaml:"timeout"`
}

// WebhookSender sends drift results to an HTTP endpoint.
type WebhookSender struct {
	cfg    WebhookConfig
	client *http.Client
}

// NewWebhookSender creates a WebhookSender with the given config.
func NewWebhookSender(cfg WebhookConfig) *WebhookSender {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSender{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

// Send delivers the drift result as a JSON payload to the configured URL.
func (w *WebhookSender) Send(result DriftResult) error {
	payload := WebhookPayload{
		Service:   result.Service,
		Drifted:   result.HasDrift(),
		Timestamp: time.Now().UTC(),
		Entries:   result.Entries,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, w.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.cfg.Headers {
		req.Header.Set(k, v)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
