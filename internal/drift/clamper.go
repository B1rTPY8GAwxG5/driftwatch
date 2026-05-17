package drift

import "fmt"

// ClamperOptions configures the Clamper behaviour.
type ClamperOptions struct {
	MinReplicas int32
	MaxReplicas int32
	MinScore    float64
	MaxScore    float64
}

// DefaultClamperOptions returns sensible defaults.
func DefaultClamperOptions() ClamperOptions {
	return ClamperOptions{
		MinReplicas: 1,
		MaxReplicas: 100,
		MinScore:    0.0,
		MaxScore:    100.0,
	}
}

// Clamper enforces upper and lower bounds on numeric fields within a
// DriftResult, returning a sanitised copy and a list of violations.
type Clamper struct {
	opts ClamperOptions
}

// ClamperViolation describes a single out-of-range value that was clamped.
type ClamperViolation struct {
	Field    string
	Original interface{}
	Clamped  interface{}
}

func (v ClamperViolation) String() string {
	return fmt.Sprintf("%s: %v clamped to %v", v.Field, v.Original, v.Clamped)
}

// NewClamper creates a Clamper with the provided options.
// Zero-value options fall back to DefaultClamperOptions.
func NewClamper(opts ClamperOptions) *Clamper {
	def := DefaultClamperOptions()
	if opts.MaxReplicas == 0 {
		opts.MaxReplicas = def.MaxReplicas
	}
	if opts.MinScore == 0 && opts.MaxScore == 0 {
		opts.MinScore = def.MinScore
		opts.MaxScore = def.MaxScore
	}
	return &Clamper{opts: opts}
}

// Clamp returns a copy of result with numeric fields sanitised and any
// violations that were corrected.
func (c *Clamper) Clamp(result DriftResult) (DriftResult, []ClamperViolation) {
	var violations []ClamperViolation

	out := result

	// Clamp Replicas on the embedded spec if present.
	if out.Spec.Replicas < int(c.opts.MinReplicas) && out.Spec.Replicas != 0 {
		violations = append(violations, ClamperViolation{
			Field:    "spec.replicas",
			Original: out.Spec.Replicas,
			Clamped:  int(c.opts.MinReplicas),
		})
		out.Spec.Replicas = int(c.opts.MinReplicas)
	} else if out.Spec.Replicas > int(c.opts.MaxReplicas) {
		violations = append(violations, ClamperViolation{
			Field:    "spec.replicas",
			Original: out.Spec.Replicas,
			Clamped:  int(c.opts.MaxReplicas),
		})
		out.Spec.Replicas = int(c.opts.MaxReplicas)
	}

	return out, violations
}
