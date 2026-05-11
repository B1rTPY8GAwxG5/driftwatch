package drift

import (
	"testing"
	"time"
)

func TestNewRateLimiter_Defaults(t *testing.T) {
	rl := NewRateLimiter(0, 0)
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
	if rl.max != 10 {
		t.Errorf("expected default max 10, got %d", rl.max)
	}
	if rl.window != time.Minute {
		t.Errorf("expected default window 1m, got %v", rl.window)
	}
}

func TestNewRateLimiter_CustomValues(t *testing.T) {
	rl := NewRateLimiter(5, 30*time.Second)
	if rl.max != 5 {
		t.Errorf("expected max 5, got %d", rl.max)
	}
	if rl.window != 30*time.Second {
		t.Errorf("expected window 30s, got %v", rl.window)
	}
}

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow("svc-a") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	rl.Allow("svc-b")
	rl.Allow("svc-b")
	if rl.Allow("svc-b") {
		t.Error("expected Allow to return false when limit exceeded")
	}
}

func TestRateLimiter_Allow_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	if !rl.Allow("svc-x") {
		t.Error("expected Allow true for svc-x")
	}
	if !rl.Allow("svc-y") {
		t.Error("expected Allow true for svc-y (independent key)")
	}
	if rl.Allow("svc-x") {
		t.Error("expected Allow false for svc-x after limit")
	}
}

func TestRateLimiter_Reset_ClearsKey(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	rl.Allow("svc-c")
	rl.Reset("svc-c")
	if !rl.Allow("svc-c") {
		t.Error("expected Allow true after Reset")
	}
}

func TestRateLimiter_Count_ReflectsAttempts(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	rl.Allow("svc-d")
	rl.Allow("svc-d")
	if c := rl.Count("svc-d"); c != 2 {
		t.Errorf("expected count 2, got %d", c)
	}
}

func TestRateLimiter_Count_EmptyKey(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	if c := rl.Count("unknown"); c != 0 {
		t.Errorf("expected count 0 for unknown key, got %d", c)
	}
}

func TestErrRateLimited_NotNil(t *testing.T) {
	if ErrRateLimited == nil {
		t.Error("expected ErrRateLimited to be non-nil")
	}
}
