package drift

import (
	"testing"
	"time"
)

func TestNewThrottle_DefaultInterval(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{})
	if th == nil {
		t.Fatal("expected non-nil Throttle")
	}
	if th.policy.MinInterval != time.Minute {
		t.Errorf("expected default MinInterval=1m, got %v", th.policy.MinInterval)
	}
}

func TestNewThrottle_CustomInterval(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: 5 * time.Second})
	if th.policy.MinInterval != 5*time.Second {
		t.Errorf("expected 5s, got %v", th.policy.MinInterval)
	}
}

func TestThrottle_Allow_FirstCall(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Hour})
	if !th.Allow("svc-a") {
		t.Error("expected first call to be allowed")
	}
}

func TestThrottle_Allow_BlockedWithinInterval(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Hour})
	th.Allow("svc-a")
	if th.Allow("svc-a") {
		t.Error("expected second call within interval to be blocked")
	}
}

func TestThrottle_Allow_PermittedAfterInterval(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Millisecond})
	th.Allow("svc-b")
	time.Sleep(5 * time.Millisecond)
	if !th.Allow("svc-b") {
		t.Error("expected call after interval to be allowed")
	}
}

func TestThrottle_Allow_IndependentServices(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Hour})
	th.Allow("svc-x")
	if !th.Allow("svc-y") {
		t.Error("expected independent service to be allowed")
	}
}

func TestThrottle_Reset_ClearsService(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Hour})
	th.Allow("svc-c")
	th.Reset("svc-c")
	if !th.Allow("svc-c") {
		t.Error("expected allow after reset")
	}
}

func TestThrottle_Flush_ClearsAll(t *testing.T) {
	th := NewThrottle(ThrottlePolicy{MinInterval: time.Hour})
	th.Allow("svc-d")
	th.Allow("svc-e")
	th.Flush()
	if !th.Allow("svc-d") || !th.Allow("svc-e") {
		t.Error("expected all services allowed after flush")
	}
}
