package drift

import (
	"testing"
	"time"
)

func TestNewDecayModel_Defaults(t *testing.T) {
	d := NewDecayModel(0, 0)
	if d.halfLife != 10*time.Minute {
		t.Errorf("expected default half-life 10m, got %v", d.halfLife)
	}
	if d.floor != 0 {
		t.Errorf("expected floor 0, got %v", d.floor)
	}
}

func TestNewDecayModel_NegativeFloor_Clamped(t *testing.T) {
	d := NewDecayModel(time.Minute, -5)
	if d.floor != 0 {
		t.Errorf("expected floor clamped to 0, got %v", d.floor)
	}
}

func TestDecayModel_Score_NoRecord_ReturnsFloor(t *testing.T) {
	d := NewDecayModel(time.Minute, 1.0)
	s := d.Score("svc-a")
	if s != 1.0 {
		t.Errorf("expected floor 1.0, got %v", s)
	}
}

func TestDecayModel_Record_IncreasesScore(t *testing.T) {
	d := NewDecayModel(time.Hour, 0)
	d.Record("svc-a", 10.0)
	s := d.Score("svc-a")
	if s < 9.9 {
		t.Errorf("expected score near 10, got %v", s)
	}
}

func TestDecayModel_Record_Accumulates(t *testing.T) {
	d := NewDecayModel(time.Hour, 0)
	d.Record("svc-a", 5.0)
	d.Record("svc-a", 5.0)
	s := d.Score("svc-a")
	if s < 9.0 {
		t.Errorf("expected accumulated score near 10, got %v", s)
	}
}

func TestDecayModel_Score_DecaysOverTime(t *testing.T) {
	halfLife := 100 * time.Millisecond
	d := NewDecayModel(halfLife, 0)
	d.Record("svc-b", 100.0)
	time.Sleep(halfLife)
	s := d.Score("svc-b")
	// After one half-life the score should be near 50
	if s > 60 || s < 30 {
		t.Errorf("expected score near 50 after half-life, got %v", s)
	}
}

func TestDecayModel_Reset_ClearsScore(t *testing.T) {
	d := NewDecayModel(time.Hour, 0)
	d.Record("svc-c", 42.0)
	d.Reset("svc-c")
	s := d.Score("svc-c")
	if s != 0 {
		t.Errorf("expected 0 after reset, got %v", s)
	}
}

func TestDecayModel_IndependentServices(t *testing.T) {
	d := NewDecayModel(time.Hour, 0)
	d.Record("svc-x", 20.0)
	d.Record("svc-y", 80.0)
	if d.Score("svc-x") > d.Score("svc-y") {
		t.Errorf("svc-y should have higher score than svc-x")
	}
}
