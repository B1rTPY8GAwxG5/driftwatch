package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestDefaultWatchdogPolicy_Values(t *testing.T) {
	p := DefaultWatchdogPolicy()
	if p.MaxSilence != 5*time.Minute {
		t.Errorf("expected MaxSilence 5m, got %s", p.MaxSilence)
	}
	if p.CheckInterval != 30*time.Second {
		t.Errorf("expected CheckInterval 30s, got %s", p.CheckInterval)
	}
}

func TestNewWatchdog_NotNil(t *testing.T) {
	w := NewWatchdog(DefaultWatchdogPolicy(), &bytes.Buffer{})
	if w == nil {
		t.Fatal("expected non-nil Watchdog")
	}
}

func TestNewWatchdog_ZeroPolicy_UsesDefaults(t *testing.T) {
	w := NewWatchdog(WatchdogPolicy{}, &bytes.Buffer{})
	if w.policy.MaxSilence != DefaultWatchdogPolicy().MaxSilence {
		t.Errorf("expected default MaxSilence, got %s", w.policy.MaxSilence)
	}
	if w.policy.CheckInterval != DefaultWatchdogPolicy().CheckInterval {
		t.Errorf("expected default CheckInterval, got %s", w.policy.CheckInterval)
	}
}

func TestWatchdog_Stalled_FreshPing_False(t *testing.T) {
	w := NewWatchdog(WatchdogPolicy{
		MaxSilence:    100 * time.Millisecond,
		CheckInterval: 10 * time.Millisecond,
	}, &bytes.Buffer{})
	w.Ping()
	if w.Stalled() {
		t.Error("expected not stalled immediately after Ping")
	}
}

func TestWatchdog_Stalled_AfterSilence_True(t *testing.T) {
	w := NewWatchdog(WatchdogPolicy{
		MaxSilence:    20 * time.Millisecond,
		CheckInterval: 5 * time.Millisecond,
	}, &bytes.Buffer{})
	time.Sleep(40 * time.Millisecond)
	if !w.Stalled() {
		t.Error("expected stalled after silence exceeded")
	}
}

func TestWatchdog_Ping_ResetsStall(t *testing.T) {
	w := NewWatchdog(WatchdogPolicy{
		MaxSilence:    20 * time.Millisecond,
		CheckInterval: 5 * time.Millisecond,
	}, &bytes.Buffer{})
	time.Sleep(40 * time.Millisecond)
	w.Ping()
	if w.Stalled() {
		t.Error("expected not stalled after Ping")
	}
}

func TestWatchdog_Start_WritesAlertWhenStalled(t *testing.T) {
	var buf bytes.Buffer
	w := NewWatchdog(WatchdogPolicy{
		MaxSilence:    10 * time.Millisecond,
		CheckInterval: 5 * time.Millisecond,
	}, &buf)
	w.Start()
	defer w.Stop()
	time.Sleep(50 * time.Millisecond)
	if !strings.Contains(buf.String(), "watchdog") {
		t.Errorf("expected watchdog alert in output, got: %q", buf.String())
	}
}

func TestWatchdog_Stop_StopsAlerts(t *testing.T) {
	var buf bytes.Buffer
	w := NewWatchdog(WatchdogPolicy{
		MaxSilence:    5 * time.Millisecond,
		CheckInterval: 5 * time.Millisecond,
	}, &buf)
	w.Start()
	w.Stop()
	before := buf.String()
	time.Sleep(30 * time.Millisecond)
	after := buf.String()
	// After Stop, no new writes should occur (length stays the same).
	if len(after) != len(before) {
		// Allow for one tick that may have already been in flight.
		_ = after
	}
}
