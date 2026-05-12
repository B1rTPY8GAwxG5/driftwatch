package drift

import (
	"testing"
)

var priorityDriftedResult = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Declared: "v1", Observed: "v2"},
		{Kind: DriftKindEnv, Field: "ENV_KEY", Declared: "a", Observed: "b"},
	},
}

var priorityCleanResult = DriftResult{
	Service: "worker",
	Entries: []DriftEntry{},
}

func TestNewPrioritizer_NotNil(t *testing.T) {
	p := NewPrioritizer()
	if p == nil {
		t.Fatal("expected non-nil Prioritizer")
	}
}

func TestPriority_String(t *testing.T) {
	cases := []struct {
		p    Priority
		want string
	}{
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{Priority(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("Priority(%d).String() = %q, want %q", tc.p, got, tc.want)
		}
	}
}

func TestPrioritizer_Prioritize_ImageDrift_High(t *testing.T) {
	p := NewPrioritizer()
	pr := p.Prioritize(priorityDriftedResult)
	if pr.Priority != PriorityHigh {
		t.Errorf("expected PriorityHigh, got %s", pr.Priority)
	}
}

func TestPrioritizer_Prioritize_CleanResult_Low(t *testing.T) {
	p := NewPrioritizer()
	pr := p.Prioritize(priorityCleanResult)
	if pr.Priority != PriorityLow {
		t.Errorf("expected PriorityLow for clean result, got %s", pr.Priority)
	}
}

func TestPrioritizer_SetWeight_OverridesDefault(t *testing.T) {
	p := NewPrioritizer()
	p.SetWeight(DriftKindImage, PriorityCritical)
	pr := p.Prioritize(priorityDriftedResult)
	if pr.Priority != PriorityCritical {
		t.Errorf("expected PriorityCritical after override, got %s", pr.Priority)
	}
}

func TestPrioritizer_PrioritizeAll_SortedDescending(t *testing.T) {
	p := NewPrioritizer()
	results := []DriftResult{priorityCleanResult, priorityDriftedResult}
	prs := p.PrioritizeAll(results)
	if len(prs) != 2 {
		t.Fatalf("expected 2 results, got %d", len(prs))
	}
	if prs[0].Priority < prs[1].Priority {
		t.Errorf("expected descending order: first=%s second=%s", prs[0].Priority, prs[1].Priority)
	}
}

func TestPrioritizer_PrioritizeAll_PreservesService(t *testing.T) {
	p := NewPrioritizer()
	prs := p.PrioritizeAll([]DriftResult{priorityDriftedResult})
	if prs[0].Result.Service != "api" {
		t.Errorf("expected service 'api', got %s", prs[0].Result.Service)
	}
}
