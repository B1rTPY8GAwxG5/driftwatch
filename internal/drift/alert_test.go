package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func driftedResult() DriftResult {
	return DriftResult{
		ServiceName: "api-server",
		Entries: []DriftEntry{
			{Kind: KindImage, Expected: "nginx:1.25", Actual: "nginx:1.24"},
		},
	}
}

func TestNewAlert_Fields(t *testing.T) {
	result := driftedResult()
	a := NewAlert(result, AlertLevelWarn)

	if a.Service != "api-server" {
		t.Errorf("expected service api-server, got %s", a.Service)
	}
	if a.Level != AlertLevelWarn {
		t.Errorf("expected level WARN, got %s", a.Level)
	}
	if len(a.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(a.Entries))
	}
	if a.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAlert_String_ContainsService(t *testing.T) {
	a := NewAlert(driftedResult(), AlertLevelError)
	s := a.String()

	if !strings.Contains(s, "api-server") {
		t.Errorf("expected service name in alert string, got: %s", s)
	}
	if !strings.Contains(s, "ERROR") {
		t.Errorf("expected ERROR level in alert string, got: %s", s)
	}
	if !strings.Contains(s, "nginx:1.25") {
		t.Errorf("expected expected value in alert string, got: %s", s)
	}
}

func TestWriteAlert_WritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	a := Alert{
		Level:     AlertLevelWarn,
		Service:   "svc",
		Message:   "1 drift(s) detected",
		Timestamp: time.Now().UTC(),
		Entries:   []DriftEntry{{Kind: KindReplicas, Expected: "3", Actual: "2"}},
	}

	if err := WriteAlert(&buf, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestAlertLevel_Constants(t *testing.T) {
	if AlertLevelWarn != "WARN" {
		t.Errorf("unexpected WARN value: %s", AlertLevelWarn)
	}
	if AlertLevelError != "ERROR" {
		t.Errorf("unexpected ERROR value: %s", AlertLevelError)
	}
}
