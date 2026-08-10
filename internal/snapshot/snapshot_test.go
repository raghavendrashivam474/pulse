package snapshot_test

import (
	"testing"

	"pulse/internal/project"
	"pulse/internal/snapshot"
	"pulse/internal/testhelpers"
)

// TestDiscover_BasicGoProject verifies the full discovery pipeline
// produces a correct snapshot from a small Go project fixture.
//
// Fixture:
//
//	go.mod
//	main.go
//	README.md
//	internal/app.go
func TestDiscover_BasicGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":          "module basic\n\ngo 1.21\n",
		"main.go":         "package main\n",
		"README.md":       "# basic\n",
		"internal/app.go": "package internal\n",
	})

	snap, err := snapshot.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// Root is preserved.
	testhelpers.AssertEqual(t, root, snap.Root)

	// Name is the directory name.
	if snap.Name == "" {
		t.Error("Name must not be empty")
	}

	// Type detected from go.mod.
	testhelpers.AssertEqual(t, project.TypeGo, snap.Type)

	// File count: go.mod, main.go, README.md, internal/app.go = 4.
	testhelpers.AssertEqual(t, 4, snap.FileCount)

	// Directory count: internal/ only, root excluded.
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)

	// Languages: Go (.go), Markdown (.md). Alphabetical order.
	testhelpers.AssertEqual(t, 2, len(snap.Languages))
	testhelpers.AssertEqual(t, project.LangGo, snap.Languages[0])
	testhelpers.AssertEqual(t, project.LangMarkdown, snap.Languages[1])

	// Files slice matches FileCount.
	testhelpers.AssertEqual(t, snap.FileCount, len(snap.Files))

	// Directories slice matches DirectoryCount.
	testhelpers.AssertEqual(t, snap.DirectoryCount, len(snap.Directories))
}

// TestDiscover_EmptyProject verifies that an empty directory produces
// a valid snapshot without panicking.
func TestDiscover_EmptyProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})

	snap, err := snapshot.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, root, snap.Root)
	testhelpers.AssertEqual(t, project.TypeUnknown, snap.Type)
	testhelpers.AssertEqual(t, 0, snap.FileCount)
	testhelpers.AssertEqual(t, 0, snap.DirectoryCount)
	testhelpers.AssertEqual(t, 0, len(snap.Languages))
}

// TestDiscover_MixedLanguages verifies that a project with multiple
// file types reports all detected languages.
func TestDiscover_MixedLanguages(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"package.json": "{}\n",
		"index.ts":     "console.log('hello');\n",
		"style.css":    "body {}\n",
		"README.md":    "# app\n",
		"src/app.tsx":  "export default App;\n",
		"src/utils.js": "module.exports = {};\n",
	})

	snap, err := snapshot.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, project.TypeNode, snap.Type)

	// CSS, JavaScript, Markdown, TypeScript (alphabetical).
	testhelpers.AssertEqual(t, 4, len(snap.Languages))
	testhelpers.AssertEqual(t, project.LangCSS, snap.Languages[0])
	testhelpers.AssertEqual(t, project.LangJavaScript, snap.Languages[1])
	testhelpers.AssertEqual(t, project.LangMarkdown, snap.Languages[2])
	testhelpers.AssertEqual(t, project.LangTypeScript, snap.Languages[3])

	// 6 files: package.json, index.ts, style.css, README.md, src/app.tsx, src/utils.js.
	testhelpers.AssertEqual(t, 6, snap.FileCount)

	// 1 directory: src/.
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)
}

// TestDiscover_InvalidTarget verifies that a nonexistent target returns
// an error rather than a partial snapshot.
func TestDiscover_InvalidTarget(t *testing.T) {
	_, err := snapshot.Discover("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent target, got nil")
	}
}

// TestDiscover_DeterministicOutput verifies that two calls with the same
// input produce identical snapshots.
func TestDiscover_DeterministicOutput(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":    "module test\n\ngo 1.21\n",
		"main.go":   "package main\n",
		"README.md": "# test\n",
	})

	snap1, err1 := snapshot.Discover(root)
	snap2, err2 := snapshot.Discover(root)

	if err1 != nil {
		t.Fatalf("first Discover: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("second Discover: %v", err2)
	}

	testhelpers.AssertEqual(t, snap1.Name, snap2.Name)
	testhelpers.AssertEqual(t, snap1.Root, snap2.Root)
	testhelpers.AssertEqual(t, snap1.Type, snap2.Type)
	testhelpers.AssertEqual(t, snap1.FileCount, snap2.FileCount)
	testhelpers.AssertEqual(t, snap1.DirectoryCount, snap2.DirectoryCount)
	testhelpers.AssertEqual(t, len(snap1.Languages), len(snap2.Languages))

	for i := range snap1.Languages {
		testhelpers.AssertEqual(t, snap1.Languages[i], snap2.Languages[i])
	}

	testhelpers.AssertEqual(t, len(snap1.Files), len(snap2.Files))
	for i := range snap1.Files {
		testhelpers.AssertEqual(t, snap1.Files[i].RelPath, snap2.Files[i].RelPath)
	}
}
