package drift

import (
	"strings"
	"unicode"
)

// NormalizerOptions controls which fields are normalised during comparison.
type NormalizerOptions struct {
	LowercaseImage  bool
	TrimEnvValues   bool
	CollapseReplicas bool
}

// DefaultNormalizerOptions returns sensible defaults.
func DefaultNormalizerOptions() NormalizerOptions {
	return NormalizerOptions{
		LowercaseImage:  true,
		TrimEnvValues:   true,
		CollapseReplicas: false,
	}
}

// Normalizer applies field-level normalisation to a ServiceSpec before
// comparison so that cosmetic differences do not produce false drift.
type Normalizer struct {
	opts NormalizerOptions
}

// NewNormalizer constructs a Normalizer with the supplied options.
func NewNormalizer(opts NormalizerOptions) *Normalizer {
	return &Normalizer{opts: opts}
}

// Normalize returns a copy of spec with normalised field values.
func (n *Normalizer) Normalize(spec ServiceSpec) ServiceSpec {
	out := spec

	if n.opts.LowercaseImage {
		out.Image = strings.ToLower(strings.TrimSpace(out.Image))
	}

	if n.opts.TrimEnvValues {
		normEnv := make(map[string]string, len(out.Env))
		for k, v := range out.Env {
			normEnv[k] = strings.TrimFunc(v, unicode.IsSpace)
		}
		out.Env = normEnv
	}

	if n.opts.CollapseReplicas && out.Replicas < 1 {
		out.Replicas = 1
	}

	return out
}

// NormalizeAll applies Normalize to each spec in the slice.
func (n *Normalizer) NormalizeAll(specs []ServiceSpec) []ServiceSpec {
	out := make([]ServiceSpec, len(specs))
	for i, s := range specs {
		out[i] = n.Normalize(s)
	}
	return out
}
