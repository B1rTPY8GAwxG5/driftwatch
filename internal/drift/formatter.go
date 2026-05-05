package drift

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// OutputFormat specifies the format for drift report output.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// Formatter writes a DriftReport to an output stream in a given format.
type Formatter interface {
	Format(w io.Writer, report *DriftReport) error
}

// TextFormatter renders a drift report as a human-readable table.
type TextFormatter struct{}

// NewTextFormatter creates a new TextFormatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format writes the drift report as plain text to w.
func (f *TextFormatter) Format(w io.Writer, report *DriftReport) error {
	if !report.HasDrift() {
		_, err := fmt.Fprintln(w, "✓ No drift detected.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tEXPECTED\tACTUAL")
	fmt.Fprintln(tw, "-----\t--------\t------")

	for _, entry := range report.Entries {
		fmt.Fprintf(tw, "%s\t%v\t%v\n", entry.Field, entry.Expected, entry.Actual)
	}

	return tw.Flush()
}

// NewFormatter returns a Formatter for the given OutputFormat.
func NewFormatter(format OutputFormat) (Formatter, error) {
	switch format {
	case FormatText, "":
		return NewTextFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q", format)
	}
}
