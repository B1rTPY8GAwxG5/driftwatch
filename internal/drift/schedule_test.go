package drift

import (
	"testing"
	"time"
)

func TestDefaultSchedule_Values(t *testing.T) {
	s := DefaultSchedule()
	if s.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", s.Interval)
	}
	if s.MaxRuns != 0 {
		t.Errorf("expected MaxRuns=0, got %d", s.MaxRuns)
	}
}

func TestSchedule_Validate_Valid(t *testing.T) {
	s := Schedule{Interval: 10 * time.Second, MaxRuns: 5}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSchedule_Validate_ZeroInterval(t *testing.T) {
	s := Schedule{Interval: 0}
	if err := s.Validate(); err == nil {
		t.Error("expected error for zero interval")
	}
}

func TestSchedule_Validate_NegativeInterval(t *testing.T) {
	s := Schedule{Interval: -1 * time.Second}
	if err := s.Validate(); err == nil {
		t.Error("expected error for negative interval")
	}
}

func TestSchedule_Validate_NegativeMaxRuns(t *testing.T) {
	s := Schedule{Interval: 5 * time.Second, MaxRuns: -1}
	if err := s.Validate(); err == nil {
		t.Error("expected error for negative max runs")
	}
}

func TestSchedule_IsLimited_True(t *testing.T) {
	s := Schedule{Interval: 5 * time.Second, MaxRuns: 3}
	if !s.IsLimited() {
		t.Error("expected IsLimited to return true")
	}
}

func TestSchedule_IsLimited_False(t *testing.T) {
	s := Schedule{Interval: 5 * time.Second, MaxRuns: 0}
	if s.IsLimited() {
		t.Error("expected IsLimited to return false")
	}
}

func TestSchedule_NextTick_NotNil(t *testing.T) {
	s := Schedule{Interval: 50 * time.Millisecond}
	ch := s.NextTick()
	if ch == nil {
		t.Error("expected non-nil channel from NextTick")
	}
	select {
	case <-ch:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Error("NextTick channel did not fire in time")
	}
}
