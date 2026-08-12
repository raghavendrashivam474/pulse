package codebase_test

import (
	"encoding/json"
	"testing"

	"pulse/internal/codebase"
)

// ---------------------------------------------------------------------------
// Node construction
// ---------------------------------------------------------------------------

func TestNode_Fields(t *testing.T) {
	n := codebase.Node{
		ID:   "package:cli",
		Kind: codebase.NodePackage,
		Name: "cli",
	}

	if n.ID != "package:cli" {
		t.Errorf("want ID %q, got %q", "package:cli", n.ID)
	}
	if n.Kind != codebase.NodePackage {
		t.Errorf("want Kind %q, got %q", codebase.NodePackage, n.Kind)
	}
	if n.Name != "cli" {
		t.Errorf("want Name %q, got %q", "cli", n.Name)
	}
}

// ---------------------------------------------------------------------------
// Relationship construction
// ---------------------------------------------------------------------------

func TestRelationship_Fields(t *testing.T) {
	r := codebase.Relationship{
		From: "package:cli",
		To:   "package:snapshot",
		Kind: codebase.RelationshipDependsOn,
	}

	if r.From != "package:cli" {
		t.Errorf("want From %q, got %q", "package:cli", r.From)
	}
	if r.To != "package:snapshot" {
		t.Errorf("want To %q, got %q", "package:snapshot", r.To)
	}
	if r.Kind != codebase.RelationshipDependsOn {
		t.Errorf("want Kind %q, got %q", codebase.RelationshipDependsOn, r.Kind)
	}
}

// ---------------------------------------------------------------------------
// CodeGraph — AddNode deduplication
// ---------------------------------------------------------------------------

func TestCodeGraph_AddNode_Deduplication(t *testing.T) {
	g := codebase.NewCodeGraph()

	n := codebase.Node{ID: "package:cli", Kind: codebase.NodePackage, Name: "cli"}
	g.AddNode(n)
	g.AddNode(n)
	g.AddNode(n)

	if g.NodeCount() != 1 {
		t.Errorf("want 1 node after deduplication, got %d", g.NodeCount())
	}
}

// ---------------------------------------------------------------------------
// CodeGraph — AddEdge deduplication
// ---------------------------------------------------------------------------

