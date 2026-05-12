package drift

import (
	"fmt"
	"io"
	"strings"
)

// PriorityFormatter renders a prioritized list of drift results as human-readable text.
type PriorityFormatter struct {
	prioritizer *Prioritizer
}

// NewPriorityFormatter returns a PriorityFormatter backed by the given Prioritizer.
func NewPriorityFormatter(p *Prioritizer) *PriorityFormatter {
	if p == nil {
		p = NewPrioritizer()
	}
	return &PriorityFormatter{prioritizer: p}
}

// Format writes a prioritized summary of the given results to w.
func (f *PriorityFormatter) Format(w io.Writer, results []DriftResult) error {
	prs := f.prioritizer.PrioritizeAll(results)
	if len(prs) == 0 {
		_, err := fmt.Fprintln(w, "no results to display")
		return err
	}
	for _, pr := range prs {
		if err := f.writeEntry(w, pr); err != nil {
			return err
		}
	}
	return nil
}

func (f *PriorityFormatter) writeEntry(w io.Writer, pr PrioritizedResult) error {
	status := "clean"
	if pr.Result.HasDrift() {
		status = "drifted"
	}
	_, err := fmt.Fprintf(w, "[%s] %s (%s)\n",
		strings.ToUpper(pr.Priority.String()),
		pr.Result.Service,
		status,
	)
	if err != nil {
		return err
	}
	for _, entry := range pr.Result.Entries {
		_, err = fmt.Fprintf(w, "  - %s: declared=%q observed=%q\n",
			entry.Field, entry.Declared, entry.Observed)
		if err != nil {
			return err
		}
	}
	return nil
}
