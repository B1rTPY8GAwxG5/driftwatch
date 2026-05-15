package drift

import (
	"testing"
)

func TestDefaultNormalizerOptions_Values(t *testing.T) {
	opts := DefaultNormalizerOptions()
	if !opts.LowercaseImage {
		t.Error("expected LowercaseImage to be true")
	}
	if !opts.TrimEnvValues {
		t.Error("expected TrimEnvValues to be true")
	}
	if opts.CollapseReplicas {
		t.Error("expected CollapseReplicas to be false")
	}
}

func TestNewNormalizer_NotNil(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerOptions())
	if n == nil {
		t.Fatal("expected non-nil Normalizer")
	}
}

func TestNormalizer_Normalize_LowercasesImage(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerOptions())
	spec := ServiceSpec{Name: "svc", Image: "MyApp:Latest", Replicas: 1}
	out := n.Normalize(spec)
	if out.Image != "myapp:latest" {
		t.Errorf("expected 'myapp:latest', got %q", out.Image)
	}
}

func TestNormalizer_Normalize_TrimsImageWhitespace(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerOptions())
	spec := ServiceSpec{Name: "svc", Image: "  nginx:1.25  ", Replicas: 1}
	out := n.Normalize(spec)
	if out.Image != "nginx:1.25" {
		t.Errorf("expected 'nginx:1.25', got %q", out.Image)
	}
}

func TestNormalizer_Normalize_TrimsEnvValues(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerOptions())
	spec := ServiceSpec{
		Name:  "svc",
		Image: "img",
		Env:   map[string]string{"KEY": "  value  "},
	}
	out := n.Normalize(spec)
	if got := out.Env["KEY"]; got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
}

func TestNormalizer_Normalize_CollapseReplicas(t *testing.T) {
	opts := DefaultNormalizerOptions()
	opts.CollapseReplicas = true
	n := NewNormalizer(opts)
	spec := ServiceSpec{Name: "svc", Image: "img", Replicas: 0}
	out := n.Normalize(spec)
	if out.Replicas != 1 {
		t.Errorf("expected replicas=1, got %d", out.Replicas)
	}
}

func TestNormalizer_Normalize_CollapseReplicas_PositiveUnchanged(t *testing.T) {
	opts := DefaultNormalizerOptions()
	opts.CollapseReplicas = true
	n := NewNormalizer(opts)
	spec := ServiceSpec{Name: "svc", Image: "img", Replicas: 3}
	out := n.Normalize(spec)
	if out.Replicas != 3 {
		t.Errorf("expected replicas=3, got %d", out.Replicas)
	}
}

func TestNormalizer_NormalizeAll_Length(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerOptions())
	specs := []ServiceSpec{
		{Name: "a", Image: "IMG:1", Replicas: 1},
		{Name: "b", Image: "IMG:2", Replicas: 2},
	}
	out := n.NormalizeAll(specs)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if out[0].Image != "img:1" {
		t.Errorf("expected 'img:1', got %q", out[0].Image)
	}
	if out[1].Image != "img:2" {
		t.Errorf("expected 'img:2', got %q", out[1].Image)
	}
}
