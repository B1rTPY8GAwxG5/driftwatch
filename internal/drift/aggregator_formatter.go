package drift

import (
	"fmt"
	"io"
	"strings"
)

// AggregatorFormatter renders an AggregatedResult as human-readable text.
type AggregatorFormatter struct{}

// NewAggregatorFormatter returns a new AggregatorFormatter.
func NewAggregatorFormatter() *AggregatorFormatter {
	return &AggregatorFormatter{}
}

// Format writes a text representation of the AggregatedResult to w.
func (f *AggregatorFormatter) Format(w io.Writer, ar AggregatedResult) error {
	_, err := fmt.Fprintf(w, "Aggregation Mode : %s\n", ar.Mode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Total Services   : %d\n", len(ar.Results))
	if err != nil {
		return err
	}
	status := "clean"
	if ar.HasDrift() {
		status = "DRIFTED"
	}
	_, err = fmt.Fprintf(w, "Status           : %s\n", status)
	if err != nil {
		return err
	}
	if len(ar.Services) > 0 {
		_, err = fmt.Fprintf(w, "Drifted Services : %s\n", strings.Join(ar.Services, ", "))
		if err != nil {
			return err
		}
	}
	return nil
}
