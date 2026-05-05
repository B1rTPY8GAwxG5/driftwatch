package drift

import (
	"context"
	"fmt"
	"io"
)

// Runner executes scheduled drift checks against a set of service specs.
type Runner struct {
	detector *Detector
	schedule Schedule
	specs    []ServiceSpec
	out      io.Writer
}

// NewRunner creates a Runner with the given detector, schedule, specs and output writer.
func NewRunner(d *Detector, s Schedule, specs []ServiceSpec, out io.Writer) (*Runner, error) {
	if d == nil {
		return nil, fmt.Errorf("detector must not be nil")
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("output writer must not be nil")
	}
	return &Runner{detector: d, schedule: s, specs: specs, out: out}, nil
}

// RunOnce performs a single drift check across all specs and writes results.
func (r *Runner) RunOnce(ctx context.Context, liveServiceFn func(string) (*ServiceSpec, error)) error {
	for _, spec := range r.specs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		live, err := liveServiceFn(spec.Name)
		if err != nil {
			fmt.Fprintf(r.out, "error fetching live spec for %s: %v\n", spec.Name, err)
			continue
		}
		result := r.detector.Compare(spec, *live)
		report := BuildReport(result)
		if err := report.WriteTo(r.out); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
	}
	return nil
}

// Start runs drift checks on the configured schedule until ctx is cancelled or MaxRuns is reached.
func (r *Runner) Start(ctx context.Context, liveServiceFn func(string) (*ServiceSpec, error)) error {
	runs := 0
	for {
		if err := r.RunOnce(ctx, liveServiceFn); err != nil {
			return err
		}
		runs++
		if r.schedule.IsLimited() && runs >= r.schedule.MaxRuns {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.schedule.NextTick():
		}
	}
}
