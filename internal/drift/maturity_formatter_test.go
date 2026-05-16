package drift

import (
	"strings"
	"testing"
)

func TestMaturityFormatter_NoServices(t *testing.T) {
	m := NewMaturityModel()
	f := NewMaturityFormatter(m)
	var buf strings.Builder
	if err := f.Format(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no services") {
		t.Errorf("expected 'no services' in output, got: %q", buf.String())
	}
}

func TestMaturityFormatter_NilModel_UsesDefault(t *testing.T) {
	f := NewMaturityFormatter(nil)
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
}

func TestMaturityFormatter_SingleService_Rendered(t *testing.T) {
	m := NewMaturityModel()
	for i := 0; i < 5; i++ {
		m.Record("api-gateway", true)
	}
	f := NewMaturityFormatter(m)
	var buf strings.Builder
	if err := f.Format(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api-gateway") {
		t.Errorf("expected service name in output, got: %q", out)
	}
	if !strings.Contains(out, "unstable") {
		t.Errorf("expected level 'unstable' in output, got: %q", out)
	}
}

func TestMaturityFormatter_MultipleServices_SortedAlphabetically(t *testing.T) {
	m := NewMaturityModel()
	m.Record("zebra-svc", false)
	m.Record("alpha-svc", true)
	m.Record("alpha-svc", true)
	f := NewMaturityFormatter(m)
	var buf strings.Builder
	if err := f.Format(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	alpha := strings.Index(out, "alpha-svc")
	zebra := strings.Index(out, "zebra-svc")
	if alpha == -1 || zebra == -1 {
		t.Fatalf("expected both services in output")
	}
	if alpha > zebra {
		t.Errorf("expected alpha-svc before zebra-svc in output")
	}
}

func TestMaturityFormatter_ContainsHeader(t *testing.T) {
	m := NewMaturityModel()
	m.Record("svc", false)
	f := NewMaturityFormatter(m)
	var buf strings.Builder
	_ = f.Format(&buf)
	if !strings.Contains(buf.String(), "Maturity Report") {
		t.Errorf("expected header in output, got: %q", buf.String())
	}
}
