package drift

import (
	"testing"
	"time"
)

func TestNewDriftBudget_Valid(t *testing.T) {
	b, err := NewDriftBudget(5, BudgetPeriodHour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil budget")
	}
}

func TestNewDriftBudget_ZeroLimit(t *testing.T) {
	_, err := NewDriftBudget(0, BudgetPeriodHour)
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestNewDriftBudget_UnknownPeriod(t *testing.T) {
	_, err := NewDriftBudget(5, BudgetPeriod("month"))
	if err == nil {
		t.Fatal("expected error for unknown period")
	}
}

func TestDriftBudget_Record_UnderLimit(t *testing.T) {
	b, _ := NewDriftBudget(3, BudgetPeriodHour)
	if !b.Record() {
		t.Error("expected first record to succeed")
	}
	if !b.Record() {
		t.Error("expected second record to succeed")
	}
}

func TestDriftBudget_Record_ExceedsLimit(t *testing.T) {
	b, _ := NewDriftBudget(2, BudgetPeriodHour)
	b.Record()
	b.Record()
	if b.Record() {
		t.Error("expected third record to be rejected")
	}
}

func TestDriftBudget_Remaining_DecreasesOnRecord(t *testing.T) {
	b, _ := NewDriftBudget(5, BudgetPeriodDay)
	if b.Remaining() != 5 {
		t.Errorf("expected 5 remaining, got %d", b.Remaining())
	}
	b.Record()
	if b.Remaining() != 4 {
		t.Errorf("expected 4 remaining, got %d", b.Remaining())
	}
}

func TestDriftBudget_Exhausted_True(t *testing.T) {
	b, _ := NewDriftBudget(1, BudgetPeriodHour)
	b.Record()
	if !b.Exhausted() {
		t.Error("expected budget to be exhausted")
	}
}

func TestDriftBudget_Exhausted_False(t *testing.T) {
	b, _ := NewDriftBudget(3, BudgetPeriodHour)
	b.Record()
	if b.Exhausted() {
		t.Error("expected budget to not be exhausted")
	}
}

func TestDriftBudget_Reset_ClearsEvents(t *testing.T) {
	b, _ := NewDriftBudget(2, BudgetPeriodHour)
	b.Record()
	b.Record()
	b.Reset()
	if b.Remaining() != 2 {
		t.Errorf("expected 2 remaining after reset, got %d", b.Remaining())
	}
}

func TestDriftBudget_Prune_RemovesExpiredEvents(t *testing.T) {
	b, _ := NewDriftBudget(3, BudgetPeriodHour)
	now := time.Now()
	// Simulate old events recorded 2 hours ago.
	b.mu.Lock()
	b.events = []time.Time{now.Add(-2 * time.Hour), now.Add(-90 * time.Minute)}
	b.mu.Unlock()
	if b.Remaining() != 3 {
		t.Errorf("expected expired events to be pruned, got remaining=%d", b.Remaining())
	}
}

func TestBudgetPeriod_Week(t *testing.T) {
	b, err := NewDriftBudget(10, BudgetPeriodWeek)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.period != 7*24*time.Hour {
		t.Errorf("expected week period, got %v", b.period)
	}
}
