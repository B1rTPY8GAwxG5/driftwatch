package drift

import (
	"testing"
)

func sensitiveEntry() DriftEntry {
	return DriftEntry{
		Field: "env.SECRET_TOKEN",
		Kind:  KindEnv,
		Got:   "abc123",
		Want:  "xyz789",
	}
}

func TestNewRedactor_DefaultMask(t *testing.T) {
	r := NewRedactor([]string{"secret"}, "")
	if r.mask != "***" {
		t.Fatalf("expected default mask *** got %q", r.mask)
	}
}

func TestNewRedactor_CustomMask(t *testing.T) {
	r := NewRedactor([]string{"secret"}, "[REDACTED]")
	if r.mask != "[REDACTED]" {
		t.Fatalf("expected [REDACTED] got %q", r.mask)
	}
}

func TestRedactor_Redact_SensitiveField(t *testing.T) {
	r := NewRedactor([]string{"secret", "password"}, "***")
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{sensitiveEntry()},
	}
	out := r.Redact(result)
	if out.Entries[0].Got != "***" {
		t.Errorf("expected Got to be masked, got %q", out.Entries[0].Got)
	}
	if out.Entries[0].Want != "***" {
		t.Errorf("expected Want to be masked, got %q", out.Entries[0].Want)
	}
}

func TestRedactor_Redact_NonSensitiveField(t *testing.T) {
	r := NewRedactor([]string{"secret"}, "***")
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{{Field: "image", Kind: KindImage, Got: "v1", Want: "v2"}},
	}
	out := r.Redact(result)
	if out.Entries[0].Got != "v1" {
		t.Errorf("expected Got unchanged, got %q", out.Entries[0].Got)
	}
}

func TestRedactor_Redact_PreservesService(t *testing.T) {
	r := NewRedactor([]string{"secret"}, "***")
	result := DriftResult{Service: "my-service", Entries: []DriftEntry{}}
	out := r.Redact(result)
	if out.Service != "my-service" {
		t.Errorf("expected service name preserved, got %q", out.Service)
	}
}

func TestRedactor_WithFunc_Override(t *testing.T) {
	r := NewRedactor(nil, "***").WithFunc(func(key, value string) string {
		return "CUSTOM"
	})
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{{Field: "image", Got: "v1", Want: "v2"}},
	}
	out := r.Redact(result)
	if out.Entries[0].Got != "CUSTOM" {
		t.Errorf("expected CUSTOM from fn, got %q", out.Entries[0].Got)
	}
}

func TestRedactor_RedactAll_Length(t *testing.T) {
	r := NewRedactor([]string{"secret"}, "***")
	results := []DriftResult{
		{Service: "a", Entries: []DriftEntry{sensitiveEntry()}},
		{Service: "b", Entries: []DriftEntry{sensitiveEntry()}},
	}
	out := r.RedactAll(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	for _, res := range out {
		if res.Entries[0].Got != "***" {
			t.Errorf("expected masked entry in service %q", res.Service)
		}
	}
}

func TestRedactor_CaseInsensitiveKey(t *testing.T) {
	r := NewRedactor([]string{"SECRET"}, "***")
	result := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{{Field: "env.secret_key", Got: "val", Want: "other"}},
	}
	out := r.Redact(result)
	if out.Entries[0].Got != "***" {
		t.Errorf("expected case-insensitive match to mask value")
	}
}
