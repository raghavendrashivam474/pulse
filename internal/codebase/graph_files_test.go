package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildFileGraph is a convenience wrapper: scan root, discover codebase,
// build file graph.
func buildFileGraph(t *testing.T, root string) codebase.CodeGraph {
	t.Helper()
	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}
	cb := codebase.Discover(root, inv)
	return codebase.BuildFileGraph(root, cb, inv)
}

// findNode returns the node with the given ID, or nil.
func findNode(g codebase.CodeGraph, id string) *codebase.Node {
	for i, n := range g.Nodes {
		if n.ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// hasEdge reports whether the graph contains a specific directed edge.
func hasEdge(g codebase.CodeGraph, from, to string, kind codebase.RelationshipKind) bool {
	for _, e := range g.Edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Empty project
// ---------------------------------------------------------------------------

func TestBuildFileGraph_EmptyProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})
	g := buildFileGraph(t, root)

	if g.NodeCount() != 0 {
		t.Errorf("want 0 nodes for empty project, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("want 0 edges for empty project, got %d", g.EdgeCount())
	}
}

// ---------------------------------------------------------------------------
// Non-Go project — no package or import edges
// ---------------------------------------------------------------------------

func TestBuildFileGraph_NonGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"index.html": "<html></html>",
		"style.css":  "body {}",
	})
	g := buildFileGraph(t, root)

	// Nodes: 2 files + 1 root dir node.
	for _, n := range g.Nodes {
		if n.Kind == codebase.NodePackage {
			t.Errorf("want no package nodes for non-Go project, got %+v", n)
		}
	}

	// Only contains edges (dir->file) — no imports or belongs_to.
	for _, e := range g.Edges {
		if e.Kind != codebase.RelationshipContains {
			t.Errorf("want only contains edges for non-Go project, got kind %q", e.Kind)
		}
	}
}

// ---------------------------------------------------------------------------
// Single Go file — directory contains file, file belongs_to package
// ---------------------------------------------------------------------------

