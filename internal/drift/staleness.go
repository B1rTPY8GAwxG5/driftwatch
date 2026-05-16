package drift

import (
	"fmt"
	"time"
)

// StalenessPolicy defines thresholds for classifying how stale a drift result is.
type StalenessPolicy struct {
	WarnAfter  time.Duration
	StaleAfter time.Duration
}

// DefaultStalenessPolicy returns a StalenessPolicy with sensible defaults.
func DefaultStalenessPolicy() StalenessPolicy {
	return StalenessPolicy{
		WarnAfter:  5 * time.Minute,
		StaleAfter: 15 * time.Minute,
	}
}

// StalenessLevel represents how stale a result is considered.
type StalenessLevel int

const (
	StalenessLevelFresh  StalenessLevel = iota
	StalenessLevelWarn
	StalenessLevelStale
)

// String returns a human-readable label for the staleness level.
func (s StalenessLevel) String() string {
	switch s {
	case StalenessLevelFresh:
		return "fresh"
	case StalenessLevelWarn:
		return "warn"
	case StalenessLevelStale:
		return "stale"
	default:
		return "unknown"
	}
}

// StalenessChecker evaluates how stale a drift result is based on its timestamp.
type StalenessChecker struct {
	policy StalenessPolicy
	now    func() time.Time
}

// NewStalenessChecker creates a StalenessChecker with the given policy.
func NewStalenessChecker(policy StalenessPolicy) *StalenessChecker {
	return &StalenessChecker{
		policy: policy,
		now:    time.Now,
	}
}

// Check returns the StalenessLevel for a result observed at the given time.
func (sc *StalenessChecker) Check(observedAt time.Time) StalenessLevel {
	age := sc.now().Sub(observedAt)
	switch {
	case age >= sc.policy.StaleAfter:
		return StalenessLevelStale
	case age >= sc.policy.WarnAfter:
		return StalenessLevelWarn
	default:
		return StalenessLevelFresh
	}
}

// Describe returns a human-readable description of the staleness for the given time.
func (sc *StalenessChecker) Describe(observedAt time.Time) string {
	level := sc.Check(observedAt)
	age := sc.now().Sub(observedAt).Round(time.Second)
	return fmt.Sprintf("%s (age: %s)", level, age)
}
