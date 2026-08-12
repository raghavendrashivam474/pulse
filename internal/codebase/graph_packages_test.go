package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/project"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makePackage builds a minimal Package for test inputs.
func makePackage(name, path string) codebase.Package {
	return codebase.Package{
		Name:     name,
		Path:     path,
		Language: project.LangGo,
		Files:    []string{},
	}
}

// makeDep builds a minimal Dependency for test inputs.
func makeDep(from, to string) codebase.Dependency {
	return codebase.Dependency{From: from, To: to, Type: codebase.DependencyImport}
}

// ---------------------------------------------------------------------------
// Empty codebase
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_Empty(t *testing.T) {
	cb := codebase.Codebase{}
	g := codebase.BuildPackageGraph(cb)

	if g.NodeCount() != 0 {
		t.Errorf("want 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("want 0 edges, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// Single package, no dependencies
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_SinglePackage(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{makePackage("cli", "internal/cli")},
	}
	g := codebase.BuildPackageGraph(cb)

	if g.NodeCount() != 1 {
		t.Fatalf("want 1 node, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("want 0 edges, got %d", g.EdgeCount())
	}

	n := g.Nodes[0]
	if n.ID != "package:cli" {
		t.Errorf("want node ID %q, got %q", "package:cli", n.ID)
	}
	if n.Kind != codebase.NodePackage {
		t.Errorf("want node Kind %q, got %q", codebase.NodePackage, n.Kind)
	}
	if n.Name != "cli" {
		t.Errorf("want node Name %q, got %q", "cli", n.Name)
	}
}

// ---------------------------------------------------------------------------
// Single dependency A → B
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_SingleDependency(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("cli", "internal/cli"),
			makePackage("config", "internal/config"),
		},
		Dependencies: []codebase.Dependency{makeDep("cli", "config")},
	}
	g := codebase.BuildPackageGraph(cb)

	if g.NodeCount() != 2 {
		t.Fatalf("want 2 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("want 1 edge, got %d", g.EdgeCount())
	}

	e := g.Edges[0]
	if e.From != "package:cli" {
		t.Errorf("want edge From %q, got %q", "package:cli", e.From)
	}
	if e.To != "package:config" {
		t.Errorf("want edge To %q, got %q", "package:config", e.To)
	}
	if e.Kind != codebase.RelationshipDependsOn {
		t.Errorf("want edge Kind %q, got %q", codebase.RelationshipDependsOn, e.Kind)
	}
}

// ---------------------------------------------------------------------------
// Multiple dependencies A → B, A → C, A → D
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_MultipleDependencies(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("cli", "internal/cli"),
			makePackage("config", "internal/config"),
			makePackage("errors", "internal/errors"),
			makePackage("output", "internal/output"),
		},
		Dependencies: []codebase.Dependency{
			makeDep("cli", "config"),
			makeDep("cli", "errors"),
			makeDep("cli", "output"),
		},
	}
	g := codebase.BuildPackageGraph(cb)

	if g.NodeCount() != 4 {
		t.Fatalf("want 4 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 3 {
		t.Fatalf("want 3 edges, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// Duplicate dependencies collapse to one edge
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_DuplicateDependencies(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("cli", "internal/cli"),
			makePackage("config", "internal/config"),
		},
		Dependencies: []codebase.Dependency{
			makeDep("cli", "config"),
			makeDep("cli", "config"),
			makeDep("cli", "config"),
		},
	}
	g := codebase.BuildPackageGraph(cb)

	if g.EdgeCount() != 1 {
		t.Errorf("want 1 edge after deduplication, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// Determinism — same input always produces same output
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_Determinism(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("snapshot", "internal/snapshot"),
			makePackage("cli", "internal/cli"),
			makePackage("config", "internal/config"),
			makePackage("output", "internal/output"),
		},
		Dependencies: []codebase.Dependency{
			makeDep("cli", "snapshot"),
			makeDep("cli", "config"),
			makeDep("snapshot", "output"),
		},
	}

	g1 := codebase.BuildPackageGraph(cb)
	g2 := codebase.BuildPackageGraph(cb)

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
		e1, e2 := g1.Edges[i], g2.Edges[i]
		if e1.From != e2.From || e1.To != e2.To || e1.Kind != e2.Kind {
			t.Errorf("edge[%d] mismatch", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Node ordering is alphabetical by ID
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_NodeOrdering(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("snapshot", "internal/snapshot"),
			makePackage("cli", "internal/cli"),
			makePackage("config", "internal/config"),
		},
	}
	g := codebase.BuildPackageGraph(cb)

	want := []string{"package:cli", "package:config", "package:snapshot"}
	if len(g.Nodes) != len(want) {
		t.Fatalf("want %d nodes, got %d", len(want), len(g.Nodes))
	}
	for i, n := range g.Nodes {
		if n.ID != want[i] {
			t.Errorf("node[%d]: want %q, got %q", i, want[i], n.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge ordering is by (From, To)
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_EdgeOrdering(t *testing.T) {
	cb := codebase.Codebase{
		Packages: []codebase.Package{
			makePackage("cli", "internal/cli"),
			makePackage("snapshot", "internal/snapshot"),
			makePackage("config", "internal/config"),
		},
		Dependencies: []codebase.Dependency{
			makeDep("cli", "snapshot"),
			makeDep("cli", "config"),
		},
	}
	g := codebase.BuildPackageGraph(cb)

	// After sort: cli→config comes before cli→snapshot.
	if g.Edges[0].To != "package:config" {
		t.Errorf("want first edge To %q, got %q", "package:config", g.Edges[0].To)
	}
	if g.Edges[1].To != "package:snapshot" {
		t.Errorf("want second edge To %q, got %q", "package:snapshot", g.Edges[1].To)
	}
}

// ---------------------------------------------------------------------------
// Integration — real project via TempProject
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_RealProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/myapp\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/myapp/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	cb := codebase.Discover(root, inv)
	g := codebase.BuildPackageGraph(cb)

	// Should have at least cli and config nodes.
	found := make(map[string]bool)
	for _, n := range g.Nodes {
		found[n.ID] = true
	}

	if !found["package:cli"] {
		t.Error("want node package:cli")
	}
	if !found["package:config"] {
		t.Error("want node package:config")
	}

	// Should have cli → config edge.
	edgeFound := false
	for _, e := range g.Edges {
		if e.From == "package:cli" && e.To == "package:config" && e.Kind == codebase.RelationshipDependsOn {
			edgeFound = true
		}
	}
	if !edgeFound {
		t.Error("want depends_on edge from package:cli to package:config")
	}
}

// ---------------------------------------------------------------------------
// Non-Go project produces empty graph
// ---------------------------------------------------------------------------

func TestBuildPackageGraph_NonGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"index.html": "<html></html>",
		"style.css":  "body {}",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	cb := codebase.Discover(root, inv)
	g := codebase.BuildPackageGraph(cb)

	if g.NodeCount() != 0 {
		t.Errorf("want 0 nodes for non-Go project, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("want 0 edges for non-Go project, got %d", g.EdgeCount())
	}
}
