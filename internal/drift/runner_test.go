package drift

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func makeSpec(name string) ServiceSpec {
	return ServiceSpec{
		Name:     name,
		Image:    "app:latest",
		Replicas: 2,
		Env:      map[string]string{"PORT": "8080"},
	}
}

func TestNewRunner_NilDetector(t *testing.T) {
	_, err := NewRunner(nil, DefaultSchedule(), nil, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for nil detector")
	}
}

func TestNewRunner_InvalidSchedule(t *testing.T) {
	d := NewDetector()
	_, err := NewRunner(d, Schedule{Interval: 0}, nil, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for invalid schedule")
	}
}

func TestNewRunner_NilWriter(t *testing.T) {
	d := NewDetector()
	_, err := NewRunner(d, DefaultSchedule(), nil, nil)
	if err == nil {
		t.Error("expected error for nil writer")
	}
}

func TestNewRunner_Valid(t *testing.T) {
	d := NewDetector()
	r, err := NewRunner(d, DefaultSchedule(), nil, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Error("expected non-nil runner")
	}
}

func TestRunOnce_NoDrift(t *testing.T) {
	d := NewDetector()
	spec := makeSpec("svc")
	var buf bytes.Buffer
	r, _ := NewRunner(d, DefaultSchedule(), []ServiceSpec{spec}, &buf)

	liveFn := func(name string) (*ServiceSpec, error) {
		c := makeSpec(name)
		return &c, nil
	}
	if err := r.RunOnce(context.Background(), liveFn); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunOnce_LiveFetchError(t *testing.T) {
	d := NewDetector()
	spec := makeSpec("svc")
	var buf bytes.Buffer
	r, _ := NewRunner(d, DefaultSchedule(), []ServiceSpec{spec}, &buf)

	liveFn := func(name string) (*ServiceSpec, error) {
		return nil, errors.New("fetch failed")
	}
	if err := r.RunOnce(context.Background(), liveFn); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected error message written to output")
	}
}

func TestStart_MaxRuns(t *testing.T) {
	d := NewDetector()
	spec := makeSpec("svc")
	var buf bytes.Buffer
	sched := Schedule{Interval: 10 * time.Millisecond, MaxRuns: 2}
	r, _ := NewRunner(d, sched, []ServiceSpec{spec}, &buf)

	calls := 0
	liveFn := func(name string) (*ServiceSpec, error) {
		calls++
		c := makeSpec(name)
		return &c, nil
	}
	if err := r.Start(context.Background(), liveFn); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}
