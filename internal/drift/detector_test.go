package drift

import (
	"testing"
)

func baseService() ServiceConfig {
	return ServiceConfig{
		Name:     "api",
		Image:    "myapp:v1.2.0",
		Replicas: 3,
		Environment: map[string]string{
			"LOG_LEVEL": "info",
			"PORT":      "8080",
		},
		Labels: map[string]string{
			"team": "platform",
		},
	}
}

func TestCompare_NoDrift(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()

	results := d.Compare(declared, live)
	if len(results) != 0 {
		t.Errorf("expected no drift, got %d results", len(results))
	}
}

func TestCompare_ImageDrift(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()
	live.Image = "myapp:v1.1.0"

	results := d.Compare(declared, live)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if results[0].Field != "image" {
		t.Errorf("expected field 'image', got '%s'", results[0].Field)
	}
	if results[0].Status != StatusDrifted {
		t.Errorf("expected status drifted, got %s", results[0].Status)
	}
}

func TestCompare_ReplicasDrift(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()
	live.Replicas = 1

	results := d.Compare(declared, live)
	if len(results) != 1 || results[0].Field != "replicas" {
		t.Errorf("expected replicas drift, got %+v", results)
	}
}

func TestCompare_MissingEnvVar(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()
	delete(live.Environment, "LOG_LEVEL")

	results := d.Compare(declared, live)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusMissing {
		t.Errorf("expected status missing, got %s", results[0].Status)
	}
}

func TestCompare_MultipleDrifts(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()
	live.Image = "myapp:latest"
	live.Replicas = 5
	live.Environment["LOG_LEVEL"] = "debug"

	results := d.Compare(declared, live)
	if len(results) != 3 {
		t.Errorf("expected 3 drift results, got %d", len(results))
	}
}

func TestCompare_MissingLabel(t *testing.T) {
	d := NewDetector()
	declared := baseService()
	live := baseService()
	delete(live.Labels, "team")

	results := d.Compare(declared, live)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusMissing {
		t.Errorf("expected status missing, got %s", results[0].Status)
	}
}
