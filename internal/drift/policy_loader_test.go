package drift

import (
	"testing"
)

var validPolicyYAML = []byte(`
policies:
  - name: default
    rules:
      - kind: image
        action: block
      - kind: replicas
        action: warn
  - name: relaxed
    rules:
      - kind: image
        action: warn
`)

func TestLoadPolicyFromBytes_Valid(t *testing.T) {
	p, err := LoadPolicyFromBytes(validPolicyYAML, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "default" {
		t.Errorf("expected name=default, got %s", p.Name)
	}
	if len(p.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(p.Rules))
	}
}

func TestLoadPolicyFromBytes_NotFound(t *testing.T) {
	_, err := LoadPolicyFromBytes(validPolicyYAML, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing policy")
	}
}

func TestLoadPolicyFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadPolicyFromBytes([]byte("::bad yaml::"), "default")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadPolicy_FileNotFound(t *testing.T) {
	_, err := LoadPolicy("/nonexistent/policy.yaml", "default")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadPolicy_ValidFile(t *testing.T) {
	p, err := LoadPolicy("../../testdata/policy-example.yaml", "strict")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "strict" {
		t.Errorf("expected name=strict, got %s", p.Name)
	}
}
