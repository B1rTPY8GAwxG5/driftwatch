package drift

import (
	"testing"
)

var indexerDriftedResult = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		{Kind: KindReplicas, Field: "replicas", Declared: "2", Observed: "3"},
	},
}

var indexerCleanResult = DriftResult{
	Service: "worker",
	Entries: []DriftEntry{},
}

func TestNewIndexer_DefaultsToService(t *testing.T) {
	idx := NewIndexer("unknown")
	if idx.Mode() != "service" {
		t.Errorf("expected service mode, got %s", idx.Mode())
	}
}

func TestNewIndexer_KnownModes(t *testing.T) {
	for _, m := range []IndexMode{IndexByService, IndexByKind, IndexByBoth} {
		idx := NewIndexer(m)
		if idx.Mode() != string(m) {
			t.Errorf("expected %s, got %s", m, idx.Mode())
		}
	}
}

func TestIndexer_Index_ByService_Lookup(t *testing.T) {
	idx := NewIndexer(IndexByService)
	idx.Index(indexerDriftedResult)
	results := idx.Lookup("api")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Service != "api" {
		t.Errorf("unexpected service: %s", results[0].Service)
	}
}

func TestIndexer_Index_ByKind_Lookup(t *testing.T) {
	idx := NewIndexer(IndexByKind)
	idx.Index(indexerDriftedResult)
	results := idx.Lookup(string(KindImage))
	if len(results) != 1 {
		t.Fatalf("expected 1 result for image kind, got %d", len(results))
	}
}

func TestIndexer_Index_ByBoth_LookupService(t *testing.T) {
	idx := NewIndexer(IndexByBoth)
	idx.Index(indexerDriftedResult)
	results := idx.Lookup("api")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for service key, got %d", len(results))
	}
}

func TestIndexer_Index_ByBoth_LookupComposite(t *testing.T) {
	idx := NewIndexer(IndexByBoth)
	idx.Index(indexerDriftedResult)
	key := "api::" + string(KindImage)
	results := idx.Lookup(key)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for composite key, got %d", len(results))
	}
}

func TestIndexer_Lookup_Missing_ReturnsEmpty(t *testing.T) {
	idx := NewIndexer(IndexByService)
	results := idx.Lookup("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d results", len(results))
	}
}

func TestIndexer_Keys_Sorted(t *testing.T) {
	idx := NewIndexer(IndexByService)
	idx.Index(indexerDriftedResult)
	idx.Index(indexerCleanResult)
	keys := idx.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] > keys[1] {
		t.Errorf("keys not sorted: %v", keys)
	}
}

func TestIndexer_Flush_ClearsAll(t *testing.T) {
	idx := NewIndexer(IndexByService)
	idx.Index(indexerDriftedResult)
	idx.Flush()
	if len(idx.Keys()) != 0 {
		t.Errorf("expected empty index after flush")
	}
}

func TestIndexer_CleanResult_ByKind_FallsToNone(t *testing.T) {
	idx := NewIndexer(IndexByKind)
	idx.Index(indexerCleanResult)
	results := idx.Lookup("none")
	if len(results) != 1 {
		t.Fatalf("expected 1 result under 'none' key, got %d", len(results))
	}
}
