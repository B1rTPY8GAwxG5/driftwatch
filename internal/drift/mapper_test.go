package drift

import (
	"testing"
)

var mapperCleanResult = DriftResult{Service: "alpha", Entries: nil}
var mapperDriftedImage = DriftResult{
	Service: "beta",
	Entries: []DriftEntry{{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"}},
}
var mapperDriftedReplicas = DriftResult{
	Service: "gamma",
	Entries: []DriftEntry{{Kind: KindReplicas, Field: "replicas", Declared: "2", Observed: "3"}},
}

func TestNewMapper_DefaultsToService(t *testing.T) {
	m := NewMapper("unknown")
	if m.Mode() != MapByService {
		t.Fatalf("expected MapByService, got %s", m.Mode())
	}
}

func TestNewMapper_KnownModes(t *testing.T) {
	for _, mode := range []MapMode{MapByService, MapByKind} {
		m := NewMapper(mode)
		if m.Mode() != mode {
			t.Fatalf("expected %s, got %s", mode, m.Mode())
		}
	}
}

func TestMapper_Map_ByService(t *testing.T) {
	m := NewMapper(MapByService)
	results := []DriftResult{mapperCleanResult, mapperDriftedImage}
	out := m.Map(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
	if _, ok := out["alpha"]; !ok {
		t.Error("expected key 'alpha'")
	}
	if _, ok := out["beta"]; !ok {
		t.Error("expected key 'beta'")
	}
}

func TestMapper_Map_ByKind_NoEntries_FallsToNone(t *testing.T) {
	m := NewMapper(MapByKind)
	out := m.Map([]DriftResult{mapperCleanResult})
	if _, ok := out["none"]; !ok {
		t.Error("expected 'none' bucket for clean result")
	}
}

func TestMapper_Map_ByKind_GroupsCorrectly(t *testing.T) {
	m := NewMapper(MapByKind)
	out := m.Map([]DriftResult{mapperDriftedImage, mapperDriftedReplicas})
	if len(out[string(KindImage)]) != 1 {
		t.Errorf("expected 1 image result, got %d", len(out[string(KindImage)]))
	}
	if len(out[string(KindReplicas)]) != 1 {
		t.Errorf("expected 1 replicas result, got %d", len(out[string(KindReplicas)]))
	}
}

func TestMapper_Lookup_Found(t *testing.T) {
	m := NewMapper(MapByService)
	results := []DriftResult{mapperDriftedImage}
	got := m.Lookup(results, "beta")
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
}

func TestMapper_Lookup_NotFound(t *testing.T) {
	m := NewMapper(MapByService)
	got := m.Lookup([]DriftResult{mapperCleanResult}, "missing")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
