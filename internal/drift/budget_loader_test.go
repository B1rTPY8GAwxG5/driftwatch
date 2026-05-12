package drift

import (
	"os"
	"path/filepath"
	"testing"
)

const validBudgetYAML = `
limit: 10
period: hour
`

func TestLoadBudgetConfigFromBytes_Valid(t *testing.T) {
	cfg, err := LoadBudgetConfigFromBytes([]byte(validBudgetYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 10 {
		t.Errorf("expected limit 10, got %d", cfg.Limit)
	}
	if cfg.Period != "hour" {
		t.Errorf("expected period 'hour', got %q", cfg.Period)
	}
}

func TestLoadBudgetConfigFromBytes_ZeroLimit(t *testing.T) {
	_, err := LoadBudgetConfigFromBytes([]byte("limit: 0\nperiod: day\n"))
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestLoadBudgetConfigFromBytes_MissingPeriod(t *testing.T) {
	_, err := LoadBudgetConfigFromBytes([]byte("limit: 5\n"))
	if err == nil {
		t.Fatal("expected error for missing period")
	}
}

func TestLoadBudgetConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadBudgetConfigFromBytes([]byte(":::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadBudgetConfig_FileNotFound(t *testing.T) {
	_, err := LoadBudgetConfig("/nonexistent/budget.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadBudgetConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "budget.yaml")
	if err := os.WriteFile(p, []byte(validBudgetYAML), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg, err := LoadBudgetConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 10 {
		t.Errorf("expected limit 10, got %d", cfg.Limit)
	}
}

func TestBuildDriftBudget_Valid(t *testing.T) {
	cfg := &BudgetConfig{Limit: 5, Period: "day"}
	b, err := BuildDriftBudget(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil budget")
	}
}

func TestBuildDriftBudget_NilConfig(t *testing.T) {
	_, err := BuildDriftBudget(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestBuildDriftBudget_InvalidPeriod(t *testing.T) {
	cfg := &BudgetConfig{Limit: 3, Period: "fortnight"}
	_, err := BuildDriftBudget(cfg)
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}
