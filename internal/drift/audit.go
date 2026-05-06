package drift

import (
	"fmt"
	"io"
	"time"
)

// AuditEvent represents a single recorded drift detection event.
type AuditEvent struct {
	Timestamp time.Time
	Service   string
	Drifted   bool
	Entries   []DriftEntry
	Message   string
}

// AuditLog holds an ordered list of audit events.
type AuditLog struct {
	events []AuditEvent
}

// NewAuditLog returns an initialised, empty AuditLog.
func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

// Record appends a new event derived from the given DriftResult.
func (a *AuditLog) Record(result DriftResult) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		Service:   result.Service,
		Drifted:   result.HasDrift(),
		Entries:   result.Entries,
	}
	if result.HasDrift() {
		event.Message = fmt.Sprintf("drift detected in %d field(s)", len(result.Entries))
	} else {
		event.Message = "no drift detected"
	}
	a.events = append(a.events, event)
}

// Events returns a copy of all recorded audit events.
func (a *AuditLog) Events() []AuditEvent {
	out := make([]AuditEvent, len(a.events))
	copy(out, a.events)
	return out
}

// Len returns the number of recorded events.
func (a *AuditLog) Len() int {
	return len(a.events)
}

// WriteTo writes a human-readable audit trail to w.
func (a *AuditLog) WriteTo(w io.Writer) error {
	for _, e := range a.events {
		line := fmt.Sprintf("[%s] service=%s drifted=%v msg=%q\n",
			e.Timestamp.Format(time.RFC3339), e.Service, e.Drifted, e.Message)
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}
	return nil
}
