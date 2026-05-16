package drift

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewStepper_NotNil(t *testing.T) {
	s := NewStepper("svc-a")
	if s == nil {
		t.Fatal("expected non-nil Stepper")
	}
}

func TestStepper_Steps_InitiallyEmpty(t *testing.T) {
	s := NewStepper("svc-a")
	if len(s.Steps()) != 0 {
		t.Fatalf("expected 0 steps, got %d", len(s.Steps()))
	}
}

func TestStepper_Record_AddsStep(t *testing.T) {
	s := NewStepper("svc-a")
	s.Record(StepKindDetect, "run-detector", 5*time.Millisecond, nil)
	steps := s.Steps()
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Kind != StepKindDetect {
		t.Errorf("expected kind %q, got %q", StepKindDetect, steps[0].Kind)
	}
	if steps[0].Name != "run-detector" {
		t.Errorf("expected name %q, got %q", "run-detector", steps[0].Name)
	}
}

func TestStepper_Skip_MarksSkipped(t *testing.T) {
	s := NewStepper("svc-b")
	s.Skip(StepKindNotify, "send-alert")
	steps := s.Steps()
	if !steps[0].Skipped {
		t.Error("expected step to be marked skipped")
	}
}

func TestStepper_HasError_False_WhenNoErrors(t *testing.T) {
	s := NewStepper("svc-a")
	s.Record(StepKindEvaluate, "policy-check", time.Millisecond, nil)
	if s.HasError() {
		t.Error("expected HasError to be false")
	}
}

func TestStepper_HasError_True_WhenErrorPresent(t *testing.T) {
	s := NewStepper("svc-a")
	s.Record(StepKindRemediate, "apply-patch", 2*time.Millisecond, errors.New("timeout"))
	if !s.HasError() {
		t.Error("expected HasError to be true")
	}
}

func TestStepper_Steps_ReturnsCopy(t *testing.T) {
	s := NewStepper("svc-c")
	s.Record(StepKindDetect, "detect", time.Millisecond, nil)
	a := s.Steps()
	a[0].Name = "mutated"
	b := s.Steps()
	if b[0].Name == "mutated" {
		t.Error("Steps should return an independent copy")
	}
}

func TestStepper_Summary_ContainsService(t *testing.T) {
	s := NewStepper("my-service")
	s.Record(StepKindDetect, "detect", 3*time.Millisecond, nil)
	sum := s.Summary()
	if !strings.Contains(sum, "my-service") {
		t.Errorf("summary should contain service name, got: %s", sum)
	}
}

func TestStepper_Summary_ContainsErrorStatus(t *testing.T) {
	s := NewStepper("svc-err")
	s.Record(StepKindRemediate, "patch", time.Millisecond, errors.New("forbidden"))
	sum := s.Summary()
	if !strings.Contains(sum, "forbidden") {
		t.Errorf("summary should contain error text, got: %s", sum)
	}
}

func TestStepper_Summary_ContainsSkippedStatus(t *testing.T) {
	s := NewStepper("svc-skip")
	s.Skip(StepKindNotify, "email-alert")
	sum := s.Summary()
	if !strings.Contains(sum, "skipped") {
		t.Errorf("summary should contain 'skipped', got: %s", sum)
	}
}

func TestStepKind_Constants(t *testing.T) {
	kinds := []StepKind{StepKindDetect, StepKindEvaluate, StepKindNotify, StepKindRemediate}
	seen := map[StepKind]bool{}
	for _, k := range kinds {
		if seen[k] {
			t.Errorf("duplicate StepKind value: %q", k)
		}
		seen[k] = true
	}
}
