package drift

import (
	"testing"
	"time"
)

func TestDefaultWindowPolicy_Values(t *testing.T) {
	p := DefaultWindowPolicy()
	if p.Size != 5*time.Minute {
		t.Errorf("expected 5m, got %v", p.Size)
	}
	if p.MaxItems != 100 {
		t.Errorf("expected 100, got %d", p.MaxItems)
	}
}

func TestWindowPolicy_Validate_Valid(t *testing.T) {
	p := DefaultWindowPolicy()
	if err := p.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWindowPolicy_Validate_ZeroSize(t *testing.T) {
	p := WindowPolicy{Size: 0, MaxItems: 10}
	if err := p.Validate(); err == nil {
		t.Error("expected error for zero size")
	}
}

func TestWindowPolicy_Validate_ZeroMaxItems(t *testing.T) {
	p := WindowPolicy{Size: time.Minute, MaxItems: 0}
	if err := p.Validate(); err == nil {
		t.Error("expected error for zero max_items")
	}
}

func TestNewSlidingWindow_NotNil(t *testing.T) {
	w, err := NewSlidingWindow(DefaultWindowPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil window")
	}
}

func TestNewSlidingWindow_ZeroValueUsesDefaults(t *testing.T) {
	w, err := NewSlidingWindow(WindowPolicy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil window")
	}
}

func TestNewSlidingWindow_InvalidPolicy(t *testing.T) {
	_, err := NewSlidingWindow(WindowPolicy{Size: -1, MaxItems: 10})
	if err == nil {
		t.Error("expected error for invalid policy")
	}
}

func TestSlidingWindow_Add_IncreasesLen(t *testing.T) {
	w, _ := NewSlidingWindow(DefaultWindowPolicy())
	w.Add(DriftResult{Service: "svc-a"})
	w.Add(DriftResult{Service: "svc-b"})
	if w.Len() != 2 {
		t.Errorf("expected 2, got %d", w.Len())
	}
}

func TestSlidingWindow_Results_ReturnsSnapshot(t *testing.T) {
	w, _ := NewSlidingWindow(DefaultWindowPolicy())
	w.Add(DriftResult{Service: "svc-a"})
	res := w.Results()
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Service != "svc-a" {
		t.Errorf("unexpected service: %s", res[0].Service)
	}
}

func TestSlidingWindow_MaxItems_Evicts(t *testing.T) {
	w, _ := NewSlidingWindow(WindowPolicy{Size: time.Hour, MaxItems: 3})
	for i := 0; i < 5; i++ {
		w.Add(DriftResult{Service: "svc"})
	}
	if w.Len() != 3 {
		t.Errorf("expected 3 after eviction, got %d", w.Len())
	}
}
