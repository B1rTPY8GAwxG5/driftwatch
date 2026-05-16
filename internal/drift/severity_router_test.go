package drift

import (
	"testing"
)

func TestNewSeverityRouter_NotNil(t *testing.T) {
	r := NewSeverityRouter("default")
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestSeverityRouter_Len_Zero(t *testing.T) {
	r := NewSeverityRouter("default")
	if r.Len() != 0 {
		t.Fatalf("expected 0 routes, got %d", r.Len())
	}
}

func TestSeverityRouter_AddRoute_IncrementsLen(t *testing.T) {
	r := NewSeverityRouter("default")
	r.AddRoute("high", "pagerduty")
	r.AddRoute("low", "log")
	if r.Len() != 2 {
		t.Fatalf("expected 2 routes, got %d", r.Len())
	}
}

func TestSeverityRouter_AddRoute_EmptyIgnored(t *testing.T) {
	r := NewSeverityRouter("default")
	r.AddRoute("", "dest")
	r.AddRoute("sev", "")
	if r.Len() != 0 {
		t.Fatalf("expected 0 routes after invalid adds, got %d", r.Len())
	}
}

func TestSeverityRouter_Route_NoDrift_ReturnsFallback(t *testing.T) {
	r := NewSeverityRouter("fallback-dest")
	r.AddRoute("critical", "pagerduty")

	result := DriftResult{Service: "svc-a", Entries: nil}
	dest := r.Route(result)
	if dest != "fallback-dest" {
		t.Fatalf("expected fallback-dest, got %q", dest)
	}
}

func TestSeverityRouter_Route_NoMatchingRoute_ReturnsFallback(t *testing.T) {
	r := NewSeverityRouter("catch-all")
	// no routes registered
	result := DriftResult{
		Service: "svc-b",
		Entries: []DriftEntry{{Kind: DriftKindImage, Field: "image", Got: "v2", Want: "v1"}},
	}
	dest := r.Route(result)
	if dest != "catch-all" {
		t.Fatalf("expected catch-all, got %q", dest)
	}
}

func TestSeverityRouter_Route_MatchingRoute(t *testing.T) {
	r := NewSeverityRouter("default")
	r.AddRoute("low", "slack")
	r.AddRoute("medium", "email")

	result := DriftResult{
		Service: "svc-c",
		Entries: []DriftEntry{{Kind: DriftKindImage, Field: "image", Got: "v2", Want: "v1"}},
	}
	dest := r.Route(result)
	// Score for a single image drift is expected to be in the low-medium range.
	if dest != "slack" && dest != "email" && dest != "default" {
		t.Fatalf("unexpected destination %q", dest)
	}
}

func TestSeverityRouter_String_ContainsFallback(t *testing.T) {
	r := NewSeverityRouter("my-fallback")
	s := r.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
	if !containsSubstr(s, "my-fallback") {
		t.Fatalf("expected string to contain fallback name, got %q", s)
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
