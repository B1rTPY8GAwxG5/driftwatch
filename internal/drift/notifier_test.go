package drift_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func driftedResultForNotifier() drift.DriftResult {
	return drift.DriftResult{
		Service: "api",
		Entries: []drift.DriftEntry{
			{Kind: drift.KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		},
	}
}

func cleanResult() drift.DriftResult {
	return drift.DriftResult{Service: "api", Entries: nil}
}

func TestNewNotifier_NoWriters(t *testing.T) {
	_, err := drift.NewNotifier(drift.NotifyAll)
	if err == nil {
		t.Fatal("expected error for no writers")
	}
}

func TestNewNotifier_Valid(t *testing.T) {
	n, err := drift.NewNotifier(drift.NotifyAll, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}

func TestNotifier_Notify_DriftOnly_SkipsClean(t *testing.T) {
	var buf bytes.Buffer
	n, _ := drift.NewNotifier(drift.NotifyDriftOnly, &buf)
	_ = n.Notify(cleanResult())
	if buf.Len() != 0 {
		t.Errorf("expected no output for clean result, got: %s", buf.String())
	}
}

func TestNotifier_Notify_DriftOnly_WritesDrift(t *testing.T) {
	var buf bytes.Buffer
	n, _ := drift.NewNotifier(drift.NotifyDriftOnly, &buf)
	_ = n.Notify(driftedResultForNotifier())
	if !strings.Contains(buf.String(), "api") {
		t.Errorf("expected service name in output, got: %s", buf.String())
	}
}

func TestNotifier_Notify_All_WritesClean(t *testing.T) {
	var buf bytes.Buffer
	n, _ := drift.NewNotifier(drift.NotifyAll, &buf)
	_ = n.Notify(cleanResult())
	if buf.Len() == 0 {
		t.Error("expected output for NotifyAll even with clean result")
	}
}

func TestNotifier_Notify_MultipleWriters(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	n, _ := drift.NewNotifier(drift.NotifyAll, &buf1, &buf2)
	_ = n.Notify(driftedResultForNotifier())
	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("expected both writers to receive output")
	}
}

func TestNotifier_WithFilters_FiltersEntries(t *testing.T) {
	var buf bytes.Buffer
	n, _ := drift.NewNotifier(drift.NotifyDriftOnly, &buf)
	f, _ := drift.NewFilter(drift.KindReplicas)
	n.WithFilters(f)
	// result has KindImage drift, filter only passes KindReplicas — no match
	_ = n.Notify(driftedResultForNotifier())
	if buf.Len() != 0 {
		t.Errorf("expected filtered output to be empty, got: %s", buf.String())
	}
}
