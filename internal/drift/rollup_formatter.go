package drift

import (
	"fmt"
	"io"
	"strings"
)

// RollupFormatter writes a Rollup as text.
type RollupFormatter struct{}

// NewRollupFormatter returns a new RollupFormatter.
func NewRollupFormatter() *RollupFormatter {
	return &RollupFormatter{}
}

// Format writes the rollup summary to w.
func (f *RollupFormatter) Format(w io.Writer, r Rollup) error {
	_, err := fmt.Fprintf(w, "=== Drift Rollup ===\n%s\n", r.Summary())
	if err != nil {
		return err
	}
	for _, e := range r.Entries {
		status := "clean"
		kinds := ""
		if e.Drifted {
			status = "drifted"
			kindStrs := make([]string, len(e.Kinds))
			for i, k := range e.Kinds {
				kindStrs[i] = string(k)
			}
			kinds = " [" + strings.Join(kindStrs, ",") + "]"
		}
		_, err = fmt.Fprintf(w, "  %-30s score=%-4d severity=%-8s status=%s%s\n",
			e.Service, e.Score, e.Severity, status, kinds)
		if err != nil {
			return err
		}
	}
	return nil
}
