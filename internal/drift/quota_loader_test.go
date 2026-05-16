package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadQuotaConfigFromBytes_Valid(t *testing.T) {
	yaml := []byte("limit: 10\nperiod: hour\n")
	pol, err := LoadQuotaConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Limit != 10 {
		t.Errorf("expected limit 10, got %d", pol.Limit)
	}
	if pol.Period != QuotaPeriodHour {
		t.Errorf("expected period hour, got %q", pol.Period)
	}
}

func TestLoadQuotaConfigFromBytes_ZeroLimit(t *testing.T) {
	yaml := []byte("limit: 0\nperiod: minute\n")
	_, err := LoadQuotaConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestLoadQuotaConfigFromBytes_MissingPeriod(t *testing.T) {
	yaml := []byte("limit: 5\n")
	_, err := LoadQuotaConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for missing period")
	}
}

func TestLoadQuotaConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadQuotaConfigFromBytes([]byte("::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadQuotaConfig_FileNotFound(t *testing.T) {
	_, err := LoadQuotaConfig("/nonexistent/quota.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadQuotaConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "quota.yaml")
	_ = os.WriteFile(p, []byte("limit: 20\nperiod: day\n"), 0o644)
	pol, err := LoadQuotaConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Limit != 20 {
		t.Errorf("expected 20, got %d", pol.Limit)
	}
}

func TestBuildQuotaEnforcer_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "quota.yaml")
	_ = os.WriteFile(p, []byte("limit: 3\nperiod: minute\n"), 0o644)
	enf, err := BuildQuotaEnforcer(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enf == nil {
		t.Fatal("expected non-nil enforcer")
	}
}
