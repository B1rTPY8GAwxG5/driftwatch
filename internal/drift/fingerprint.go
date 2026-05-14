package drift

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// Fingerprint is a stable hash representing the state of a DriftResult.
type Fingerprint string

// FingerprintOptions controls which fields contribute to the fingerprint.
type FingerprintOptions struct {
	IncludeEnv      bool
	IncludeReplicas bool
	IncludeImage    bool
}

// DefaultFingerprintOptions returns options that include all fields.
func DefaultFingerprintOptions() FingerprintOptions {
	return FingerprintOptions{
		IncludeEnv:      true,
		IncludeReplicas: true,
		IncludeImage:    true,
	}
}

// Fingerprinter computes stable fingerprints for drift results.
type Fingerprinter struct {
	opts FingerprintOptions
}

// NewFingerprinter creates a Fingerprinter with the given options.
func NewFingerprinter(opts FingerprintOptions) *Fingerprinter {
	return &Fingerprinter{opts: opts}
}

// Compute returns a stable Fingerprint for the given DriftResult.
func (f *Fingerprinter) Compute(r DriftResult) Fingerprint {
	var parts []string
	parts = append(parts, r.Service)

	for _, e := range r.Entries {
		switch e.Kind {
		case KindImage:
			if f.opts.IncludeImage {
				parts = append(parts, fmt.Sprintf("image:%s->%s", e.Declared, e.Detected))
			}
		case KindReplicas:
			if f.opts.IncludeReplicas {
				parts = append(parts, fmt.Sprintf("replicas:%s->%s", e.Declared, e.Detected))
			}
		case KindEnv:
			if f.opts.IncludeEnv {
				parts = append(parts, fmt.Sprintf("env:%s:%s->%s", e.Field, e.Declared, e.Detected))
			}
		}
	}

	sort.Strings(parts[1:]) // stable order for entries, preserve service at index 0
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return Fingerprint(fmt.Sprintf("%x", h[:8]))
}

// Equal returns true if two fingerprints match.
func (fp Fingerprint) Equal(other Fingerprint) bool {
	return fp == other
}

// String returns the string representation of the fingerprint.
func (fp Fingerprint) String() string {
	return string(fp)
}