func TestBuildFileGraph_SingleGoFile(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/app/app.go": `package app
`,
	})
	g := buildFileGraph(t, root)

	fileID := "file:internal/app/app.go"
	dirID := "dir:internal/app"
	pkgID := "package:app"

	if findNode(g, fileID) == nil {
		t.Errorf("want node %q", fileID)
	}
	if findNode(g, dirID) == nil {
		t.Errorf("want node %q", dirID)
	}
	if findNode(g, pkgID) == nil {
		t.Errorf("want node %q", pkgID)
	}

	if !hasEdge(g, dirID, fileID, codebase.RelationshipContains) {
		t.Errorf("want contains edge %q -> %q", dirID, fileID)
	}
	if !hasEdge(g, fileID, pkgID, codebase.RelationshipBelongsTo) {
		t.Errorf("want belongs_to edge %q -> %q", fileID, pkgID)
	}
}

// ---------------------------------------------------------------------------
// File imports package
// ---------------------------------------------------------------------------

func TestBuildFileGraph_FileImportsPackage(t *testing.T) {
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
	g := buildFileGraph(t, root)

	fileID := "file:internal/cli/cli.go"
	pkgID := "package:config"

	if !hasEdge(g, fileID, pkgID, codebase.RelationshipImports) {
		t.Errorf("want imports edge %q -> %q", fileID, pkgID)
	}
}

// ---------------------------------------------------------------------------
// Standard library imports produce no edges
// ---------------------------------------------------------------------------

func TestBuildFileGraph_StdlibImports_NoEdge(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import (
    "fmt"
    "os"
    "strings"
)

var _ = fmt.Sprintf
var _ = os.Exit
var _ = strings.TrimSpace
`,
	})
	g := buildFileGraph(t, root)

	for _, e := range g.Edges {
		if e.Kind == codebase.RelationshipImports {
			t.Errorf("want no imports edges for stdlib-only file, got %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Duplicate imports collapse to one edge
// ---------------------------------------------------------------------------

func TestBuildFileGraph_DuplicateImports_OneEdge(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/cli/cli.go": `package cli

import (
    "example.com/app/internal/config"
)

var _ = config.Config{}
var _ = config.Config{}
`,
		"internal/config/config.go": `package config

type Config struct{}
`,
	})
	g := buildFileGraph(t, root)

	fileID := "file:internal/cli/cli.go"
	pkgID := "package:config"

	count := 0
	for _, e := range g.Edges {
		if e.From == fileID && e.To == pkgID && e.Kind == codebase.RelationshipImports {
			count++
		}
	}
	if count != 1 {
		t.Errorf("want exactly 1 imports edge, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Nested directories
// ---------------------------------------------------------------------------

func TestBuildFileGraph_NestedDirectories(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/a/b/c/deep.go": `package c
`,
	})
	g := buildFileGraph(t, root)

	fileID := "file:internal/a/b/c/deep.go"
	dirID := "dir:internal/a/b/c"

	if findNode(g, fileID) == nil {
		t.Errorf("want node %q", fileID)
	}
	if findNode(g, dirID) == nil {
		t.Errorf("want node %q", dirID)
	}
	if !hasEdge(g, dirID, fileID, codebase.RelationshipContains) {
		t.Errorf("want contains edge from %q to %q", dirID, fileID)
	}
}

// ---------------------------------------------------------------------------
// Multiple files in one package
// ---------------------------------------------------------------------------

func TestBuildFileGraph_MultipleFilesInPackage(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                      "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go":   "package config\n",
		"internal/config/defaults.go": "package config\n",
		"internal/config/validate.go": "package config\n",
	})
	g := buildFileGraph(t, root)

	pkgID := "package:config"

	count := 0
	for _, e := range g.Edges {
		if e.To == pkgID && e.Kind == codebase.RelationshipBelongsTo {
			count++
		}
	}
	if count != 3 {
		t.Errorf("want 3 belongs_to edges to %q, got %d", pkgID, count)
	}
}

// ---------------------------------------------------------------------------
// No false relationships between unrelated files
// ---------------------------------------------------------------------------

func TestBuildFileGraph_NoFalseRelationships(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/app\n\ngo 1.21\n",
		"internal/alpha/alpha.go": `package alpha
`,
		"internal/beta/beta.go": `package beta
`,
	})
	g := buildFileGraph(t, root)

	// alpha and beta do not import each other.
	// No imports edge should exist between them.
	alphaFileID := "file:internal/alpha/alpha.go"
	betaPkgID := "package:beta"
	betaFileID := "file:internal/beta/beta.go"
	alphaPkgID := "package:alpha"

	if hasEdge(g, alphaFileID, betaPkgID, codebase.RelationshipImports) {
		t.Error("false imports edge: alpha -> beta")
	}
	if hasEdge(g, betaFileID, alphaPkgID, codebase.RelationshipImports) {
		t.Error("false imports edge: beta -> alpha")
	}
}

// ---------------------------------------------------------------------------
// Determinism
// ---------------------------------------------------------------------------

func TestBuildFileGraph_Determinism(t *testing.T) {
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

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}
	cb := codebase.Discover(root, inv)

	g1 := codebase.BuildFileGraph(root, cb, inv)
	g2 := codebase.BuildFileGraph(root, cb, inv)

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
		e1, e2 := g1.Edges[i], g1.Edges[i]
		if e1.From != e2.From || e1.To != e2.To || e1.Kind != e2.Kind {
			t.Errorf("edge[%d] mismatch", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Windows-style paths must not appear as node IDs
// ---------------------------------------------------------------------------

func TestBuildFileGraph_NoWindowsPathsInIDs(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":                    "module example.com/app\n\ngo 1.21\n",
		"internal/config/config.go": "package config\n",
	})
	g := buildFileGraph(t, root)

	for _, n := range g.Nodes {
		if containsBackslash(n.ID) {
			t.Errorf("node ID contains backslash (Windows path): %q", n.ID)
		}
	}
}

// containsBackslash reports whether s contains a backslash character.
func containsBackslash(s string) bool {
	for _, c := range s {
		if c == '\\' {
			return true
		}
	}
	return false
}
