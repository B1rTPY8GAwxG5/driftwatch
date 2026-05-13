package drift

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func makeEvent(service string, drifted bool, offset time.Duration) ReplayEvent {
	entries := []DriftEntry{}
	if drifted {
		entries = append(entries, DriftEntry{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"})
	}
	return ReplayEvent{
		Timestamp: t0.Add(offset),
		Result:    DriftResult{Service: service, Entries: entries},
	}
}

func TestNewReplayer_NotNil(t *testing.T) {
	r := NewReplayer(nil, func(ReplayEvent) error { return nil }, 0)
	if r == nil {
		t.Fatal("expected non-nil Replayer")
	}
}

func TestReplayer_Len(t *testing.T) {
	events := []ReplayEvent{makeEvent("svc-a", false, 0), makeEvent("svc-b", true, time.Second)}
	r := NewReplayer(events, func(ReplayEvent) error { return nil }, 0)
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
}

func TestReplayer_Run_AllHandlersCalled(t *testing.T) {
	events := []ReplayEvent{
		makeEvent("svc-a", false, 0),
		makeEvent("svc-b", true, time.Second),
		makeEvent("svc-c", false, 2*time.Second),
	}
	var seen []string
	r := NewReplayer(events, func(ev ReplayEvent) error {
		seen = append(seen, ev.Result.Service)
		return nil
	}, 0)
	if err := r.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(seen))
	}
}

func TestReplayer_Run_SortsChronologically(t *testing.T) {
	events := []ReplayEvent{
		makeEvent("third", false, 2*time.Second),
		makeEvent("first", false, 0),
		makeEvent("second", false, time.Second),
	}
	var order []string
	r := NewReplayer(events, func(ev ReplayEvent) error {
		order = append(order, ev.Result.Service)
		return nil
	}, 0)
	_ = r.Run()
	if order[0] != "first" || order[1] != "second" || order[2] != "third" {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestReplayer_Run_HandlerError_Stops(t *testing.T) {
	events := []ReplayEvent{makeEvent("svc-a", false, 0), makeEvent("svc-b", false, time.Second)}
	calls := 0
	r := NewReplayer(events, func(ReplayEvent) error {
		calls++
		return errors.New("boom")
	}, 0)
	err := r.Run()
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("expected handler called once, got %d", calls)
	}
}

func TestWriteReplaySummary_NoDrift(t *testing.T) {
	events := []ReplayEvent{makeEvent("svc-a", false, 0)}
	var buf bytes.Buffer
	if err := WriteReplaySummary(events, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "clean") {
		t.Errorf("expected 'clean' in output, got: %s", buf.String())
	}
}

func TestWriteReplaySummary_WithDrift(t *testing.T) {
	events := []ReplayEvent{makeEvent("svc-b", true, 0)}
	var buf bytes.Buffer
	_ = WriteReplaySummary(events, &buf)
	if !strings.Contains(buf.String(), "drifted") {
		t.Errorf("expected 'drifted' in output, got: %s", buf.String())
	}
}
