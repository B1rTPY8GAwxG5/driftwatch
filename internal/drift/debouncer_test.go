package drift

import (
	"sync"
	"testing"
	"time"
)

func TestNewDebouncer_DefaultQuiet(t *testing.T) {
	called := 0
	d := NewDebouncer(0, func(DriftResult) { called++ })
	if d.quiet != 5*time.Second {
		t.Errorf("expected default quiet 5s, got %v", d.quiet)
	}
}

func TestNewDebouncer_CustomQuiet(t *testing.T) {
	d := NewDebouncer(200*time.Millisecond, func(DriftResult) {})
	if d.quiet != 200*time.Millisecond {
		t.Errorf("unexpected quiet period: %v", d.quiet)
	}
}

func TestDebouncer_Pending_Zero(t *testing.T) {
	d := NewDebouncer(100*time.Millisecond, func(DriftResult) {})
	if d.Pending() != 0 {
		t.Errorf("expected 0 pending, got %d", d.Pending())
	}
}

func TestDebouncer_Submit_IncrementsPending(t *testing.T) {
	d := NewDebouncer(500*time.Millisecond, func(DriftResult) {})
	d.Submit(DriftResult{Service: "svc-a"})
	if d.Pending() != 1 {
		t.Errorf("expected 1 pending, got %d", d.Pending())
	}
	d.Flush()
}

func TestDebouncer_Submit_ForwardsAfterQuiet(t *testing.T) {
	var mu sync.Mutex
	received := []string{}

	d := NewDebouncer(50*time.Millisecond, func(r DriftResult) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, r.Service)
	})

	d.Submit(DriftResult{Service: "svc-b"})
	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 || received[0] != "svc-b" {
		t.Errorf("expected [svc-b], got %v", received)
	}
}

func TestDebouncer_Submit_DebouncesRapidCalls(t *testing.T) {
	var mu sync.Mutex
	count := 0

	d := NewDebouncer(80*time.Millisecond, func(DriftResult) {
		mu.Lock()
		defer mu.Unlock()
		count++
	})

	for i := 0; i < 5; i++ {
		d.Submit(DriftResult{Service: "svc-c"})
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("expected 1 forwarded call, got %d", count)
	}
}

func TestDebouncer_Flush_ForwardsImmediately(t *testing.T) {
	received := []string{}

	d := NewDebouncer(10*time.Second, func(r DriftResult) {
		received = append(received, r.Service)
	})

	d.Submit(DriftResult{Service: "svc-d"})
	d.Submit(DriftResult{Service: "svc-e"})

	if d.Pending() != 2 {
		t.Fatalf("expected 2 pending before flush, got %d", d.Pending())
	}

	d.Flush()

	if d.Pending() != 0 {
		t.Errorf("expected 0 pending after flush, got %d", d.Pending())
	}
	if len(received) != 2 {
		t.Errorf("expected 2 forwarded results, got %d", len(received))
	}
}

func TestDebouncer_Flush_EmptyIsNoop(t *testing.T) {
	d := NewDebouncer(100*time.Millisecond, func(DriftResult) {
		t.Error("forward should not be called on empty flush")
	})
	d.Flush()
}
