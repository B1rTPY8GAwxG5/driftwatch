package drift

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewSignalBus_NotNil(t *testing.T) {
	b := NewSignalBus()
	if b == nil {
		t.Fatal("expected non-nil SignalBus")
	}
}

func TestSignalBus_Len_Zero(t *testing.T) {
	b := NewSignalBus()
	if b.Len() != 0 {
		t.Fatalf("expected 0, got %d", b.Len())
	}
}

func TestSignalBus_Subscribe_NilIgnored(t *testing.T) {
	b := NewSignalBus()
	b.Subscribe(nil)
	if b.Len() != 0 {
		t.Fatal("nil handler should not be registered")
	}
}

func TestSignalBus_Subscribe_IncrementsLen(t *testing.T) {
	b := NewSignalBus()
	b.Subscribe(func(Signal) {})
	b.Subscribe(func(Signal) {})
	if b.Len() != 2 {
		t.Fatalf("expected 2, got %d", b.Len())
	}
}

func TestSignalBus_Publish_AllHandlersCalled(t *testing.T) {
	b := NewSignalBus()
	var mu sync.Mutex
	var received []Signal
	for i := 0; i < 3; i++ {
		b.Subscribe(func(s Signal) {
			mu.Lock()
			received = append(received, s)
			mu.Unlock()
		})
	}
	b.Publish(Signal{Kind: SignalDriftDetected, Service: "svc-a", Message: "image changed"})
	if len(received) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(received))
	}
}

func TestSignalBus_Publish_SetsTimestamp(t *testing.T) {
	b := NewSignalBus()
	var got Signal
	b.Subscribe(func(s Signal) { got = s })
	before := time.Now()
	b.Publish(Signal{Kind: SignalDriftResolved, Service: "svc"})
	if got.Timestamp.Before(before) {
		t.Fatal("timestamp should be set on publish")
	}
}

func TestSignalBus_Publish_PreservesExistingTimestamp(t *testing.T) {
	b := NewSignalBus()
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var got Signal
	b.Subscribe(func(s Signal) { got = s })
	b.Publish(Signal{Kind: SignalStaleness, Service: "svc", Timestamp: fixed})
	if !got.Timestamp.Equal(fixed) {
		t.Fatalf("expected fixed timestamp, got %v", got.Timestamp)
	}
}

func TestSignalKind_Constants(t *testing.T) {
	kinds := []SignalKind{SignalDriftDetected, SignalDriftResolved, SignalBudgetExceeded, SignalStaleness}
	for _, k := range kinds {
		if string(k) == "" {
			t.Fatalf("signal kind must not be empty")
		}
	}
}

func TestSignal_String_ContainsKindAndService(t *testing.T) {
	s := Signal{Kind: SignalDriftDetected, Service: "my-service", Message: "replica mismatch"}
	out := s.String()
	if !strings.Contains(out, "drift_detected") {
		t.Errorf("expected kind in string, got: %s", out)
	}
	if !strings.Contains(out, "my-service") {
		t.Errorf("expected service in string, got: %s", out)
	}
}
