package drift

import (
	"fmt"
	"strings"
	"time"
)

// StepKind identifies the type of a remediation step.
type StepKind string

const (
	StepKindDetect  StepKind = "detect"
	StepKindEvaluate StepKind = "evaluate"
	StepKindNotify  StepKind = "notify"
	StepKindRemediate StepKind = "remediate"
)

// Step represents a single stage in a drift response workflow.
type Step struct {
	Kind      StepKind
	Name      string
	RunAt     time.Time
	Duration  time.Duration
	Error     error
	Skipped   bool
}

// Stepper records and reports the sequential steps taken when responding to drift.
type Stepper struct {
	steps   []Step
	service string
}

// NewStepper creates a new Stepper for the given service.
func NewStepper(service string) *Stepper {
	return &Stepper{service: service}
}

// Record appends a completed step.
func (s *Stepper) Record(kind StepKind, name string, duration time.Duration, err error) {
	s.steps = append(s.steps, Step{
		Kind:     kind,
		Name:     name,
		RunAt:    time.Now(),
		Duration: duration,
		Error:    err,
	})
}

// Skip marks a step as intentionally skipped.
func (s *Stepper) Skip(kind StepKind, name string) {
	s.steps = append(s.steps, Step{
		Kind:    kind,
		Name:    name,
		RunAt:   time.Now(),
		Skipped: true,
	})
}

// Steps returns a copy of all recorded steps.
func (s *Stepper) Steps() []Step {
	out := make([]Step, len(s.steps))
	copy(out, s.steps)
	return out
}

// HasError reports whether any step recorded a non-nil error.
func (s *Stepper) HasError() bool {
	for _, st := range s.steps {
		if st.Error != nil {
			return true
		}
	}
	return false
}

// Summary returns a human-readable summary of all steps.
func (s *Stepper) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "stepper: service=%s steps=%d\n", s.service, len(s.steps))
	for i, st := range s.steps {
		status := "ok"
		if st.Skipped {
			status = "skipped"
		} else if st.Error != nil {
			status = fmt.Sprintf("error: %s", st.Error)
		}
		fmt.Fprintf(&b, "  [%d] kind=%-10s name=%-20s duration=%s status=%s\n",
			i+1, st.Kind, st.Name, st.Duration.Round(time.Millisecond), status)
	}
	return b.String()
}
