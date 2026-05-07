package drift

import (
	"testing"
	"time"
)

func baseSpec() ServiceSpec {
	return ServiceSpec{
		Name:     "svc",
		Image:    "app:v1",
		Replicas: 2,
		Env:      map[string]string{"KEY": "val"},
	}
}

func TestNewComparator_NotNil(t *testing.T) {
	d := NewDetector()
	c := NewComparator(d)
	if c == nil {
		t.Fatal("expected non-nil Comparator")
	}
}

func TestComparator_Compare_NoDrift(t *testing.T) {
	d := NewDetector()
	c := NewComparator(d)
	spec := baseSpec()
	result, err := c.Compare(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasDrift() {
		t.Errorf("expected no drift, got %d entries", len(result.Entries))
	}
}

func TestComparator_Compare_WithFilter(t *testing.T) {
	d := NewDetector()
	f := NewFilter([]DriftKind{KindImage})
	c := NewComparator(d, WithFilter(f))

	spec := baseSpec()
	live := baseSpec()
	live.Image = "app:v2"
	live.Replicas = 5

	result, err := c.Compare(spec, live)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range result.Entries {
		if e.Kind == KindImage {
			t.Errorf("expected image drift to be filtered out")
		}
	}
}

func TestComparator_Compare_WithSuppression(t *testing.T) {
	d := NewDetector()
	store := NewSuppressionStore()
	store.Add(SuppressionRule{
		Service:   "svc",
		Kind:      KindReplicas,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	c := NewComparator(d, WithSuppression(store))

	spec := baseSpec()
	live := baseSpec()
	live.Replicas = 99

	result, err := c.Compare(spec, live)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range result.Entries {
		if e.Kind == KindReplicas {
			t.Errorf("expected replicas drift to be suppressed")
		}
	}
}

func TestComparator_Compare_WithPolicy(t *testing.T) {
	d := NewDetector()
	p := &Policy{
		Rules: []PolicyRule{
			{Kind: KindImage, Action: "block"},
		},
	}
	c := NewComparator(d, WithPolicy(p))

	spec := baseSpec()
	live := baseSpec()
	live.Image = "app:evil"

	result, err := c.Compare(spec, live)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.policy.Blocked(result) {
		t.Errorf("expected policy to block result")
	}
}
