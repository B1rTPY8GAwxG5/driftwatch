package drift

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// TraceEvent records a single step in the lifecycle of a drift check.
type TraceEvent struct {
	Stage     string
	Service   string
	Message   string
	Timestamp time.Time
	Duration  time.Duration
	Err       error
}

func (e TraceEvent) String() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s/%s ERR=%v (%s)", e.Timestamp.Format(time.RFC3339), e.Stage, e.Service, e.Err, e.Duration)
	}
	return fmt.Sprintf("[%s] %s/%s %s (%s)", e.Timestamp.Format(time.RFC3339), e.Stage, e.Service, e.Message, e.Duration)
}

// Tracer collects trace events across drift pipeline stages.
type Tracer struct {
	mu     sync.Mutex
	events []TraceEvent
}

// NewTracer returns an empty Tracer.
func NewTracer() *Tracer {
	return &Tracer{}
}

// Record appends a trace event.
func (t *Tracer) Record(stage, service, message string, duration time.Duration, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, TraceEvent{
		Stage:     stage,
		Service:   service,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Duration:  duration,
		Err:       err,
	})
}

// Events returns a copy of all recorded trace events.
func (t *Tracer) Events() []TraceEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TraceEvent, len(t.events))
	copy(out, t.events)
	return out
}

// Len returns the number of recorded events.
func (t *Tracer) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events)
}

// Reset clears all recorded events.
func (t *Tracer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = t.events[:0]
}

// WriteTo writes all trace events to w in human-readable form.
func (t *Tracer) WriteTo(w io.Writer) {
	for _, e := range t.Events() {
		fmt.Fprintln(w, e.String())
	}
}
