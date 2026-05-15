package drift

import (
	"strings"
	"testing"
)

func baseTransformResult() DriftResult {
	return DriftResult{
		Spec: ServiceSpec{Name: "MyService"},
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Got: "v1", Want: "v2"},
			{Kind: KindEnv, Field: "SECRET_KEY", Got: "abc123", Want: "xyz789"},
		},
	}
}

func TestNewTransformer_NotNil(t *testing.T) {
	tr := NewTransformer()
	if tr == nil {
		t.Fatal("expected non-nil Transformer")
	}
}

func TestTransformer_Transform_NoFns_ReturnsUnchanged(t *testing.T) {
	tr := NewTransformer()
	r := baseTransformResult()
	out := tr.Transform(r)
	if out.Spec.Name != "MyService" {
		t.Errorf("expected MyService, got %s", out.Spec.Name)
	}
}

func TestTransformer_Add_NilIgnored(t *testing.T) {
	tr := NewTransformer()
	tr.Add(nil)
	if len(tr.fns) != 0 {
		t.Errorf("expected 0 fns, got %d", len(tr.fns))
	}
}

func TestTransformer_Transform_AppliesInOrder(t *testing.T) {
	var order []int
	fn1 := TransformFn(func(r DriftResult) DriftResult { order = append(order, 1); return r })
	fn2 := TransformFn(func(r DriftResult) DriftResult { order = append(order, 2); return r })
	tr := NewTransformer(fn1, fn2)
	tr.Transform(baseTransformResult())
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestTransformer_TransformAll_Length(t *testing.T) {
	tr := NewTransformer()
	results := []DriftResult{baseTransformResult(), baseTransformResult()}
	out := tr.TransformAll(results)
	if len(out) != 2 {
		t.Errorf("expected 2 results, got %d", len(out))
	}
}

func TestTransformer_TransformAll_Empty(t *testing.T) {
	tr := NewTransformer(NormaliseServiceName())
	out := tr.TransformAll([]DriftResult{})
	if len(out) != 0 {
		t.Errorf("expected 0 results, got %d", len(out))
	}
}

func TestNormaliseServiceName(t *testing.T) {
	tr := NewTransformer(NormaliseServiceName())
	r := baseTransformResult()
	out := tr.Transform(r)
	if out.Spec.Name != strings.ToLower("MyService") {
		t.Errorf("expected lower-cased name, got %s", out.Spec.Name)
	}
}

func TestRedactEnvValues_RedactsEnvEntries(t *testing.T) {
	tr := NewTransformer(RedactEnvValues())
	out := tr.Transform(baseTransformResult())
	for _, e := range out.Entries {
		if e.Kind == KindEnv {
			if e.Got != "[redacted]" || e.Want != "[redacted]" {
				t.Errorf("env entry not redacted: got=%s want=%s", e.Got, e.Want)
			}
		}
	}
}

func TestRedactEnvValues_PreservesNonEnvEntries(t *testing.T) {
	tr := NewTransformer(RedactEnvValues())
	out := tr.Transform(baseTransformResult())
	for _, e := range out.Entries {
		if e.Kind == KindImage {
			if e.Got == "[redacted]" || e.Want == "[redacted]" {
				t.Error("non-env entry was incorrectly redacted")
			}
		}
	}
}
