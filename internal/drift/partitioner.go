package drift

import "sort"

// PartitionMode determines how results are partitioned.
type PartitionMode string

const (
	PartitionByDrifted PartitionMode = "drifted"
	PartitionByClean   PartitionMode = "clean"
	PartitionByKind    PartitionMode = "kind"
)

// Partition holds a named collection of drift results.
type Partition struct {
	Name    string
	Results []DriftResult
}

// Partitioner splits drift results into named partitions.
type Partitioner struct {
	mode PartitionMode
}

// NewPartitioner returns a Partitioner using the given mode.
// Defaults to PartitionByDrifted if the mode is unrecognised.
func NewPartitioner(mode PartitionMode) *Partitioner {
	switch mode {
	case PartitionByDrifted, PartitionByClean, PartitionByKind:
	default:
		mode = PartitionByDrifted
	}
	return &Partitioner{mode: mode}
}

// Mode returns the active partition mode.
func (p *Partitioner) Mode() PartitionMode { return p.mode }

// Partition divides results into labelled buckets according to the mode.
func (p *Partitioner) Partition(results []DriftResult) []Partition {
	switch p.mode {
	case PartitionByKind:
		return p.byKind(results)
	case PartitionByClean:
		return p.byClean(results)
	default:
		return p.byDrifted(results)
	}
}

func (p *Partitioner) byDrifted(results []DriftResult) []Partition {
	var drifted, clean []DriftResult
	for _, r := range results {
		if r.HasDrift() {
			drifted = append(drifted, r)
		} else {
			clean = append(clean, r)
		}
	}
	return []Partition{
		{Name: "drifted", Results: drifted},
		{Name: "clean", Results: clean},
	}
}

func (p *Partitioner) byClean(results []DriftResult) []Partition {
	parts := p.byDrifted(results)
	// reverse order: clean first
	parts[0], parts[1] = parts[1], parts[0]
	return parts
}

func (p *Partitioner) byKind(results []DriftResult) []Partition {
	buckets := map[string][]DriftResult{}
	for _, r := range results {
		if !r.HasDrift() {
			buckets["clean"] = append(buckets["clean"], r)
			continue
		}
		for _, e := range r.Entries {
			k := string(e.Kind)
			buckets[k] = append(buckets[k], r)
			break // one entry per result is sufficient for classification
		}
	}
	names := make([]string, 0, len(buckets))
	for k := range buckets {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]Partition, 0, len(names))
	for _, n := range names {
		out = append(out, Partition{Name: n, Results: buckets[n]})
	}
	return out
}
