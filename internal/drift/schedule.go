package drift

import (
	"errors"
	"time"
)

// Schedule defines how often drift checks should be performed.
type Schedule struct {
	Interval time.Duration
	MaxRuns  int // 0 means unlimited
}

// DefaultSchedule returns a Schedule with sensible defaults.
func DefaultSchedule() Schedule {
	return Schedule{
		Interval: 30 * time.Second,
		MaxRuns:  0,
	}
}

// Validate checks that the Schedule has valid field values.
func (s Schedule) Validate() error {
	if s.Interval <= 0 {
		return errors.New("schedule interval must be greater than zero")
	}
	if s.MaxRuns < 0 {
		return errors.New("schedule max runs must be non-negative")
	}
	return nil
}

// IsLimited reports whether the schedule has a finite run limit.
func (s Schedule) IsLimited() bool {
	return s.MaxRuns > 0
}

// NextTick returns a channel that fires after the schedule interval.
func (s Schedule) NextTick() <-chan time.Time {
	return time.After(s.Interval)
}
