package drift

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultStalenessPolicy_Values(t *testing.T) {
	p := DefaultStalenessPolicy()
	if p.WarnAfter != 5*time.Minute {
		t.Errorf("expected WarnAfter 5m, got %s", p.WarnAfter)
	}
	if p.StaleAfter != 15*time.Minute {
		t.Errorf("expected StaleAfter 15m, got %s", p.StaleAfter)
	}
}

func TestStalenessLevel_String(t *testing.T) {
	cases := []struct {
		level StalenessLevel
		want  string
	}{
		{StalenessLevelFresh, "fresh"},
		{StalenessLevelWarn, "warn"},
		{StalenessLevelStale, "stale"},
		{StalenessLevel(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("level %d: expected %q, got %q", tc.level, tc.want, got)
		}
	}
}

func TestStalenessChecker_Fresh(t *testing.T) {
	checker := NewStalenessChecker(DefaultStalenessPolicy())
	checker.now = func() time.Time { return time.Now() }
	observedAt := time.Now().Add(-1 * time.Minute)
	if got := checker.Check(observedAt); got != StalenessLevelFresh {
		t.Errorf("expected fresh, got %s", got)
	}
}

func TestStalenessChecker_Warn(t *testing.T) {
	checker := NewStalenessChecker(DefaultStalenessPolicy())
	fixed := time.Now()
	checker.now = func() time.Time { return fixed }
	observedAt := fixed.Add(-7 * time.Minute)
	if got := checker.Check(observedAt); got != StalenessLevelWarn {
		t.Errorf("expected warn, got %s", got)
	}
}

func TestStalenessChecker_Stale(t *testing.T) {
	checker := NewStalenessChecker(DefaultStalenessPolicy())
	fixed := time.Now()
	checker.now = func() time.Time { return fixed }
	observedAt := fixed.Add(-20 * time.Minute)
	if got := checker.Check(observedAt); got != StalenessLevelStale {
		t.Errorf("expected stale, got %s", got)
	}
}

func TestStalenessChecker_Describe_ContainsLevel(t *testing.T) {
	checker := NewStalenessChecker(DefaultStalenessPolicy())
	fixed := time.Now()
	checker.now = func() time.Time { return fixed }
	observedAt := fixed.Add(-20 * time.Minute)
	desc := checker.Describe(observedAt)
	if !strings.Contains(desc, "stale") {
		t.Errorf("expected description to contain 'stale', got %q", desc)
	}
	if !strings.Contains(desc, "age:") {
		t.Errorf("expected description to contain 'age:', got %q", desc)
	}
}

func TestStalenessChecker_Describe_FreshContainsAge(t *testing.T) {
	checker := NewStalenessChecker(DefaultStalenessPolicy())
	fixed := time.Now()
	checker.now = func() time.Time { return fixed }
	observedAt := fixed.Add(-30 * time.Second)
	desc := checker.Describe(observedAt)
	if !strings.Contains(desc, "fresh") {
		t.Errorf("expected 'fresh' in description, got %q", desc)
	}
}
