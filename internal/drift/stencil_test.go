package drift

import (
	"testing"
)

var stencilResult = DriftResult{
	Service: "api-gateway",
	Spec:    ServiceSpec{Image: "nginx:1.25", Replicas: 3},
	Entries: []DriftEntry{
		{Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25", Kind: KindImage},
	},
}

var stencilCleanResult = DriftResult{
	Service: "worker",
	Spec:    ServiceSpec{Image: "alpine:3.18", Replicas: 2},
	Entries: []DriftEntry{},
}

func TestNewStencil_DefaultFields(t *testing.T) {
	s := NewStencil(nil, "")
	if s == nil {
		t.Fatal("expected non-nil stencil")
	}
	if len(s.fields) != 3 {
		t.Errorf("expected 3 default fields, got %d", len(s.fields))
	}
}

func TestNewStencil_DefaultSeparator(t *testing.T) {
	s := NewStencil(nil, "")
	if s.sep != "|" {
		t.Errorf("expected default sep '|', got %q", s.sep)
	}
}

func TestNewStencil_CustomSeparator(t *testing.T) {
	s := NewStencil(nil, ",")
	if s.sep != "," {
		t.Errorf("expected sep ',', got %q", s.sep)
	}
}

func TestStencil_Render_ContainsService(t *testing.T) {
	s := NewStencil(nil, "|")
	out := s.Render(stencilResult)
	if !contains(out, "api-gateway") {
		t.Errorf("expected service name in output, got %q", out)
	}
}

func TestStencil_Render_DriftedTrue(t *testing.T) {
	s := NewStencil(nil, "|")
	out := s.Render(stencilResult)
	if !contains(out, "true") {
		t.Errorf("expected 'true' in drifted output, got %q", out)
	}
}

func TestStencil_Render_DriftedFalse(t *testing.T) {
	s := NewStencil(nil, "|")
	out := s.Render(stencilCleanResult)
	if !contains(out, "false") {
		t.Errorf("expected 'false' in clean output, got %q", out)
	}
}

func TestStencil_Render_EntryCount(t *testing.T) {
	s := NewStencil(nil, "|")
	out := s.Render(stencilResult)
	if !contains(out, "1") {
		t.Errorf("expected entry count '1' in output, got %q", out)
	}
}

func TestStencil_Render_CustomFields(t *testing.T) {
	fields := []StencilField{
		{Name: "service"},
		{Name: "image"},
		{Name: "replicas"},
	}
	s := NewStencil(fields, "-")
	out := s.Render(stencilResult)
	if !contains(out, "nginx:1.25") {
		t.Errorf("expected image in output, got %q", out)
	}
	if !contains(out, "3") {
		t.Errorf("expected replicas in output, got %q", out)
	}
}

func TestStencil_RenderAll_Length(t *testing.T) {
	s := NewStencil(nil, "|")
	results := []DriftResult{stencilResult, stencilCleanResult}
	out := s.RenderAll(results)
	if len(out) != 2 {
		t.Errorf("expected 2 lines, got %d", len(out))
	}
}

func TestStencil_Render_UnknownField_Empty(t *testing.T) {
	fields := []StencilField{{Name: "unknown"}}
	s := NewStencil(fields, "|")
	out := s.Render(stencilResult)
	if out != "" {
		t.Errorf("expected empty string for unknown field, got %q", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
