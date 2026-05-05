package drift

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// AlertLevel represents the severity of a drift alert.
type AlertLevel string

const (
	AlertLevelWarn  AlertLevel = "WARN"
	AlertLevelError AlertLevel = "ERROR"
)

// Alert represents a single drift notification.
type Alert struct {
	Level     AlertLevel
	Service   string
	Message   string
	Timestamp time.Time
	Entries   []DriftEntry
}

// String returns a human-readable representation of the alert.
func (a Alert) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] %s — %s (%s)\n", a.Level, a.Timestamp.Format(time.RFC3339), a.Service, a.Message)
	for _, e := range a.Entries {
		fmt.Fprintf(&sb, "  • %s: expected=%q actual=%q\n", e.Kind, e.Expected, e.Actual)
	}
	return sb.String()
}

// AlertHandler is a function that receives and processes an Alert.
type AlertHandler func(Alert)

// NewAlert builds an Alert from a DriftResult.
func NewAlert(result DriftResult, level AlertLevel) Alert {
	return Alert{
		Level:     level,
		Service:   result.ServiceName,
		Message:   fmt.Sprintf("%d drift(s) detected", len(result.Entries)),
		Timestamp: time.Now().UTC(),
		Entries:   result.Entries,
	}
}

// WriteAlert formats an Alert and writes it to the provided writer.
func WriteAlert(w io.Writer, a Alert) error {
	_, err := fmt.Fprint(w, a.String())
	return err
}
