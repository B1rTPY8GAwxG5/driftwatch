package drift

import (
	"fmt"
	"sort"
	"strings"
)

// TopologyNode represents a single service node in the dependency graph.
type TopologyNode struct {
	Service      string
	Dependencies []string
	Labels       map[string]string
}

// TopologyGraph holds the full service dependency topology.
type TopologyGraph struct {
	nodes map[string]*TopologyNode
}

// NewTopologyGraph creates an empty TopologyGraph.
func NewTopologyGraph() *TopologyGraph {
	return &TopologyGraph{nodes: make(map[string]*TopologyNode)}
}

// AddNode registers a node in the graph. Duplicate service names are overwritten.
func (g *TopologyGraph) AddNode(node TopologyNode) {
	if node.Service == "" {
		return
	}
	g.nodes[node.Service] = &node
}

// Dependents returns services that directly depend on the given service.
func (g *TopologyGraph) Dependents(service string) []string {
	var deps []string
	for name, node := range g.nodes {
		for _, d := range node.Dependencies {
			if d == service {
				deps = append(deps, name)
			}
		}
	}
	sort.Strings(deps)
	return deps
}

// ImpactedBy returns all services transitively impacted by drift in the given service.
func (g *TopologyGraph) ImpactedBy(service string) []string {
	visited := make(map[string]bool)
	g.traverse(service, visited)
	delete(visited, service)
	result := make([]string, 0, len(visited))
	for s := range visited {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

func (g *TopologyGraph) traverse(service string, visited map[string]bool) {
	if visited[service] {
		return
	}
	visited[service] = true
	for _, dep := range g.Dependents(service) {
		g.traverse(dep, visited)
	}
}

// Len returns the number of nodes in the graph.
func (g *TopologyGraph) Len() int { return len(g.nodes) }

// String returns a human-readable adjacency summary.
func (g *TopologyGraph) String() string {
	names := make([]string, 0, len(g.nodes))
	for n := range g.nodes {
		names = append(names, n)
	}
	sort.Strings(names)
	var sb strings.Builder
	for _, n := range names {
		node := g.nodes[n]
		fmt.Fprintf(&sb, "%s -> [%s]\n", n, strings.Join(node.Dependencies, ", "))
	}
	return sb.String()
}
