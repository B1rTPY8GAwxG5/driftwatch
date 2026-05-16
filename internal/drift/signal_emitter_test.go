package drift

import (
	"strings"
	"testing"
)

func TestNewSignalEmitter_NotNil(t *testing.T) {
	e := NewSignalEmitter(nil, nil)
	if e == nil {
		t.Fatal("expected non-nil SignalEmitter")
	}
}

func TestSignalEmitter_Emit_CleanResult_Resolved(t *testing.T) {
	bus := NewSignalBus()
	var got Signal
	bus.Subscribe(func(s Signal) { got = s })
	e := NewSignalEmitter(bus, nil)
	e.Emit(DriftResult{Service: "svc-clean"})
	if got.Kind != SignalDriftResolved {
		t.Fatalf("expected %s, got %s", SignalDriftResolved, got.Kind)
	}
}

func TestSignalEmitter_Emit_DriftedResult_Detected(t *testing.T) {
	bus := NewSignalBus()
	var got Signal
	bus.Subscribe(func(s Signal) { got = s })
	e := NewSignalEmitter(bus, nil)
	e.Emit(DriftResult{
		Service: "svc-drift",
		Entries: []DriftEntry{{Field: "image", Kind: KindImage}},
	})
	if got.Kind != SignalDriftDetected {
		t.Fatalf("expected %s, got %s", SignalDriftDetected, got.Kind)
	}
	if got.Service != "svc-drift" {
		t.Fatalf("unexpected service: %s", got.Service)
	}
}

func TestSignalEmitter_Emit_MessageContainsService(t *testing.T) {
	bus := NewSignalBus()
	var got Signal
	bus.Subscribe(func(s Signal) { got = s })
	e := NewSignalEmitter(bus, nil)
	e.Emit(DriftResult{Service: "my-svc"})
	if !strings.Contains(got.Message, "my-svc") {
		t.Errorf("expected service name in message, got: %s", got.Message)
	}
}

func TestSignalEmitter_Emit_StaticLabelsPresent(t *testing.T) {
	bus := NewSignalBus()
	var got Signal
	bus.Subscribe(func(s Signal) { got = s })
	e := NewSignalEmitter(bus, map[string]string{"env": "prod"})
	e.Emit(DriftResult{Service: "svc"})
	if got.Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got: %v", got.Labels)
	}
}

func TestSignalEmitter_EmitBudgetExceeded_Kind(t *testing.T) {
	bus := NewSignalBus()
	var got Signal
	bus.Subscribe(func(s Signal) { got = s })
	e := NewSignalEmitter(bus, nil)
	e.EmitBudgetExceeded("svc-x", 5, 3)
	if got.Kind != SignalBudgetExceeded {
		t.Fatalf("expected %s, got %s", SignalBudgetExceeded, got.Kind)
	}
	if !strings.Contains(got.Message, "5/3") {
		t.Errorf("expected usage in message, got: %s", got.Message)
	}
}

func TestSignalEmitter_Emit_NilBus_DoesNotPanic(t *testing.T) {
	e := NewSignalEmitter(nil, nil)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	e.Emit(DriftResult{Service: "safe"})
}
