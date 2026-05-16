package drift

import (
	"testing"
	"time"
)

func TestNewAnnotator_NotNil(t *testing.T) {
	a := NewAnnotator(nil)
	if a == nil {
		t.Fatal("expected non-nil Annotator")
	}
}

func TestAnnotator_Annotate_StaticKeys(t *testing.T) {
	a := NewAnnotator(map[string]string{"env": "production", "team": "platform"})
	r := DriftResult{Service: "svc-a"}
	ar := a.Annotate(r)
	if ar.Annotations["env"] != "production" {
		t.Errorf("expected env=production, got %q", ar.Annotations["env"])
	}
	if ar.Annotations["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", ar.Annotations["team"])
	}
}

func TestAnnotator_Annotate_DynamicProvider(t *testing.T) {
	a := NewAnnotator(nil)
	a.AddProvider(func(r DriftResult) map[string]string {
		return map[string]string{"service": r.Service}
	})
	r := DriftResult{Service: "svc-b"}
	ar := a.Annotate(r)
	if ar.Annotations["service"] != "svc-b" {
		t.Errorf("expected service=svc-b, got %q", ar.Annotations["service"])
	}
}

func TestAnnotator_Annotate_NilProviderIgnored(t *testing.T) {
	a := NewAnnotator(nil)
	a.AddProvider(nil)
	r := DriftResult{Service: "svc-c"}
	ar := a.Annotate(r)
	if ar.Annotations == nil {
		t.Fatal("expected non-nil annotations map")
	}
}

func TestAnnotator_Annotate_SetsAnnotatedAt(t *testing.T) {
	a := NewAnnotator(nil)
	before := time.Now().UTC()
	ar := a.Annotate(DriftResult{Service: "svc-d"})
	after := time.Now().UTC()
	if ar.AnnotatedAt.Before(before) || ar.AnnotatedAt.After(after) {
		t.Errorf("AnnotatedAt %v out of expected range [%v, %v]", ar.AnnotatedAt, before, after)
	}
}

func TestAnnotator_Annotate_ProviderOverridesStatic(t *testing.T) {
	a := NewAnnotator(map[string]string{"region": "us-east-1"})
	a.AddProvider(func(r DriftResult) map[string]string {
		return map[string]string{"region": "eu-west-1"}
	})
	ar := a.Annotate(DriftResult{Service: "svc-e"})
	if ar.Annotations["region"] != "eu-west-1" {
		t.Errorf("expected provider to override static, got %q", ar.Annotations["region"])
	}
}

func TestAnnotator_AnnotateAll_Length(t *testing.T) {
	a := NewAnnotator(map[string]string{"x": "1"})
	results := []DriftResult{
		{Service: "svc-1"},
		{Service: "svc-2"},
		{Service: "svc-3"},
	}
	out := a.AnnotateAll(results)
	if len(out) != len(results) {
		t.Errorf("expected %d annotated results, got %d", len(results), len(out))
	}
}

func TestAnnotator_AnnotateAll_PreservesService(t *testing.T) {
	a := NewAnnotator(nil)
	results := []DriftResult{{Service: "svc-x"}}
	out := a.AnnotateAll(results)
	if out[0].Service != "svc-x" {
		t.Errorf("expected service svc-x, got %q", out[0].Service)
	}
}
