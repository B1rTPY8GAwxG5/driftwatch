package drift

import (
	"sync"
	"testing"
)

func TestNewObserver_NotNil(t *testing.T) {
	o := NewObserver()
	if o == nil {
		t.Fatal("expected non-nil observer")
	}
}

func TestObserver_Register_NilIgnored(t *testing.T) {
	o := NewObserver()
	o.Register(nil)
	if len(o.handlers) != 0 {
		t.Errorf("expected 0 handlers, got %d", len(o.handlers))
	}
}

func TestObserver_Observe_FansOutToHandlers(t *testing.T) {
	o := NewObserver()
	var called int
	o.Register(func(e ObserverEvent) { called++ })
	o.Register(func(e ObserverEvent) { called++ })

	o.Observe("svc-a", DriftResult{Service: "svc-a"})
	if called != 2 {
		t.Errorf("expected 2 handler calls, got %d", called)
	}
}

func TestObserver_Events_ReturnsCopy(t *testing.T) {
	o := NewObserver()
	o.Observe("svc-a", DriftResult{Service: "svc-a"})
	o.Observe("svc-b", DriftResult{Service: "svc-b"})

	events := o.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	events[0].Service = "mutated"
	if o.Events()[0].Service == "mutated" {
		t.Error("expected copy, but original was mutated")
	}
}

func TestObserver_Len(t *testing.T) {
	o := NewObserver()
	if o.Len() != 0 {
		t.Fatalf("expected 0, got %d", o.Len())
	}
	o.Observe("svc", DriftResult{})
	if o.Len() != 1 {
		t.Errorf("expected 1, got %d", o.Len())
	}
}

func TestObserver_Reset_ClearsEvents(t *testing.T) {
	o := NewObserver()
	o.Observe("svc", DriftResult{})
	o.Reset()
	if o.Len() != 0 {
		t.Errorf("expected 0 after reset, got %d", o.Len())
	}
}

func TestObserver_Observe_EventContainsService(t *testing.T) {
	o := NewObserver()
	o.Observe("my-service", DriftResult{Service: "my-service"})
	events := o.Events()
	if events[0].Service != "my-service" {
		t.Errorf("expected service 'my-service', got %q", events[0].Service)
	}
	if events[0].ObservedAt.IsZero() {
		t.Error("expected non-zero ObservedAt")
	}
}

func TestObserver_Observe_Concurrent(t *testing.T) {
	o := NewObserver()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.Observe("svc", DriftResult{})
		}()
	}
	wg.Wait()
	if o.Len() != 20 {
		t.Errorf("expected 20 events, got %d", o.Len())
	}
}
