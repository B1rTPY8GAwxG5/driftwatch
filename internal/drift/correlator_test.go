package drift

import (
	"testing"
	"time"
)

func driftedResultForCorrelator(service string, kinds ...DriftKind) DriftResult {
	entries := make([]DriftEntry, 0, len(kinds))
	for _, k := range kinds {
		entries = append(entries, DriftEntry{Kind: k, Field: string(k), Expected: "a", Actual: "b"})
	}
	return DriftResult{Service: service, Entries: entries}
}

func TestNewCorrelator_NotNil(t *testing.T) {
	c := NewCorrelator()
	if c == nil {
		t.Fatal("expected non-nil Correlator")
	}
}

func TestCorrelator_Ingest_CleanResult_NoGroups(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(DriftResult{Service: "svc-a"})
	if len(c.Groups()) != 0 {
		t.Errorf("expected 0 groups for clean result, got %d", len(c.Groups()))
	}
}

func TestCorrelator_Ingest_SingleService_CreatesGroup(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage))
	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Kind != KindImage {
		t.Errorf("expected kind %s, got %s", KindImage, groups[0].Kind)
	}
	if len(groups[0].Services) != 1 || groups[0].Services[0] != "svc-a" {
		t.Errorf("unexpected services: %v", groups[0].Services)
	}
}

func TestCorrelator_Ingest_MultipleServices_SameKind(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage))
	c.Ingest(driftedResultForCorrelator("svc-b", KindImage))
	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Services) != 2 {
		t.Errorf("expected 2 services in group, got %d", len(groups[0].Services))
	}
}

func TestCorrelator_Ingest_DuplicateService_NotDoubleAdded(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage))
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage))
	groups := c.Groups()
	if len(groups[0].Services) != 1 {
		t.Errorf("expected 1 unique service, got %d", len(groups[0].Services))
	}
}

func TestCorrelator_Ingest_MultipleKinds_MultipleGroups(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage, KindReplicas))
	groups := c.Groups()
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestCorrelator_Reset_ClearsGroups(t *testing.T) {
	c := NewCorrelator()
	c.Ingest(driftedResultForCorrelator("svc-a", KindImage))
	c.Reset()
	if len(c.Groups()) != 0 {
		t.Errorf("expected 0 groups after reset, got %d", len(c.Groups()))
	}
}

func TestCorrelationGroup_Summary_ContainsKind(t *testing.T) {
	g := CorrelationGroup{
		ID:       "corr-image",
		Kind:     KindImage,
		Services: []string{"svc-a"},
		Detected: time.Now(),
	}
	s := g.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
	for _, sub := range []string{"corr-image", string(KindImage)} {
		if !containsSubstr(s, sub) {
			t.Errorf("summary missing %q: %s", sub, s)
		}
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
