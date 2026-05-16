package drift

import (
	"fmt"
	"time"
)

// MaturityLevel represents how mature (stable) a service's drift posture is.
type MaturityLevel int

const (
	MaturityUnknown  MaturityLevel = iota
	MaturityUnstable               // frequent drift detected
	MaturityDeveloping             // occasional drift
	MaturityStable                 // rare drift
	MaturityMature                 // no drift over observation window
)

func (m MaturityLevel) String() string {
	switch m {
	case MaturityUnstable:
		return "unstable"
	case MaturityDeveloping:
		return "developing"
	case MaturityStable:
		return "stable"
	case MaturityMature:
		return "mature"
	default:
		return "unknown"
	}
}

// MaturityRecord holds the computed maturity for a single service.
type MaturityRecord struct {
	Service   string
	Level     MaturityLevel
	DriftRate float64 // fraction of observations that were drifted [0,1]
	Observed  int
	ComputedAt time.Time
}

// MaturityModel computes maturity levels from historical observations.
type MaturityModel struct {
	UnstableThreshold   float64 // drift rate >= this → Unstable
	DevelopingThreshold float64 // drift rate >= this → Developing
	StableThreshold     float64 // drift rate >= this → Stable
	observations        map[string][]bool
}

// NewMaturityModel returns a MaturityModel with sensible defaults.
func NewMaturityModel() *MaturityModel {
	return &MaturityModel{
		UnstableThreshold:   0.5,
		DevelopingThreshold: 0.2,
		StableThreshold:     0.05,
		observations:        make(map[string][]bool),
	}
}

// Record adds a drift observation for the given service.
func (m *MaturityModel) Record(service string, drifted bool) {
	m.observations[service] = append(m.observations[service], drifted)
}

// Evaluate computes the MaturityRecord for a service.
func (m *MaturityModel) Evaluate(service string) (MaturityRecord, error) {
	obs, ok := m.observations[service]
	if !ok || len(obs) == 0 {
		return MaturityRecord{}, fmt.Errorf("maturity: no observations for service %q", service)
	}
	var drifted int
	for _, d := range obs {
		if d {
			drifted++
		}
	}
	rate := float64(drifted) / float64(len(obs))
	level := m.levelFor(rate)
	return MaturityRecord{
		Service:    service,
		Level:      level,
		DriftRate:  rate,
		Observed:   len(obs),
		ComputedAt: time.Now(),
	}, nil
}

func (m *MaturityModel) levelFor(rate float64) MaturityLevel {
	switch {
	case rate >= m.UnstableThreshold:
		return MaturityUnstable
	case rate >= m.DevelopingThreshold:
		return MaturityDeveloping
	case rate >= m.StableThreshold:
		return MaturityStable
	default:
		return MaturityMature
	}
}

// Services returns all services that have observations.
func (m *MaturityModel) Services() []string {
	out := make([]string, 0, len(m.observations))
	for k := range m.observations {
		out = append(out, k)
	}
	return out
}
