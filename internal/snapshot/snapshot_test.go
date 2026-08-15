package snapshot_test

import (
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
	"github.com/raghavendrashivam474/aayam/internal/testhelpers"
)

func assertLanguages(t *testing.T, got []project.Language, want []project.Language) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("language count mismatch: want %d, got %d (%v)", len(want), len(got), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("languages[%d]: want %q, got %q", i, want[i], got[i])
		}
	}
}

// TestDiscover_BasicGoProject verifies the full discovery pipeline
// produces a correct snapshot from a small Go project fixture.
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

	testhelpers.AssertEqual(t, root, snap.Root)

	if snap.Name == "" {
		t.Error("Name must not be empty")
	}

	testhelpers.AssertEqual(t, project.TypeGo, snap.Type)
	testhelpers.AssertEqual(t, 4, snap.FileCount)
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)

	assertLanguages(t, snap.Languages, []project.Language{
		project.LangGo,
		project.LangMarkdown,
	})

	testhelpers.AssertEqual(t, snap.FileCount, len(snap.Files))
	testhelpers.AssertEqual(t, snap.DirectoryCount, len(snap.Directories))
}

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

	assertLanguages(t, snap.Languages, []project.Language{
		project.LangCSS,
		project.LangJavaScript,
		project.LangMarkdown,
		project.LangTypeScript,
	})

	testhelpers.AssertEqual(t, 6, snap.FileCount)
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)
}

func TestDiscover_InvalidTarget(t *testing.T) {
	_, err := snapshot.Discover("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent target, got nil")
	}
}

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

func TestDiscover_GoFixture_BoundedToTarget(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")

	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, fixturePath, snap.Root)
	testhelpers.AssertEqual(t, "go-project", snap.Name)
	testhelpers.AssertEqual(t, project.TypeGo, snap.Type)
	testhelpers.AssertEqual(t, 4, snap.FileCount)
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)

	assertLanguages(t, snap.Languages, []project.Language{
		project.LangGo,
		project.LangMarkdown,
		project.LangPowerShell,
	})

	if snap.Name == "Aryntra Aayam" {
		t.Fatal("ancestor trap: got Aryntra Aayam instead of go-project")
	}
}

func TestDiscover_MixedFixture_BoundedToTarget(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "mixed-project")

	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, fixturePath, snap.Root)
	testhelpers.AssertEqual(t, "mixed-project", snap.Name)
	testhelpers.AssertEqual(t, 4, snap.FileCount)
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)

	assertLanguages(t, snap.Languages, []project.Language{
		project.LangGo,
		project.LangJavaScript,
		project.LangMarkdown,
		project.LangPython,
	})

	if snap.Name == "Aryntra Aayam" {
		t.Fatal("ancestor trap: got Aryntra Aayam instead of mixed-project")
	}
}

func TestDiscover_NodeFixture_BoundedToTarget(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "node-project")

	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, fixturePath, snap.Root)
	testhelpers.AssertEqual(t, "node-project", snap.Name)
	testhelpers.AssertEqual(t, project.TypeNode, snap.Type)
	testhelpers.AssertEqual(t, 5, snap.FileCount)
	testhelpers.AssertEqual(t, 1, snap.DirectoryCount)

	assertLanguages(t, snap.Languages, []project.Language{
		project.LangCSS,
		project.LangHTML,
		project.LangTypeScript,
	})
}

func TestDiscover_UnknownFixture_BoundedToTarget(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "unknown-project")

	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	testhelpers.AssertEqual(t, fixturePath, snap.Root)
	testhelpers.AssertEqual(t, "unknown-project", snap.Name)
	testhelpers.AssertEqual(t, project.TypeUnknown, snap.Type)
	testhelpers.AssertEqual(t, 2, snap.FileCount)
	testhelpers.AssertEqual(t, 0, snap.DirectoryCount)
	testhelpers.AssertEqual(t, 0, len(snap.Languages))

	if snap.Name == "Aryntra Aayam" {
		t.Fatal("ancestor trap: got Aryntra Aayam instead of unknown-project")
	}
}
