package drift

import (
	"testing"
)

func TestNewSampler_DefaultRate(t *testing.T) {
	s := NewSampler(1.0, SampleModeRandom)
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate 1.0, got %f", s.Rate())
	}
}

func TestNewSampler_ClampsNegativeRate(t *testing.T) {
	s := NewSampler(-0.5, SampleModeRandom)
	if s.Rate() != 0 {
		t.Fatalf("expected rate 0, got %f", s.Rate())
	}
}

func TestNewSampler_ClampsExceedingRate(t *testing.T) {
	s := NewSampler(1.5, SampleModeRandom)
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate 1.0, got %f", s.Rate())
	}
}

func TestSampler_ShouldSample_ZeroRate_AlwaysFalse(t *testing.T) {
	s := NewSampler(0, SampleModeRandom)
	for i := 0; i < 20; i++ {
		if s.ShouldSample() {
			t.Fatal("expected no samples at rate 0")
		}
	}
}

func TestSampler_ShouldSample_FullRate_AlwaysTrue(t *testing.T) {
	s := NewSampler(1.0, SampleModeRandom)
	for i := 0; i < 20; i++ {
		if !s.ShouldSample() {
			t.Fatal("expected all samples at rate 1.0")
		}
	}
}

func TestSampler_RoundRobin_SamplesEveryNth(t *testing.T) {
	// rate=0.5 → step=2, so every 2nd call should be sampled
	s := NewSampler(0.5, SampleModeRoundRobin)
	results := make([]bool, 10)
	for i := range results {
		results[i] = s.ShouldSample()
	}
	// odd indices (1,3,5…) should be sampled (cursor 1,3,5… % 2 == 1 != 0)
	// even indices (0,2,4…) cursor 0,2,4… % 2 == 0 — wait, cursor starts at 0
	// cursor increments first, so calls: 1%2!=0→false, 2%2==0→true, 3%2!=0→false …
	expected := []bool{false, true, false, true, false, true, false, true, false, true}
	for i, got := range results {
		if got != expected[i] {
			t.Errorf("call %d: expected %v, got %v", i, expected[i], got)
		}
	}
}

func TestSampler_Mode_ReturnsConfigured(t *testing.T) {
	s := NewSampler(0.5, SampleModeRoundRobin)
	if s.Mode() != SampleModeRoundRobin {
		t.Fatalf("expected round-robin mode, got %s", s.Mode())
	}
}

func TestSampler_Random_ApproximateRate(t *testing.T) {
	s := NewSampler(0.5, SampleModeRandom)
	hits := 0
	n := 10000
	for i := 0; i < n; i++ {
		if s.ShouldSample() {
			hits++
		}
	}
	ratio := float64(hits) / float64(n)
	if ratio < 0.45 || ratio > 0.55 {
		t.Errorf("expected ~0.5 sample rate, got %f", ratio)
	}
}
