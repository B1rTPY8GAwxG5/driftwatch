package drift

import (
	"testing"
)

func TestDriftResult_HasDrift_False(t *testing.T) {
	r := DriftResult{
		ServiceName: "svc",
		Entries:     []DriftEntry{},
	}
	if r.HasDrift() {
		t.Error("expected HasDrift() to be false for empty entries")
	}
}

func TestDriftResult_HasDrift_True(t *testing.T) {
	r := DriftResult{
		ServiceName: "svc",
		Entries: []DriftEntry{
			{Kind: DriftKindImage, Field: "image", Expected: "nginx:1.25", Actual: "nginx:1.24"},
		},
	}
	if !r.HasDrift() {
		t.Error("expected HasDrift() to be true when entries exist")
	}
}

func TestDriftKind_Constants(t *testing.T) {
	cases := []struct {
		kind     DriftKind
		expected string
	}{
		{DriftKindImage, "image"},
		{DriftKindReplicas, "replicas"},
		{DriftKindEnv, "env"},
		{DriftKindPort, "port"},
		{DriftKindMissing, "missing"},
	}
	for _, tc := range cases {
		if string(tc.kind) != tc.expected {
			t.Errorf("DriftKind %q: expected %q, got %q", tc.kind, tc.expected, string(tc.kind))
		}
	}
}

func TestDriftEntry_Fields(t *testing.T) {
	entry := DriftEntry{
		Kind:     DriftKindEnv,
		Field:    "DATABASE_URL",
		Expected: "postgres://prod",
		Actual:   "",
	}
	if entry.Kind != DriftKindEnv {
		t.Errorf("unexpected kind: %v", entry.Kind)
	}
	if entry.Field != "DATABASE_URL" {
		t.Errorf("unexpected field: %v", entry.Field)
	}
}

func TestServiceSpec_Defaults(t *testing.T) {
	spec := ServiceSpec{}
	if spec.Name != "" || spec.Image != "" || spec.Replicas != 0 {
		t.Error("expected zero-value ServiceSpec to have empty defaults")
	}
}
