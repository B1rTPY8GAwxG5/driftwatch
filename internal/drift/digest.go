package drift

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// DigestOptions controls which fields are included in the digest.
type DigestOptions struct {
	IncludeEnv      bool
	IncludeReplicas bool
	IncludeImage    bool
}

// DefaultDigestOptions returns options with all fields enabled.
func DefaultDigestOptions() DigestOptions {
	return DigestOptions{
		IncludeEnv:      true,
		IncludeReplicas: true,
		IncludeImage:    true,
	}
}

// Digester computes a stable hash digest from a DriftResult.
type Digester struct {
	opts DigestOptions
}

// NewDigester creates a Digester with the provided options.
func NewDigester(opts DigestOptions) *Digester {
	return &Digester{opts: opts}
}

// Compute returns a hex-encoded SHA-256 digest for the given DriftResult.
// The digest is stable across calls for equivalent results.
func (d *Digester) Compute(result DriftResult) string {
	var parts []string

	parts = append(parts, "service:"+result.Service)

	if d.opts.IncludeImage {
		for _, e := range result.Entries {
			if e.Kind == KindImage {
				parts = append(parts, fmt.Sprintf("image:%s->%s", e.Expected, e.Actual))
			}
		}
	}

	if d.opts.IncludeReplicas {
		for _, e := range result.Entries {
			if e.Kind == KindReplicas {
				parts = append(parts, fmt.Sprintf("replicas:%s->%s", e.Expected, e.Actual))
			}
		}
	}

	if d.opts.IncludeEnv {
		var envParts []string
		for _, e := range result.Entries {
			if e.Kind == KindEnv {
				envParts = append(envParts, fmt.Sprintf("env:%s:%s->%s", e.Field, e.Expected, e.Actual))
			}
		}
		sort.Strings(envParts)
		parts = append(parts, envParts...)
	}

	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h)
}