func TestCodeGraph_AddEdge_Deduplication(t *testing.T) {
	g := codebase.NewCodeGraph()

	r := codebase.Relationship{
		From: "package:cli",
		To:   "package:snapshot",
		Kind: codebase.RelationshipDependsOn,
	}
	g.AddEdge(r)
	g.AddEdge(r)
	g.AddEdge(r)

	if g.EdgeCount() != 1 {
		t.Errorf("want 1 edge after deduplication, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// CodeGraph — Normalise ordering
// ---------------------------------------------------------------------------

func TestCodeGraph_Normalise_NodeOrder(t *testing.T) {
	g := codebase.NewCodeGraph()

	g.AddNode(codebase.Node{ID: "package:snapshot", Kind: codebase.NodePackage, Name: "snapshot"})
	g.AddNode(codebase.Node{ID: "package:cli", Kind: codebase.NodePackage, Name: "cli"})
	g.AddNode(codebase.Node{ID: "package:config", Kind: codebase.NodePackage, Name: "config"})

	g.Normalise()

	want := []string{"package:cli", "package:config", "package:snapshot"}
	for i, n := range g.Nodes {
		if n.ID != want[i] {
			t.Errorf("node[%d]: want ID %q, got %q", i, want[i], n.ID)
		}
	}
}

func TestCodeGraph_Normalise_EdgeOrder(t *testing.T) {
	g := codebase.NewCodeGraph()

	g.AddEdge(codebase.Relationship{From: "package:snapshot", To: "package:git", Kind: codebase.RelationshipDependsOn})
	g.AddEdge(codebase.Relationship{From: "package:cli", To: "package:snapshot", Kind: codebase.RelationshipDependsOn})
	g.AddEdge(codebase.Relationship{From: "package:cli", To: "package:config", Kind: codebase.RelationshipDependsOn})

	g.Normalise()

	type edge struct{ from, to string }
	want := []edge{
		{"package:cli", "package:config"},
		{"package:cli", "package:snapshot"},
		{"package:snapshot", "package:git"},
	}

	for i, e := range g.Edges {
		if e.From != want[i].from || e.To != want[i].to {
			t.Errorf("edge[%d]: want (%q→%q), got (%q→%q)",
				i, want[i].from, want[i].to, e.From, e.To)
		}
	}
}

// ---------------------------------------------------------------------------
// CodeGraph — Determinism across multiple runs
// ---------------------------------------------------------------------------

func TestCodeGraph_Determinism(t *testing.T) {
	build := func() codebase.CodeGraph {
		g := codebase.NewCodeGraph()
		g.AddNode(codebase.Node{ID: "package:snapshot", Kind: codebase.NodePackage, Name: "snapshot"})
		g.AddNode(codebase.Node{ID: "package:cli", Kind: codebase.NodePackage, Name: "cli"})
		g.AddEdge(codebase.Relationship{From: "package:cli", To: "package:snapshot", Kind: codebase.RelationshipDependsOn})
		g.AddEdge(codebase.Relationship{From: "package:snapshot", To: "package:cli", Kind: codebase.RelationshipDependsOn})
		g.Normalise()
		return g
	}

	g1 := build()
	g2 := build()

	if len(g1.Nodes) != len(g2.Nodes) {
		t.Fatalf("node count mismatch: %d vs %d", len(g1.Nodes), len(g2.Nodes))
	}
	for i := range g1.Nodes {
		if g1.Nodes[i].ID != g2.Nodes[i].ID {
			t.Errorf("node[%d] mismatch: %q vs %q", i, g1.Nodes[i].ID, g2.Nodes[i].ID)
		}
	}

	if len(g1.Edges) != len(g2.Edges) {
		t.Fatalf("edge count mismatch: %d vs %d", len(g1.Edges), len(g2.Edges))
	}
	for i := range g1.Edges {
		if g1.Edges[i].From != g2.Edges[i].From || g1.Edges[i].To != g2.Edges[i].To {
			t.Errorf("edge[%d] mismatch", i)
		}
	}
}

// ---------------------------------------------------------------------------
// CodeGraph — JSON serialisation (acceptance criteria from brief)
// ---------------------------------------------------------------------------

func TestCodeGraph_JSON_Serialisation(t *testing.T) {
	g := codebase.NewCodeGraph()

	g.AddNode(codebase.Node{ID: "package:cli", Kind: codebase.NodePackage, Name: "cli"})
	g.AddNode(codebase.Node{ID: "package:snapshot", Kind: codebase.NodePackage, Name: "snapshot"})
	g.AddEdge(codebase.Relationship{
		From: "package:cli",
		To:   "package:snapshot",
		Kind: codebase.RelationshipDependsOn,
	})
	g.Normalise()

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok || len(nodes) != 2 {
		t.Fatalf("want 2 nodes in JSON, got %v", result["nodes"])
	}

	edges, ok := result["edges"].([]interface{})
	if !ok || len(edges) != 1 {
		t.Fatalf("want 1 edge in JSON, got %v", result["edges"])
	}

	// Verify first node after sort is cli.
	firstNode := nodes[0].(map[string]interface{})
	if firstNode["id"] != "package:cli" {
		t.Errorf("want first node id %q, got %q", "package:cli", firstNode["id"])
	}
	if firstNode["kind"] != "package" {
		t.Errorf("want first node kind %q, got %q", "package", firstNode["kind"])
	}

	// Verify the edge.
	firstEdge := edges[0].(map[string]interface{})
	if firstEdge["from"] != "package:cli" {
		t.Errorf("want edge from %q, got %q", "package:cli", firstEdge["from"])
	}
	if firstEdge["to"] != "package:snapshot" {
		t.Errorf("want edge to %q, got %q", "package:snapshot", firstEdge["to"])
	}
	if firstEdge["kind"] != "depends_on" {
		t.Errorf("want edge kind %q, got %q", "depends_on", firstEdge["kind"])
	}
}

// ---------------------------------------------------------------------------
// EdgesByKind
// ---------------------------------------------------------------------------

func TestCodeGraph_EdgesByKind(t *testing.T) {
	g := codebase.NewCodeGraph()

	g.AddEdge(codebase.Relationship{From: "package:cli", To: "package:config", Kind: codebase.RelationshipDependsOn})
	g.AddEdge(codebase.Relationship{From: "package:cli", To: "package:errors", Kind: codebase.RelationshipDependsOn})
	g.AddEdge(codebase.Relationship{From: "dir:internal", To: "dir:internal/cli", Kind: codebase.RelationshipContains})

	counts := g.EdgesByKind()

	if counts[codebase.RelationshipDependsOn] != 2 {
		t.Errorf("want 2 depends_on edges, got %d", counts[codebase.RelationshipDependsOn])
	}
	if counts[codebase.RelationshipContains] != 1 {
		t.Errorf("want 1 contains edge, got %d", counts[codebase.RelationshipContains])
	}
}

// ---------------------------------------------------------------------------
// NewCodeGraph — empty graph has non-nil slices
// ---------------------------------------------------------------------------

func TestNewCodeGraph_NonNilSlices(t *testing.T) {
	g := codebase.NewCodeGraph()

	if g.Nodes == nil {
		t.Error("want non-nil Nodes slice")
	}
	if g.Edges == nil {
		t.Error("want non-nil Edges slice")
	}
}
