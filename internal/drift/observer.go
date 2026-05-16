package drift

import (
	"sync"
	"time"
)

// ObserverEvent represents a single observation of a drift result.
type ObserverEvent struct {
	Service   string
	Result    DriftResult
	ObservedAt time.Time
}

// ObserverHandler is a function invoked for each observed event.
type ObserverHandler func(ObserverEvent)

// Observer watches drift results and fans out to registered handlers.
type Observer struct {
	mu       sync.RWMutex
	handlers []ObserverHandler
	events   []ObserverEvent
}

// NewObserver creates a new Observer with no handlers.
func NewObserver() *Observer {
	return &Observer{}
}

// Register adds a handler to the observer.
func (o *Observer) Register(h ObserverHandler) {
	if h == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers = append(o.handlers, h)
}

// Observe records a drift result and fans out to all registered handlers.
func (o *Observer) Observe(service string, result DriftResult) {
	event := ObserverEvent{
		Service:    service,
		Result:     result,
		ObservedAt: time.Now().UTC(),
	}
	o.mu.Lock()
	o.events = append(o.events, event)
	handlers := make([]ObserverHandler, len(o.handlers))
	copy(handlers, o.handlers)
	o.mu.Unlock()

	for _, h := range handlers {
		h(event)
	}
}

// Events returns a copy of all observed events.
func (o *Observer) Events() []ObserverEvent {
	o.mu.RLock()
	defer o.mu.RUnlock()
	out := make([]ObserverEvent, len(o.events))
	copy(out, o.events)
	return out
}

// Len returns the number of observed events.
func (o *Observer) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.events)
}

// Reset clears all recorded events.
func (o *Observer) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.events = nil
}
