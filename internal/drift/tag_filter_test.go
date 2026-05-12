package drift

import (
	"testing"
)

func makeLabeledResult(service string, drifted bool, extra map[string]string) LabeledResult {
	labels := map[string]string{
		"service": service,
		"drifted": boolToStr(drifted),
	}
	for k, v := range extra {
		labels[k] = v
	}
	return LabeledResult{
		Result: DriftResult{Service: service},
		Labels: labels,
	}
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestNewTagFilter_NotNil(t *testing.T) {
	tf := NewTagFilter(nil, nil)
	if tf == nil {
		t.Fatal("expected non-nil TagFilter")
	}
}

func TestTagFilter_Match_NoConstraints_MatchesAll(t *testing.T) {
	tf := NewTagFilter(nil, nil)
	if !tf.Match(map[string]string{"service": "svc-a"}) {
		t.Error("expected match with no constraints")
	}
}

func TestTagFilter_Match_RequiredPresent(t *testing.T) {
	tf := NewTagFilter(map[string]string{"drifted": "true"}, nil)
	if !tf.Match(map[string]string{"drifted": "true", "service": "svc-a"}) {
		t.Error("expected match when required label present")
	}
}

func TestTagFilter_Match_RequiredMissing(t *testing.T) {
	tf := NewTagFilter(map[string]string{"drifted": "true"}, nil)
	if tf.Match(map[string]string{"service": "svc-a"}) {
		t.Error("expected no match when required label absent")
	}
}

func TestTagFilter_Match_ExcludedPresent(t *testing.T) {
	tf := NewTagFilter(nil, map[string]string{"env": "prod"})
	if tf.Match(map[string]string{"env": "prod", "service": "svc-a"}) {
		t.Error("expected no match when excluded label present")
	}
}

func TestTagFilter_Match_ExcludedAbsent(t *testing.T) {
	tf := NewTagFilter(nil, map[string]string{"env": "prod"})
	if !tf.Match(map[string]string{"env": "staging", "service": "svc-a"}) {
		t.Error("expected match when excluded label absent")
	}
}

func TestTagFilter_Apply_FiltersResults(t *testing.T) {
	tf := NewTagFilter(map[string]string{"drifted": "true"}, nil)
	input := []LabeledResult{
		makeLabeledResult("svc-a", true, nil),
		makeLabeledResult("svc-b", false, nil),
		makeLabeledResult("svc-c", true, nil),
	}
	got := tf.Apply(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	for _, r := range got {
		if r.Labels["drifted"] != "true" {
			t.Errorf("unexpected result without drifted=true: %s", r.Result.Service)
		}
	}
}

func TestTagFilter_Apply_EmptyInput(t *testing.T) {
	tf := NewTagFilter(map[string]string{"drifted": "true"}, nil)
	got := tf.Apply(nil)
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}
