package drift

import (
	"math/rand"
	"time"
)

// SampleMode controls how sampling decisions are made.
type SampleMode string

const (
	SampleModeRandom     SampleMode = "random"
	SampleModeRoundRobin SampleMode = "round-robin"
)

// Sampler decides whether a given drift check should be executed based on
// a configured rate and mode.
type Sampler struct {
	rate   float64
	mode   SampleMode
	cursor int
	step   int
	rng    *rand.Rand
}

// NewSampler creates a Sampler with the given rate (0.0–1.0) and mode.
// A rate of 1.0 means every check is sampled; 0.0 means none are.
func NewSampler(rate float64, mode SampleMode) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	step := 1
	if rate > 0 {
		step = int(1.0 / rate)
		if step < 1 {
			step = 1
		}
	}
	return &Sampler{
		rate:   rate,
		mode:   mode,
		step:   step,
		cursor: 0,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample returns true if the current invocation should be sampled.
func (s *Sampler) ShouldSample() bool {
	if s.rate <= 0 {
		return false
	}
	if s.rate >= 1 {
		return true
	}
	switch s.mode {
	case SampleModeRoundRobin:
		s.cursor++
		return s.cursor%s.step == 0
	default: // random
		return s.rng.Float64() < s.rate
	}
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 { return s.rate }

// Mode returns the configured sampling mode.
func (s *Sampler) Mode() SampleMode { return s.mode }
