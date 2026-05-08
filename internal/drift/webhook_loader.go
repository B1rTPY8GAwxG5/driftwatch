package drift

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type webhookConfigFile struct {
	Webhook struct {
		URL     string            `yaml:"url"`
		Headers map[string]string `yaml:"headers"`
		Timeout string            `yaml:"timeout"`
	} `yaml:"webhook"`
}

// LoadWebhookConfig reads a WebhookConfig from a YAML file.
func LoadWebhookConfig(path string) (WebhookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WebhookConfig{}, fmt.Errorf("webhook config: read file: %w", err)
	}
	return LoadWebhookConfigFromBytes(data)
}

// LoadWebhookConfigFromBytes parses a WebhookConfig from raw YAML bytes.
func LoadWebhookConfigFromBytes(data []byte) (WebhookConfig, error) {
	var f webhookConfigFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return WebhookConfig{}, fmt.Errorf("webhook config: parse yaml: %w", err)
	}
	if f.Webhook.URL == "" {
		return WebhookConfig{}, fmt.Errorf("webhook config: url is required")
	}
	cfg := WebhookConfig{
		URL:     f.Webhook.URL,
		Headers: f.Webhook.Headers,
	}
	if f.Webhook.Timeout != "" {
		d, err := time.ParseDuration(f.Webhook.Timeout)
		if err != nil {
			return WebhookConfig{}, fmt.Errorf("webhook config: parse timeout: %w", err)
		}
		cfg.Timeout = d
	}
	return cfg, nil
}
