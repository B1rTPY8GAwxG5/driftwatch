package drift

import (
	"fmt"
	"sync"
	"time"
)

// SignalKind identifies the type of signal emitted.
type SignalKind string

const (
	SignalDriftDetected  SignalKind = "drift_detected"
	SignalDriftResolved  SignalKind = "drift_resolved"
	SignalBudgetExceeded SignalKind = "budget_exceeded"
	SignalStaleness      SignalKind = "staleness"
)

// Signal represents a discrete event emitted by the drift system.
type Signal struct {
	Kind      SignalKind        `json:"kind"`
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

func (s Signal) String() string {
	return fmt.Sprintf("[%s] %s — %s", s.Kind, s.Service, s.Message)
}

// SignalHandler is a function that processes a Signal.
type SignalHandler func(Signal)

// SignalBus fans out signals to registered handlers.
type SignalBus struct {
	mu       sync.RWMutex
	handlers []SignalHandler
}

// NewSignalBus returns an initialised SignalBus.
func NewSignalBus() *SignalBus {
	return &SignalBus{}
}

// Subscribe registers a handler. Nil handlers are ignored.
func (b *SignalBus) Subscribe(h SignalHandler) {
	if h == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, h)
}

// Publish sends a signal to all registered handlers.
func (b *SignalBus) Publish(s Signal) {
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}
	b.mu.RLock()
	handlers := make([]SignalHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()
	for _, h := range handlers {
		h(s)
	}
}

// Len returns the number of registered handlers.
func (b *SignalBus) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers)
}
