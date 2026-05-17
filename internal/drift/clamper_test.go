package drift

import (
	"testing"
)

func TestDefaultClamperOptions_Values(t *testing.T) {
	opts := DefaultClamperOptions()
	if opts.MinReplicas != 1 {
		t.Errorf("expected MinReplicas 1, got %d", opts.MinReplicas)
	}
	if opts.MaxReplicas != 100 {
		t.Errorf("expected MaxReplicas 100, got %d", opts.MaxReplicas)
	}
	if opts.MinScore != 0.0 {
		t.Errorf("expected MinScore 0.0, got %f", opts.MinScore)
	}
	if opts.MaxScore != 100.0 {
		t.Errorf("expected MaxScore 100.0, got %f", opts.MaxScore)
	}
}

func TestNewClamper_NotNil(t *testing.T) {
	c := NewClamper(DefaultClamperOptions())
	if c == nil {
		t.Fatal("expected non-nil Clamper")
	}
}

func TestNewClamper_ZeroOptions_UsesDefaults(t *testing.T) {
	c := NewClamper(ClamperOptions{})
	if c.opts.MaxReplicas != 100 {
		t.Errorf("expected MaxReplicas 100 from defaults, got %d", c.opts.MaxReplicas)
	}
}

func TestClamper_Clamp_NoViolations(t *testing.T) {
	c := NewClamper(DefaultClamperOptions())
	result := DriftResult{Spec: ServiceSpec{Name: "svc", Replicas: 3}}
	out, violations := c.Clamp(result)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
	if out.Spec.Replicas != 3 {
		t.Errorf("expected replicas 3, got %d", out.Spec.Replicas)
	}
}

func TestClamper_Clamp_ReplicasAboveMax(t *testing.T) {
	c := NewClamper(ClamperOptions{MinReplicas: 1, MaxReplicas: 10})
	result := DriftResult{Spec: ServiceSpec{Name: "svc", Replicas: 50}}
	out, violations := c.Clamp(result)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Field != "spec.replicas" {
		t.Errorf("unexpected field %q", violations[0].Field)
	}
	if out.Spec.Replicas != 10 {
		t.Errorf("expected clamped replicas 10, got %d", out.Spec.Replicas)
	}
}

func TestClamper_Clamp_ReplicasBelowMin(t *testing.T) {
	c := NewClamper(ClamperOptions{MinReplicas: 2, MaxReplicas: 20})
	result := DriftResult{Spec: ServiceSpec{Name: "svc", Replicas: 1}}
	out, violations := c.Clamp(result)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if out.Spec.Replicas != 2 {
		t.Errorf("expected clamped replicas 2, got %d", out.Spec.Replicas)
	}
}

func TestClamperViolation_String(t *testing.T) {
	v := ClamperViolation{Field: "spec.replicas", Original: 200, Clamped: 100}
	s := v.String()
	if s == "" {
		t.Error("expected non-empty string from ClamperViolation.String()")
	}
}

func TestClamper_Clamp_ZeroReplicasNotClamped(t *testing.T) {
	// Zero replicas should be treated as unset and left unchanged.
	c := NewClamper(DefaultClamperOptions())
	result := DriftResult{Spec: ServiceSpec{Name: "svc", Replicas: 0}}
	out, violations := c.Clamp(result)
	if len(violations) != 0 {
		t.Errorf("expected no violations for zero replicas, got %d", len(violations))
	}
	if out.Spec.Replicas != 0 {
		t.Errorf("expected replicas unchanged at 0, got %d", out.Spec.Replicas)
	}
}
