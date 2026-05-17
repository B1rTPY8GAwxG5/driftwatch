package drift

import (
	"bytes"
	"strings"
	"testing"
)

func makeResult(service string, kinds ...DriftKind) DriftResult {
	entries := make([]DriftEntry, 0, len(kinds))
	for _, k := range kinds {
		entries = append(entries, DriftEntry{Kind: k, Field: string(k), Expected: "a", Actual: "b"})
	}
	return DriftResult{Service: service, Entries: entries}
}

func TestBuildScoreReport_Empty(t *testing.T) {
	report := BuildScoreReport(nil)
	if len(report.Scores) != 0 {
		t.Errorf("expected empty scores, got %d", len(report.Scores))
	}
}

func TestBuildScoreReport_SortedDescending(t *testing.T) {
	results := []DriftResult{
		makeResult("svc-env", DriftKindEnv),
		makeResult("svc-image", DriftKindImage),
		makeResult("svc-replicas", DriftKindReplicas),
	}
	report := BuildScoreReport(results)
	if report.Scores[0].Service != "svc-image" {
		t.Errorf("expected svc-image first, got %s", report.Scores[0].Service)
	}
	if report.Scores[2].Service != "svc-env" {
		t.Errorf("expected svc-env last, got %s", report.Scores[2].Service)
	}
}

func TestScoreReport_HasCritical_True(t *testing.T) {
	report := BuildScoreReport([]DriftResult{makeResult("svc", DriftKindImage)})
	if !report.HasCritical() {
		t.Error("expected HasCritical true")
	}
}

func TestScoreReport_HasCritical_False(t *testing.T) {
	report := BuildScoreReport([]DriftResult{makeResult("svc", DriftKindEnv)})
	if report.HasCritical() {
		t.Error("expected HasCritical false")
	}
}

func TestScoreReport_WriteTo_Empty(t *testing.T) {
	var buf bytes.Buffer
	report := BuildScoreReport(nil)
	if err := report.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no results") {
		t.Errorf("expected 'no results' in output, got: %s", buf.String())
	}
}

func TestScoreReport_WriteTo_ContainsService(t *testing.T) {
	var buf bytes.Buffer
	report := BuildScoreReport([]DriftResult{makeResult("my-service", DriftKindImage)})
	if err := report.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "my-service") {
		t.Errorf("expected service name in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "critical") {
		t.Errorf("expected level critical in output, got: %s", buf.String())
	}
}

func TestBuildScoreReport_ScoreCount(t *testing.T) {
	results := []DriftResult{
		makeResult("svc-a", DriftKindImage),
		makeResult("svc-b", DriftKindEnv),
		makeResult("svc-c", DriftKindReplicas),
	}
	report := BuildScoreReport(results)
	if len(report.Scores) != len(results) {
		t.Errorf("expected %d scores, got %d", len(results), len(report.Scores))
	}
}
