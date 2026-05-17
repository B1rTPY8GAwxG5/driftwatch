package drift

import (
	"testing"
)

func makePartResult(service string, drifted bool) DriftResult {
	r := DriftResult{Service: service}
	if drifted {
		r.Entries = []DriftEntry{{Kind: KindImage, Field: "image", Got: "a", Want: "b"}}
	}
	return r
}

func TestNewPartitioner_DefaultsOnUnknown(t *testing.T) {
	p := NewPartitioner("unknown")
	if p.Mode() != PartitionByDrifted {
		t.Errorf("expected %q, got %q", PartitionByDrifted, p.Mode())
	}
}

func TestNewPartitioner_KnownModes(t *testing.T) {
	for _, m := range []PartitionMode{PartitionByDrifted, PartitionByClean, PartitionByKind} {
		if NewPartitioner(m).Mode() != m {
			t.Errorf("expected mode %q to be preserved", m)
		}
	}
}

func TestPartitioner_ByDrifted_TwoBuckets(t *testing.T) {
	p := NewPartitioner(PartitionByDrifted)
	results := []DriftResult{
		makePartResult("svc-a", true),
		makePartResult("svc-b", false),
		makePartResult("svc-c", true),
	}
	parts := p.Partition(results)
	if len(parts) != 2 {
		t.Fatalf("expected 2 partitions, got %d", len(parts))
	}
	if parts[0].Name != "drifted" {
		t.Errorf("expected first partition name 'drifted', got %q", parts[0].Name)
	}
	if len(parts[0].Results) != 2 {
		t.Errorf("expected 2 drifted results, got %d", len(parts[0].Results))
	}
	if len(parts[1].Results) != 1 {
		t.Errorf("expected 1 clean result, got %d", len(parts[1].Results))
	}
}

func TestPartitioner_ByClean_CleanFirst(t *testing.T) {
	p := NewPartitioner(PartitionByClean)
	results := []DriftResult{
		makePartResult("svc-a", true),
		makePartResult("svc-b", false),
	}
	parts := p.Partition(results)
	if parts[0].Name != "clean" {
		t.Errorf("expected first partition 'clean', got %q", parts[0].Name)
	}
}

func TestPartitioner_ByKind_SortedBuckets(t *testing.T) {
	p := NewPartitioner(PartitionByKind)
	r1 := DriftResult{Service: "svc-a", Entries: []DriftEntry{{Kind: KindImage}}}
	r2 := DriftResult{Service: "svc-b", Entries: []DriftEntry{{Kind: KindReplicas}}}
	r3 := makePartResult("svc-c", false)
	parts := p.Partition([]DriftResult{r1, r2, r3})
	names := make([]string, len(parts))
	for i, pt := range parts {
		names[i] = pt.Name
	}
	// names must be sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("partitions not sorted: %v", names)
		}
	}
}

func TestPartitioner_ByKind_EmptyInput(t *testing.T) {
	p := NewPartitioner(PartitionByKind)
	parts := p.Partition(nil)
	if len(parts) != 0 {
		t.Errorf("expected 0 partitions for empty input, got %d", len(parts))
	}
}

func TestPartitioner_ByDrifted_AllClean(t *testing.T) {
	p := NewPartitioner(PartitionByDrifted)
	parts := p.Partition([]DriftResult{
		makePartResult("svc-a", false),
		makePartResult("svc-b", false),
	})
	if len(parts[0].Results) != 0 {
		t.Errorf("expected 0 drifted, got %d", len(parts[0].Results))
	}
	if len(parts[1].Results) != 2 {
		t.Errorf("expected 2 clean, got %d", len(parts[1].Results))
	}
}
