package drift

import (
	"strings"
	"testing"
)

func specA() ServiceSpec {
	return ServiceSpec{
		Name:     "svc-a",
		Image:    "myapp:1.0",
		Replicas: 3,
		Env:      map[string]string{"PORT": "8080", "MODE": "prod"},
	}
}

func TestDiff_NoDifferences(t *testing.T) {
	a := specA()
	b := specA()
	result := Diff(a, b)
	if len(result) != 0 {
		t.Fatalf("expected no diffs, got %v", result)
	}
}

func TestDiff_ImageChanged(t *testing.T) {
	a := specA()
	b := specA()
	b.Image = "myapp:2.0"
	result := Diff(a, b)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(result), result)
	}
	if !strings.Contains(result[0], "image") {
		t.Errorf("expected diff to mention 'image', got %q", result[0])
	}
}

func TestDiff_ReplicasChanged(t *testing.T) {
	a := specA()
	b := specA()
	b.Replicas = 1
	result := Diff(a, b)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(result), result)
	}
	if !strings.Contains(result[0], "replicas") {
		t.Errorf("expected diff to mention 'replicas', got %q", result[0])
	}
}

func TestDiff_EnvValueChanged(t *testing.T) {
	a := specA()
	b := specA()
	b.Env["PORT"] = "9090"
	result := Diff(a, b)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(result), result)
	}
	if !strings.Contains(result[0], "env.PORT") {
		t.Errorf("expected diff to mention 'env.PORT', got %q", result[0])
	}
}

func TestDiff_MissingEnvInLive(t *testing.T) {
	a := specA()
	b := specA()
	delete(b.Env, "MODE")
	result := Diff(a, b)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(result), result)
	}
	if !strings.Contains(result[0], "<missing>") {
		t.Errorf("expected diff to mention '<missing>', got %q", result[0])
	}
}

func TestDiff_WithIgnoreEnv(t *testing.T) {
	a := specA()
	b := specA()
	b.Env["PORT"] = "9090"
	result := Diff(a, b, WithIgnoreEnv())
	if len(result) != 0 {
		t.Fatalf("expected no diffs with IgnoreEnv, got %v", result)
	}
}

func TestDiff_ExtraEnvInLive(t *testing.T) {
	a := specA()
	b := specA()
	b.Env["EXTRA"] = "value"
	result := Diff(a, b)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(result), result)
	}
	if !strings.Contains(result[0], "env.EXTRA") {
		t.Errorf("expected diff to mention 'env.EXTRA', got %q", result[0])
	}
}
