package drift

import (
	"testing"
	"time"
)

func TestDefaultCooldownPolicy_Values(t *testing.T) {
	p := DefaultCooldownPolicy()
	if p.Duration != 5*time.Minute {
		t.Errorf("expected 5m, got %v", p.Duration)
	}
}

func TestNewCooldown_ZeroPolicy_UsesDefaults(t *testing.T) {
	c := NewCooldown(CooldownPolicy{})
	if c.policy.Duration != DefaultCooldownPolicy().Duration {
		t.Errorf("expected default duration, got %v", c.policy.Duration)
	}
}

func TestNewCooldown_NotNil(t *testing.T) {
	c := NewCooldown(DefaultCooldownPolicy())
	if c == nil {
		t.Fatal("expected non-nil Cooldown")
	}
}

func TestCooldown_Allow_FirstCall_True(t *testing.T) {
	c := NewCooldown(CooldownPolicy{Duration: time.Second})
	if !c.Allow("svc-a") {
		t.Error("expected first call to Allow to return true")
	}
}

func TestCooldown_Allow_SecondCall_BlockedWithinPeriod(t *testing.T) {
	c := NewCooldown(CooldownPolicy{Duration: time.Hour})
	c.Allow("svc-a")
	if c.Allow("svc-a") {
		t.Error("expected second call within cooldown to return false")
	}
}

func TestCooldown_Allow_PermittedAfterPeriod(t *testing.T) {
	now := time.Now()
	c := NewCooldown(CooldownPolicy{Duration: time.Second})
	c.now = func() time.Time { return now }
	c.Allow("svc-a")

	c.now = func() time.Time { return now.Add(2 * time.Second) }
	if !c.Allow("svc-a") {
		t.Error("expected Allow to return true after cooldown expires")
	}
}

func TestCooldown_Allow_IndependentServices(t *testing.T) {
	c := NewCooldown(CooldownPolicy{Duration: time.Hour})
	c.Allow("svc-a")
	if !c.Allow("svc-b") {
		t.Error("expected independent service to be allowed")
	}
}

func TestCooldown_Reset_ClearsService(t *testing.T) {
	c := NewCooldown(CooldownPolicy{Duration: time.Hour})
	c.Allow("svc-a")
	c.Reset("svc-a")
	if !c.Allow("svc-a") {
		t.Error("expected Allow to return true after Reset")
	}
}

func TestCooldown_Flush_ClearsAll(t *testing.T) {
	c := NewCooldown(CooldownPolicy{Duration: time.Hour})
	c.Allow("svc-a")
	c.Allow("svc-b")
	c.Flush()
	if !c.Allow("svc-a") || !c.Allow("svc-b") {
		t.Error("expected all services to be allowed after Flush")
	}
}

func TestCooldown_Active_CountsInCooldown(t *testing.T) {
	now := time.Now()
	c := NewCooldown(CooldownPolicy{Duration: time.Hour})
	c.now = func() time.Time { return now }
	c.Allow("svc-a")
	c.Allow("svc-b")

	if got := c.Active(); got != 2 {
		t.Errorf("expected 2 active cooldowns, got %d", got)
	}
}

func TestCooldown_Active_ExcludesExpired(t *testing.T) {
	now := time.Now()
	c := NewCooldown(CooldownPolicy{Duration: time.Second})
	c.now = func() time.Time { return now }
	c.Allow("svc-a")

	c.now = func() time.Time { return now.Add(2 * time.Second) }
	if got := c.Active(); got != 0 {
		t.Errorf("expected 0 active cooldowns after expiry, got %d", got)
	}
}
