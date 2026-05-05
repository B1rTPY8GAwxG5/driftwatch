package drift

import (
	"fmt"
	"io"
	"strings"
)

// NotifyLevel controls which drift results trigger notifications.
type NotifyLevel int

const (
	NotifyAll NotifyLevel = iota
	NotifyDriftOnly
)

// Notifier dispatches alerts for drift results to one or more writers.
type Notifier struct {
	writers []io.Writer
	level   NotifyLevel
	filters []*Filter
}

// NewNotifier creates a Notifier that writes to the given writers.
func NewNotifier(level NotifyLevel, writers ...io.Writer) (*Notifier, error) {
	if len(writers) == 0 {
		return nil, fmt.Errorf("notifier requires at least one writer")
	}
	return &Notifier{
		writers: writers,
		level:   level,
	}, nil
}

// WithFilters attaches filters that gate which drift entries are notified.
func (n *Notifier) WithFilters(filters ...*Filter) {
	n.filters = append(n.filters, filters...)
}

// Notify evaluates the DriftResult and writes an Alert to all writers
// if the result meets the configured level and filter criteria.
func (n *Notifier) Notify(result DriftResult) error {
	if n.level == NotifyDriftOnly && !result.HasDrift() {
		return nil
	}

	filtered := result
	if len(n.filters) > 0 {
		filtered = ApplyAll(result, n.filters)
	}

	if n.level == NotifyDriftOnly && !filtered.HasDrift() {
		return nil
	}

	alert := NewAlert(filtered)
	var errs []string
	for _, w := range n.writers {
		if err := WriteAlert(w, alert); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notifier write errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
