package codebase_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/codebase"
	"github.com/raghavendrashivam474/aayam/internal/scanner"
	"github.com/raghavendrashivam474/aayam/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func buildCodeGraph(t *testing.T, root string) codebase.CodeGraph {
	t.Helper()
	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}
	cb := codebase.Discover(root, inv)
	return codebase.BuildCodeGraph(root, cb, inv)
}

// ---------------------------------------------------------------------------
// Empty project
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_Empty(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})
	g := buildCodeGraph(t, root)

	if g.NodeCount() != 0 {
		t.Errorf("want 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("want 0 edges, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// Single file, single package
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_SingleFile(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":              "module example.com/app\n\ngo 1.21\n",
		"internal/app/app.go": "package app\n",
	})
	g := buildCodeGraph(t, root)

	// Expect: file node, dir node, package node.
	found := make(map[string]bool)
	for _, n := range g.Nodes {
		found[n.ID] = true
	}

	if !found["file:internal/app/app.go"] {
		t.Error("want file node")
	}
	if !found["dir:internal/app"] {
		t.Error("want dir node")
	}
	if !found["package:app"] {
		t.Error("want package node")
	}
}

// ---------------------------------------------------------------------------
// Package depends_on edges are present
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_PackageDependsOn(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildCodeGraph(t, root)

	if !hasEdge(g, "package:cli", "package:config", codebase.RelationshipDependsOn) {
		t.Error("want depends_on edge package:cli -> package:config")
	}
}

// ---------------------------------------------------------------------------
// File imports edges are present
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_FileImports(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildCodeGraph(t, root)

	if !hasEdge(g, "file:internal/cli/cli.go", "package:config", codebase.RelationshipImports) {
		t.Error("want imports edge file:internal/cli/cli.go -> package:config")
	}
}

// ---------------------------------------------------------------------------
// Directory contains file edges are present
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_DirContainsFile(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                    "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go": "package config\n",
	})
	g := buildCodeGraph(t, root)

	if !hasEdge(g, "dir:internal/config", "file:internal/config/config.go", codebase.RelationshipContains) {
		t.Error("want contains edge dir:internal/config -> file:internal/config/config.go")
	}
}

// ---------------------------------------------------------------------------
// File belongs_to package edges are present
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_FileBelongsTo(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                    "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go": "package config\n",
	})
	g := buildCodeGraph(t, root)

	if !hasEdge(g, "file:internal/config/config.go", "package:config", codebase.RelationshipBelongsTo) {
		t.Error("want belongs_to edge file:internal/config/config.go -> package:config")
	}
}

// ---------------------------------------------------------------------------
// Nodes are not duplicated when package graph and file graph overlap
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_NoDuplicateNodes(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                    "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go": "package config\n",
	})
	g := buildCodeGraph(t, root)

	seen := make(map[string]int)
	for _, n := range g.Nodes {
		seen[n.ID]++
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("node %q appears %d times (want 1)", id, count)
		}
	}
}

// ---------------------------------------------------------------------------
// Edges are not duplicated
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_NoDuplicateEdges(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildCodeGraph(t, root)

	type edgeKey struct{ from, to, kind string }
	seen := make(map[edgeKey]int)
	for _, e := range g.Edges {
		k := edgeKey{e.From, e.To, string(e.Kind)}
		seen[k]++
	}
	for k, count := range seen {
		if count > 1 {
			t.Errorf("edge %+v appears %d times (want 1)", k, count)
		}
	}
}

// ---------------------------------------------------------------------------
// Determinism
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_Determinism(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})

	g1 := buildCodeGraph(t, root)
	g2 := buildCodeGraph(t, root)

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
// Non-Go project
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_NonGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"index.html": "<html></html>",
		"style.css":  "body {}",
	})
	g := buildCodeGraph(t, root)

	for _, n := range g.Nodes {
		if n.Kind == codebase.NodePackage {
			t.Errorf("want no package nodes for non-Go project, got %+v", n)
		}
	}

	for _, e := range g.Edges {
		if e.Kind == codebase.RelationshipDependsOn || e.Kind == codebase.RelationshipImports {
			t.Errorf("want no depends_on or imports edges for non-Go project, got %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Node IDs never contain backslashes
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_NoWindowsPathsInIDs(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                    "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go": "package config\n",
	})
	g := buildCodeGraph(t, root)

	for _, n := range g.Nodes {
		if strings.Contains(n.ID, "\\") {
			t.Errorf("node ID contains backslash: %q", n.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// JSON serialisation — graph section is present and well-formed
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_JSONSerialisation(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildCodeGraph(t, root)

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok {
		t.Fatal("want nodes array in JSON")
	}
	if len(nodes) == 0 {
		t.Error("want at least one node in JSON")
	}

	edges, ok := result["edges"].([]interface{})
	if !ok {
		t.Fatal("want edges array in JSON")
	}
	if len(edges) == 0 {
		t.Error("want at least one edge in JSON")
	}
}

// ---------------------------------------------------------------------------
// EdgesByKind reports correct counts on unified graph
// ---------------------------------------------------------------------------

func TestBuildCodeGraph_EdgesByKind(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import "example.com/app/internal/config"

var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildCodeGraph(t, root)

	byKind := g.EdgesByKind()

	if byKind[codebase.RelationshipDependsOn] == 0 {
		t.Error("want at least one depends_on edge")
	}
	if byKind[codebase.RelationshipContains] == 0 {
		t.Error("want at least one contains edge")
	}
	if byKind[codebase.RelationshipBelongsTo] == 0 {
		t.Error("want at least one belongs_to edge")
	}
	if byKind[codebase.RelationshipImports] == 0 {
		t.Error("want at least one imports edge")
	}
}
