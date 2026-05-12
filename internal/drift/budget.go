package drift

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// BudgetPeriod defines the rolling window for drift budget tracking.
type BudgetPeriod string

const (
	BudgetPeriodHour  BudgetPeriod = "hour"
	BudgetPeriodDay   BudgetPeriod = "day"
	BudgetPeriodWeek  BudgetPeriod = "week"
)

// DriftBudget tracks how many drift events are allowed within a period.
type DriftBudget struct {
	mu       sync.Mutex
	limit    int
	period   time.Duration
	events   []time.Time
	clock    func() time.Time
}

// NewDriftBudget creates a budget allowing at most limit drift events per period.
func NewDriftBudget(limit int, period BudgetPeriod) (*DriftBudget, error) {
	if limit <= 0 {
		return nil, errors.New("budget limit must be positive")
	}
	d, err := parseBudgetPeriod(period)
	if err != nil {
		return nil, err
	}
	return &DriftBudget{
		limit:  limit,
		period: d,
		clock:  time.Now,
	}, nil
}

func parseBudgetPeriod(p BudgetPeriod) (time.Duration, error) {
	switch p {
	case BudgetPeriodHour:
		return time.Hour, nil
	case BudgetPeriodDay:
		return 24 * time.Hour, nil
	case BudgetPeriodWeek:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown budget period: %q", p)
	}
}

// Record registers a drift event at the current time.
// Returns true if the budget is not yet exhausted, false if the limit is exceeded.
func (b *DriftBudget) Record() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	b.prune(now)
	if len(b.events) >= b.limit {
		return false
	}
	b.events = append(b.events, now)
	return true
}

// Remaining returns how many drift events can still be recorded in the current window.
func (b *DriftBudget) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prune(b.clock())
	r := b.limit - len(b.events)
	if r < 0 {
		return 0
	}
	return r
}

// Exhausted reports whether the budget is fully consumed.
func (b *DriftBudget) Exhausted() bool {
	return b.Remaining() == 0
}

// Reset clears all recorded events.
func (b *DriftBudget) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = nil
}

func (b *DriftBudget) prune(now time.Time) {
	cutoff := now.Add(-b.period)
	i := 0
	for i < len(b.events) && b.events[i].Before(cutoff) {
		i++
	}
	b.events = b.events[i:]
}
