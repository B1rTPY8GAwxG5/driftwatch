package drift

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWebhookConfigFromBytes_Valid(t *testing.T) {
	yaml := []byte(`
webhook:
  url: "https://hooks.example.com/drift"
  timeout: "5s"
  headers:
    Authorization: "Bearer token123"
`)
	cfg, err := LoadWebhookConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.URL != "https://hooks.example.com/drift" {
		t.Errorf("unexpected URL: %s", cfg.URL)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("unexpected timeout: %v", cfg.Timeout)
	}
	if cfg.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("unexpected header: %v", cfg.Headers)
	}
}

func TestLoadWebhookConfigFromBytes_MissingURL(t *testing.T) {
	yaml := []byte(`webhook:
  timeout: "5s"
`)
	_, err := LoadWebhookConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestLoadWebhookConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadWebhookConfigFromBytes([]byte(":::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadWebhookConfigFromBytes_InvalidTimeout(t *testing.T) {
	yaml := []byte(`webhook:
  url: "https://example.com"
  timeout: "notaduration"
`)
	_, err := LoadWebhookConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for invalid timeout")
	}
}

func TestLoadWebhookConfig_FileNotFound(t *testing.T) {
	_, err := LoadWebhookConfig("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadWebhookConfig_ValidFile(t *testing.T) {
	content := []byte(`webhook:
  url: "https://hooks.example.com/alert"
`)
	tmp := filepath.Join(t.TempDir(), "webhook.yaml")
	if err := os.WriteFile(tmp, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg, err := LoadWebhookConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.URL != "https://hooks.example.com/alert" {
		t.Errorf("unexpected URL: %s", cfg.URL)
	}
}
