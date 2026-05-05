package drift

import "fmt"

// DiffOption controls how a Diff is computed.
type DiffOption func(*diffConfig)

type diffConfig struct {
	ignoreEnv bool
}

// WithIgnoreEnv skips environment variable comparisons during diff.
func WithIgnoreEnv() DiffOption {
	return func(c *diffConfig) {
		c.ignoreEnv = true
	}
}

// Diff computes a human-readable list of differences between two ServiceSpecs.
// It returns a slice of formatted strings describing each field that differs.
func Diff(declared, live ServiceSpec, opts ...DiffOption) []string {
	cfg := &diffConfig{}
	for _, o := range opts {
		o(cfg)
	}

	var diffs []string

	if declared.Image != live.Image {
		diffs = append(diffs, fmt.Sprintf("image: declared=%q live=%q", declared.Image, live.Image))
	}

	if declared.Replicas != live.Replicas {
		diffs = append(diffs, fmt.Sprintf("replicas: declared=%d live=%d", declared.Replicas, live.Replicas))
	}

	if !cfg.ignoreEnv {
		for k, dv := range declared.Env {
			lv, ok := live.Env[k]
			if !ok {
				diffs = append(diffs, fmt.Sprintf("env.%s: declared=%q live=<missing>", k, dv))
			} else if dv != lv {
				diffs = append(diffs, fmt.Sprintf("env.%s: declared=%q live=%q", k, dv, lv))
			}
		}
		for k := range live.Env {
			if _, ok := declared.Env[k]; !ok {
				diffs = append(diffs, fmt.Sprintf("env.%s: declared=<missing> live=%q", k, live.Env[k]))
			}
		}
	}

	return diffs
}
