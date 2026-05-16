package drift

import (
	"testing"
)

func TestNewQuotaEnforcer_Valid(t *testing.T) {
	q, err := NewQuotaEnforcer(QuotaPolicy{Limit: 5, Period: QuotaPeriodHour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil enforcer")
	}
}

func TestNewQuotaEnforcer_ZeroLimit(t *testing.T) {
	_, err := NewQuotaEnforcer(QuotaPolicy{Limit: 0, Period: QuotaPeriodHour})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestNewQuotaEnforcer_UnknownPeriod(t *testing.T) {
	_, err := NewQuotaEnforcer(QuotaPolicy{Limit: 3, Period: "weekly"})
	if err == nil {
		t.Fatal("expected error for unknown period")
	}
}

func TestQuotaEnforcer_Allow_UnderLimit(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 3, Period: QuotaPeriodMinute})
	for i := 0; i < 3; i++ {
		if !q.Allow("svc-a") {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
}

func TestQuotaEnforcer_Allow_ExceedsLimit(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 2, Period: QuotaPeriodMinute})
	q.Allow("svc-b")
	q.Allow("svc-b")
	if q.Allow("svc-b") {
		t.Fatal("expected deny after limit exceeded")
	}
}

func TestQuotaEnforcer_Allow_IndependentServices(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 1, Period: QuotaPeriodHour})
	q.Allow("svc-x")
	if !q.Allow("svc-y") {
		t.Fatal("expected svc-y to be allowed independently")
	}
}

func TestQuotaEnforcer_Remaining_Full(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 5, Period: QuotaPeriodDay})
	if got := q.Remaining("svc-c"); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestQuotaEnforcer_Remaining_AfterUse(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 4, Period: QuotaPeriodHour})
	q.Allow("svc-d")
	q.Allow("svc-d")
	if got := q.Remaining("svc-d"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestQuotaEnforcer_Reset_ClearsState(t *testing.T) {
	q, _ := NewQuotaEnforcer(QuotaPolicy{Limit: 1, Period: QuotaPeriodMinute})
	q.Allow("svc-e")
	if q.Allow("svc-e") {
		t.Fatal("expected deny before reset")
	}
	q.Reset("svc-e")
	if !q.Allow("svc-e") {
		t.Fatal("expected allow after reset")
	}
}

func TestQuotaPeriod_Validate_AllValid(t *testing.T) {
	periods := []QuotaPeriod{QuotaPeriodMinute, QuotaPeriodHour, QuotaPeriodDay}
	for _, p := range periods {
		pol := QuotaPolicy{Limit: 1, Period: p}
		if err := pol.Validate(); err != nil {
			t.Errorf("unexpected error for period %q: %v", p, err)
		}
	}
}
