package drift

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ProjectionField names a field that can be included in a projection.
type ProjectionField string

const (
	ProjectionFieldService  ProjectionField = "service"
	ProjectionFieldKind     ProjectionField = "kind"
	ProjectionFieldField    ProjectionField = "field"
	ProjectionFieldExpected ProjectionField = "expected"
	ProjectionFieldActual   ProjectionField = "actual"
	ProjectionFieldDrifted  ProjectionField = "drifted"
	ProjectionFieldTime     ProjectionField = "time"
)

// ProjectedRow is a single row produced by a Projection.
type ProjectedRow map[string]string

// Projection selects and shapes fields from DriftResult values.
type Projection struct {
	fields []ProjectionField
	time   time.Time
}

// NewProjection returns a Projection that includes the given fields.
// If no fields are provided all known fields are included.
func NewProjection(fields ...ProjectionField) *Projection {
	if len(fields) == 0 {
		fields = []ProjectionField{
			ProjectionFieldService,
			ProjectionFieldKind,
			ProjectionFieldField,
			ProjectionFieldExpected,
			ProjectionFieldActual,
			ProjectionFieldDrifted,
			ProjectionFieldTime,
		}
	}
	return &Projection{fields: fields, time: time.Now()}
}

// Project converts a DriftResult into a slice of ProjectedRows, one per
// DriftEntry. If the result has no entries a single summary row is returned.
func (p *Projection) Project(r DriftResult) []ProjectedRow {
	ts := p.time.Format(time.RFC3339)
	if len(r.Entries) == 0 {
		return []ProjectedRow{p.buildRow(r.Service, "", "", "", "", false, ts)}
	}
	rows := make([]ProjectedRow, 0, len(r.Entries))
	for _, e := range r.Entries {
		rows = append(rows, p.buildRow(
			r.Service,
			string(e.Kind),
			e.Field,
			fmt.Sprintf("%v", e.Expected),
			fmt.Sprintf("%v", e.Actual),
			true,
			ts,
		))
	}
	return rows
}

// Headers returns the ordered column headers for this projection.
func (p *Projection) Headers() []string {
	h := make([]string, len(p.fields))
	for i, f := range p.fields {
		h[i] = string(f)
	}
	return h
}

func (p *Projection) buildRow(service, kind, field, expected, actual string, drifted bool, ts string) ProjectedRow {
	all := map[ProjectionField]string{
		ProjectionFieldService:  service,
		ProjectionFieldKind:     kind,
		ProjectionFieldField:    field,
		ProjectionFieldExpected: expected,
		ProjectionFieldActual:   actual,
		ProjectionFieldDrifted:  boolStr(drifted),
		ProjectionFieldTime:     ts,
	}
	row := make(ProjectedRow, len(p.fields))
	for _, f := range p.fields {
		row[string(f)] = all[f]
	}
	return row
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// ProjectAll projects multiple results and returns all rows.
func (p *Projection) ProjectAll(results []DriftResult) []ProjectedRow {
	var rows []ProjectedRow
	for _, r := range results {
		rows = append(rows, p.Project(r)...)
	}
	return rows
}

// FormatTable renders projected rows as a plain-text table.
func FormatTable(headers []string, rows []ProjectedRow) string {
	widths := make(map[string]int)
	for _, h := range headers {
		widths[h] = len(h)
	}
	for _, row := range rows {
		for _, h := range headers {
			if l := len(row[h]); l > widths[h] {
				widths[h] = l
			}
		}
	}
	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s  ", widths[h], h))
	}
	sb.WriteString("\n")
	sorted := make([]ProjectedRow, len(rows))
	copy(sorted, rows)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i][string(ProjectionFieldService)] < sorted[j][string(ProjectionFieldService)]
	})
	for _, row := range sorted {
		for _, h := range headers {
			sb.WriteString(fmt.Sprintf("%-*s  ", widths[h], row[h]))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
