package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/project"
	"pulse/internal/scanner"
)

// ---------------------------------------------------------------------------
// File model
// ---------------------------------------------------------------------------

func TestBuildFiles_SingleFile(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "main.go", Extension: ".go", SizeBytes: 100},
	}

	files := codebase.BuildFiles(entries)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	assertEqual(t, "main.go", f.Path)
	assertEqual(t, "main.go", f.Name)
	assertEqual(t, ".go", f.Extension)
	assertEqual(t, project.LangGo, f.Language)
}

func TestBuildFiles_NestedPath(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "internal/cli/cli.go", Extension: ".go", SizeBytes: 200},
	}

	files := codebase.BuildFiles(entries)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	assertEqual(t, "internal/cli/cli.go", files[0].Path)
	assertEqual(t, "cli.go", files[0].Name)
}

func TestBuildFiles_ExtensionExtraction(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "Makefile", Extension: "", SizeBytes: 50},
		{RelPath: "config.yaml", Extension: ".yaml", SizeBytes: 30},
		{RelPath: "main.go", Extension: ".go", SizeBytes: 100},
	}

	files := codebase.BuildFiles(entries)

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	// Sorted alphabetically by path
	assertEqual(t, "Makefile", files[0].Name)
	assertEqual(t, "", files[0].Extension)

	assertEqual(t, "config.yaml", files[1].Name)
	assertEqual(t, ".yaml", files[1].Extension)

	assertEqual(t, "main.go", files[2].Name)
	assertEqual(t, ".go", files[2].Extension)
}

func TestBuildFiles_LanguageAssociation(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "main.go", Extension: ".go"},
		{RelPath: "README.md", Extension: ".md"},
		{RelPath: "script.py", Extension: ".py"},
		{RelPath: "Makefile", Extension: ""},
	}

	files := codebase.BuildFiles(entries)

	want := map[string]project.Language{
		"main.go":   project.LangGo,
		"README.md": project.LangMarkdown,
		"script.py": project.LangPython,
		"Makefile":  "",
	}

	for _, f := range files {
		expected := want[f.Name]
		if f.Language != expected {
			t.Errorf("file %q: want language %q, got %q", f.Name, expected, f.Language)
		}
	}
}

func TestBuildFiles_DeterministicOrdering(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "z.go", Extension: ".go"},
		{RelPath: "a.go", Extension: ".go"},
		{RelPath: "m.go", Extension: ".go"},
	}

	files1 := codebase.BuildFiles(entries)
	files2 := codebase.BuildFiles(entries)

	if len(files1) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files1))
	}

	for i := range files1 {
		if files1[i].Path != files2[i].Path {
			t.Errorf("non-deterministic at index %d: %q vs %q", i, files1[i].Path, files2[i].Path)
		}
	}

	assertEqual(t, "a.go", files1[0].Path)
	assertEqual(t, "m.go", files1[1].Path)
	assertEqual(t, "z.go", files1[2].Path)
}

func TestBuildFiles_Empty(t *testing.T) {
	files := codebase.BuildFiles(nil)
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// Directory model
// ---------------------------------------------------------------------------

func TestBuildDirectories_SingleDir(t *testing.T) {
	entries := []scanner.DirEntry{
		{RelPath: "internal"},
	}

	dirs := codebase.BuildDirectories(entries)

	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(dirs))
	}
	assertEqual(t, "internal", dirs[0].Path)
	assertEqual(t, "internal", dirs[0].Name)
}

func TestBuildDirectories_NestedDirs(t *testing.T) {
	entries := []scanner.DirEntry{
		{RelPath: "internal"},
		{RelPath: "internal/cli"},
		{RelPath: "internal/git"},
		{RelPath: "internal/output"},
	}

	dirs := codebase.BuildDirectories(entries)

	if len(dirs) != 4 {
		t.Fatalf("expected 4 dirs, got %d", len(dirs))
	}

	assertEqual(t, "internal", dirs[0].Path)
	assertEqual(t, "internal", dirs[0].Name)

	assertEqual(t, "internal/cli", dirs[1].Path)
	assertEqual(t, "cli", dirs[1].Name)

	assertEqual(t, "internal/git", dirs[2].Path)
	assertEqual(t, "git", dirs[2].Name)

	assertEqual(t, "internal/output", dirs[3].Path)
	assertEqual(t, "output", dirs[3].Name)
}

func TestBuildDirectories_DeterministicOrdering(t *testing.T) {
	entries := []scanner.DirEntry{
		{RelPath: "z"},
		{RelPath: "a"},
		{RelPath: "m"},
	}

	dirs1 := codebase.BuildDirectories(entries)
	dirs2 := codebase.BuildDirectories(entries)

	for i := range dirs1 {
		if dirs1[i].Path != dirs2[i].Path {
			t.Errorf("non-deterministic at index %d: %q vs %q", i, dirs1[i].Path, dirs2[i].Path)
		}
	}

	assertEqual(t, "a", dirs1[0].Path)
	assertEqual(t, "m", dirs1[1].Path)
	assertEqual(t, "z", dirs1[2].Path)
}

func TestBuildDirectories_Empty(t *testing.T) {
	dirs := codebase.BuildDirectories(nil)
	if len(dirs) != 0 {
		t.Errorf("expected 0 dirs, got %d", len(dirs))
	}
}

// ---------------------------------------------------------------------------
// FilesByDirectory
// ---------------------------------------------------------------------------

func TestFilesByDirectory(t *testing.T) {
	files := []codebase.File{
		{Path: "main.go", Name: "main.go"},
		{Path: "internal/cli/cli.go", Name: "cli.go"},
		{Path: "internal/cli/flags.go", Name: "flags.go"},
		{Path: "internal/git/git.go", Name: "git.go"},
	}

	grouped := codebase.FilesByDirectory(files)

	if len(grouped[""]) != 1 {
		t.Errorf("expected 1 root file, got %d", len(grouped[""]))
	}
	if len(grouped["internal/cli"]) != 2 {
		t.Errorf("expected 2 cli files, got %d", len(grouped["internal/cli"]))
	}
	if len(grouped["internal/git"]) != 1 {
		t.Errorf("expected 1 git file, got %d", len(grouped["internal/git"]))
	}
}

// ---------------------------------------------------------------------------
// Boundary: paths stay within target
// ---------------------------------------------------------------------------

func TestBuildFiles_PathsAreRelative(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "internal/app.go", Extension: ".go"},
	}

	files := codebase.BuildFiles(entries)

	for _, f := range files {
		if len(f.Path) > 0 && f.Path[0] == '/' {
			t.Errorf("absolute path detected: %q", f.Path)
		}
	}
}

func TestBuildDirectories_PathsAreRelative(t *testing.T) {
	entries := []scanner.DirEntry{
		{RelPath: "internal"},
	}

	dirs := codebase.BuildDirectories(entries)

	for _, d := range dirs {
		if len(d.Path) > 0 && d.Path[0] == '/' {
			t.Errorf("absolute path detected: %q", d.Path)
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func assertEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}
