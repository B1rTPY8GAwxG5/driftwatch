package drift

import (
	"fmt"
	"io"
	"strings"
)

// GrouperFormatter renders grouped drift results as human-readable text.
type GrouperFormatter struct {
	grouper *Grouper
}

// NewGrouperFormatter returns a GrouperFormatter backed by the given Grouper.
// If grouper is nil a default service-mode grouper is used.
func NewGrouperFormatter(g *Grouper) *GrouperFormatter {
	if g == nil {
		g = NewGrouper(GroupByService)
	}
	return &GrouperFormatter{grouper: g}
}

// Format writes a textual representation of the grouped results to w.
func (f *GrouperFormatter) Format(w io.Writer, results []DriftResult) error {
	groups := f.grouper.Group(results)
	if len(groups) == 0 {
		_, err := fmt.Fprintln(w, "no results to display")
		return err
	}
	for _, g := range groups {
		driftCount := 0
		for _, r := range g.Results {
			if len(r.Entries) > 0 {
				driftCount++
			}
		}
		header := fmt.Sprintf("[%s] %d result(s), %d drifted",
			strings.ToUpper(g.Key), len(g.Results), driftCount)
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		for _, r := range g.Results {
			if len(r.Entries) == 0 {
				if _, err := fmt.Fprintf(w, "  %s: clean\n", r.Service); err != nil {
					return err
				}
				continue
			}
			for _, e := range r.Entries {
				line := fmt.Sprintf("  %s: %s %s -> %s\n",
					r.Service, e.Kind, e.Declared, e.Actual)
				if _, err := fmt.Fprint(w, line); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
