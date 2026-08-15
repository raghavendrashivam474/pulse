package codebase

import "github.com/raghavendrashivam474/aayam/internal/scanner"

// BuildCodeGraph constructs the unified CodeGraph for a project.
//
// It combines:
//   - Package nodes and depends_on edges  (from BuildPackageGraph)
//   - File nodes, directory nodes, contains edges,
//     belongs_to edges, and imports edges  (from BuildFileGraph)
//
// The two sub-graphs are merged into a single CodeGraph.
// Duplicate nodes and edges that appear in both sub-graphs are
// deduplicated automatically by AddNode / AddEdge.
//
// The returned graph is normalised (deterministic ordering).
//
// This is the single entry point for graph construction.
// Callers should not call BuildPackageGraph or BuildFileGraph directly
// unless they need an isolated sub-graph for testing.
func BuildCodeGraph(root string, cb Codebase, inv scanner.Inventory) CodeGraph {
	g := NewCodeGraph()

	// Merge package graph.
	pkg := BuildPackageGraph(cb)
	for _, n := range pkg.Nodes {
		g.AddNode(n)
	}
	for _, e := range pkg.Edges {
		g.AddEdge(e)
	}

	// Merge file graph.
	file := BuildFileGraph(root, cb, inv)
	for _, n := range file.Nodes {
		g.AddNode(n)
	}
	for _, e := range file.Edges {
		g.AddEdge(e)
	}

	g.Normalise()
	return g
}
