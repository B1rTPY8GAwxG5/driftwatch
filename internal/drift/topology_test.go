package drift

import (
	"strings"
	"testing"
)

func TestNewTopologyGraph_Empty(t *testing.T) {
	g := NewTopologyGraph()
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Len() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.Len())
	}
}

func TestTopologyGraph_AddNode_EmptyServiceIgnored(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: ""})
	if g.Len() != 0 {
		t.Fatal("expected empty graph after adding blank node")
	}
}

func TestTopologyGraph_AddNode_Valid(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "api", Dependencies: []string{"db"}})
	if g.Len() != 1 {
		t.Fatalf("expected 1 node, got %d", g.Len())
	}
}

func TestTopologyGraph_Dependents_None(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "db"})
	deps := g.Dependents("db")
	if len(deps) != 0 {
		t.Fatalf("expected no dependents, got %v", deps)
	}
}

func TestTopologyGraph_Dependents_Found(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "db"})
	g.AddNode(TopologyNode{Service: "api", Dependencies: []string{"db"}})
	g.AddNode(TopologyNode{Service: "worker", Dependencies: []string{"db"}})
	deps := g.Dependents("db")
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependents, got %v", deps)
	}
	if deps[0] != "api" || deps[1] != "worker" {
		t.Fatalf("unexpected order: %v", deps)
	}
}

func TestTopologyGraph_ImpactedBy_Transitive(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "db"})
	g.AddNode(TopologyNode{Service: "api", Dependencies: []string{"db"}})
	g.AddNode(TopologyNode{Service: "frontend", Dependencies: []string{"api"}})
	impacted := g.ImpactedBy("db")
	if len(impacted) != 2 {
		t.Fatalf("expected 2 impacted services, got %v", impacted)
	}
}

func TestTopologyGraph_ImpactedBy_NoCycles(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "a", Dependencies: []string{"b"}})
	g.AddNode(TopologyNode{Service: "b", Dependencies: []string{"a"}})
	// Should not loop forever
	impacted := g.ImpactedBy("a")
	if len(impacted) == 0 {
		t.Fatal("expected at least one impacted service")
	}
}

func TestTopologyGraph_String_ContainsService(t *testing.T) {
	g := NewTopologyGraph()
	g.AddNode(TopologyNode{Service: "api", Dependencies: []string{"db"}})
	out := g.String()
	if !strings.Contains(out, "api") {
		t.Fatalf("expected 'api' in string output, got: %s", out)
	}
}
