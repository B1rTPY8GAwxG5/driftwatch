package drift

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// MaturityFormatter renders a collection of MaturityRecords as human-readable text.
type MaturityFormatter struct {
	model *MaturityModel
}

// NewMaturityFormatter creates a MaturityFormatter backed by the given model.
func NewMaturityFormatter(model *MaturityModel) *MaturityFormatter {
	if model == nil {
		model = NewMaturityModel()
	}
	return &MaturityFormatter{model: model}
}

// Format evaluates all known services and writes a summary to w.
func (f *MaturityFormatter) Format(w io.Writer) error {
	svcs := f.model.Services()
	if len(svcs) == 0 {
		_, err := fmt.Fprintln(w, "maturity: no services recorded")
		return err
	}

	records := make([]MaturityRecord, 0, len(svcs))
	for _, svc := range svcs {
		rec, err := f.model.Evaluate(svc)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Service < records[j].Service
	})

	var sb strings.Builder
	sb.WriteString("=== Maturity Report ===\n")
	for _, r := range records {
		sb.WriteString(fmt.Sprintf("  %-30s level=%-12s drift_rate=%.2f%% observations=%d\n",
			r.Service,
			r.Level.String(),
			r.DriftRate*100,
			r.Observed,
		))
	}
	_, err := fmt.Fprint(w, sb.String())
	return err
}
