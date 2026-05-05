package drift

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Report aggregates drift results across multiple services.
type Report struct {
	Service string
	Results []DriftResult
}

// HasDrift returns true if any drift was detected.
func (r *Report) HasDrift() bool {
	return len(r.Results) > 0
}

// Summary returns a human-readable one-line summary.
func (r *Report) Summary() string {
	if !r.HasDrift() {
		return fmt.Sprintf("[OK] %s: no drift detected", r.Service)
	}
	return fmt.Sprintf("[DRIFT] %s: %d field(s) differ", r.Service, len(r.Results))
}

// WriteTo writes a formatted drift table to the given writer.
func (r *Report) WriteTo(w io.Writer) error {
	fmt.Fprintln(w, r.Summary())
	if !r.HasDrift() {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  FIELD\tSTATUS\tDECLARED\tLIVE")
	fmt.Fprintln(tw, "  "+strings.Repeat("-", 60))
	for _, res := range r.Results {
		fmt.Fprintf(tw, "  %s\t%s\t%v\t%v\n",
			res.Field,
			res.Status,
			res.Declared,
			res.Live,
		)
	}
	return tw.Flush()
}

// BuildReport runs the detector and returns a populated Report.
func BuildReport(declared, live ServiceConfig) *Report {
	d := NewDetector()
	results := d.Compare(declared, live)
	return &Report{
		Service: declared.Name,
		Results: results,
	}
}
