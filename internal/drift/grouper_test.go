package drift

import (
	"testing"
)

var groupImageEntry = DriftEntry{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Actual: "nginx:1.25"}
var groupReplicasEntry = DriftEntry{Kind: KindReplicas, Field: "replicas", Declared: "2", Actual: "3"}

func TestNewGrouper_DefaultsToService(t *testing.T) {
	g := NewGrouper("unknown-mode")
	if g.Mode() != GroupByService {
		t.Errorf("expected GroupByService, got %q", g.Mode())
	}
}

func TestNewGrouper_KnownModes(t *testing.T) {
	for _, mode := range []GroupMode{GroupByService, GroupByKind, GroupBySeverity} {
		g := NewGrouper(mode)
		if g.Mode() != mode {
			t.Errorf("expected %q, got %q", mode, g.Mode())
		}
	}
}

func TestGrouper_Group_ByService(t *testing.T) {
	g := NewGrouper(GroupByService)
	results := []DriftResult{
		{Service: "alpha", Entries: []DriftEntry{groupImageEntry}},
		{Service: "beta", Entries: []DriftEntry{groupReplicasEntry}},
		{Service: "alpha", Entries: []DriftEntry{groupReplicasEntry}},
	}
	groups := g.Group(results)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Key != "alpha" {
		t.Errorf("expected first key alpha, got %q", groups[0].Key)
	}
	if len(groups[0].Results) != 2 {
		t.Errorf("expected 2 results for alpha, got %d", len(groups[0].Results))
	}
}

func TestGrouper_Group_ByKind(t *testing.T) {
	g := NewGrouper(GroupByKind)
	results := []DriftResult{
		{Service: "svc", Entries: []DriftEntry{groupImageEntry}},
		{Service: "svc2", Entries: []DriftEntry{groupReplicasEntry}},
	}
	groups := g.Group(results)
	if len(groups) != 2 {
		t.Fatalf("expected 2 kind groups, got %d", len(groups))
	}
}

func TestGrouper_Group_ByKind_NoEntries_FallsToNone(t *testing.T) {
	g := NewGrouper(GroupByKind)
	results := []DriftResult{{Service: "empty", Entries: nil}}
	groups := g.Group(results)
	if len(groups) != 1 || groups[0].Key != "none" {
		t.Errorf("expected single 'none' group, got %+v", groups)
	}
}

func TestGrouper_Group_BySeverity(t *testing.T) {
	g := NewGrouper(GroupBySeverity)
	results := []DriftResult{
		{Service: "a", Entries: []DriftEntry{groupImageEntry}},
		{Service: "b", Entries: nil},
	}
	groups := g.Group(results)
	if len(groups) == 0 {
		t.Fatal("expected at least one severity group")
	}
}

func TestGrouper_Group_EmptyInput(t *testing.T) {
	g := NewGrouper(GroupByService)
	groups := g.Group(nil)
	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
}

func TestGrouper_Group_Sorted(t *testing.T) {
	g := NewGrouper(GroupByService)
	results := []DriftResult{
		{Service: "zeta"}, {Service: "alpha"}, {Service: "mango"},
	}
	groups := g.Group(results)
	keys := []string{groups[0].Key, groups[1].Key, groups[2].Key}
	if keys[0] > keys[1] || keys[1] > keys[2] {
		t.Errorf("groups not sorted: %v", keys)
	}
}
