package drift

import (
	"fmt"
	"strings"
)

// StencilField represents a named output field in a stencil template.
type StencilField struct {
	Name   string
	Format string // "text" or "json"
}

// Stencil renders a DriftResult into a structured string using a fixed set of fields.
type Stencil struct {
	fields []StencilField
	sep    string
}

// DefaultStencilFields returns the default set of fields rendered by a Stencil.
func DefaultStencilFields() []StencilField {
	return []StencilField{
		{Name: "service", Format: "text"},
		{Name: "drifted", Format: "text"},
		{Name: "entries", Format: "text"},
	}
}

// NewStencil creates a Stencil with the given fields and separator.
// If fields is empty, DefaultStencilFields is used.
// If sep is empty, "|" is used.
func NewStencil(fields []StencilField, sep string) *Stencil {
	if len(fields) == 0 {
		fields = DefaultStencilFields()
	}
	if sep == "" {
		sep = "|"
	}
	return &Stencil{fields: fields, sep: sep}
}

// Render produces a single-line string representation of the result
// using the configured fields.
func (s *Stencil) Render(r DriftResult) string {
	parts := make([]string, 0, len(s.fields))
	for _, f := range s.fields {
		parts = append(parts, s.renderField(f, r))
	}
	return strings.Join(parts, s.sep)
}

// RenderAll renders each result and returns one line per result.
func (s *Stencil) RenderAll(results []DriftResult) []string {
	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, s.Render(r))
	}
	return out
}

func (s *Stencil) renderField(f StencilField, r DriftResult) string {
	switch f.Name {
	case "service":
		return r.Service
	case "drifted":
		if r.HasDrift() {
			return "true"
		}
		return "false"
	case "entries":
		return fmt.Sprintf("%d", len(r.Entries))
	case "image":
		return r.Spec.Image
	case "replicas":
		return fmt.Sprintf("%d", r.Spec.Replicas)
	default:
		return ""
	}
}
