package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Codebase.Discover — full integration
// ---------------------------------------------------------------------------

func TestDiscover_FullGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import (
    "testmod/internal/cli"
    "testmod/internal/output"
)

func main() {
    cli.Run()
    output.Print()
}
`,
		"internal/cli/cli.go": `package cli

import "testmod/internal/config"

func Run() { config.Load() }
`,
		"internal/output/output.go": "package output\n\nfunc Print() {}\n",
		"internal/config/config.go": "package config\n\nfunc Load() {}\n",
		"README.md":                 "# testmod\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	cb := codebase.Discover(root, inv)

	// Files: go.mod, main.go, README.md,
	//        internal/cli/cli.go, internal/output/output.go, internal/config/config.go
	if len(cb.Files) != 6 {
		t.Errorf("expected 6 files, got %d", len(cb.Files))
		for _, f := range cb.Files {
			t.Logf("  file: %s", f.Path)
		}
	}

	// Directories: internal, internal/cli, internal/output, internal/config
	if len(cb.Directories) != 4 {
		t.Errorf("expected 4 directories, got %d", len(cb.Directories))
		for _, d := range cb.Directories {
			t.Logf("  dir: %s", d.Path)
		}
	}

	// Packages: main, cli, config, output
	if len(cb.Packages) != 4 {
		t.Errorf("expected 4 packages, got %d", len(cb.Packages))
		for _, p := range cb.Packages {
			t.Logf("  package: %s at %s", p.Name, p.Path)
		}
	}

	names := codebase.PackageNames(cb.Packages)
	expectedNames := []string{"cli", "config", "main", "output"}
	if len(names) != len(expectedNames) {
		t.Fatalf("expected %d package names, got %d: %v", len(expectedNames), len(names), names)
	}
	for i := range expectedNames {
		assertEqual(t, expectedNames[i], names[i])
	}

	// Dependencies: cli→config, main→cli, main→output
	if len(cb.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d: %v", len(cb.Dependencies), cb.Dependencies)
	}

	assertEqual(t, "cli", cb.Dependencies[0].From)
	assertEqual(t, "config", cb.Dependencies[0].To)

	assertEqual(t, "main", cb.Dependencies[1].From)
	assertEqual(t, "cli", cb.Dependencies[1].To)

	assertEqual(t, "main", cb.Dependencies[2].From)
	assertEqual(t, "output", cb.Dependencies[2].To)
}

func TestDiscover_EmptyProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	cb := codebase.Discover(root, inv)

	if len(cb.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(cb.Files))
	}
	if len(cb.Directories) != 0 {
		t.Errorf("expected 0 directories, got %d", len(cb.Directories))
	}
	if len(cb.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(cb.Packages))
	}
	if len(cb.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(cb.Dependencies))
	}
}

func TestDiscover_NonGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"package.json": "{}\n",
		"index.js":     "console.log('hi');\n",
		"src/app.ts":   "export default {};\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	cb := codebase.Discover(root, inv)

	// Files and directories should still be populated.
	if len(cb.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(cb.Files))
	}
	if len(cb.Directories) != 1 {
		t.Errorf("expected 1 directory, got %d", len(cb.Directories))
	}

	// No Go packages or dependencies.
	if len(cb.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(cb.Packages))
	}
	if len(cb.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(cb.Dependencies))
	}
}

func TestDiscover_Deterministic(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import "testmod/internal/svc"

func main() { svc.Start() }
`,
		"internal/svc/svc.go": "package svc\n\nfunc Start() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	cb1 := codebase.Discover(root, inv)
	cb2 := codebase.Discover(root, inv)

	// Files
	if len(cb1.Files) != len(cb2.Files) {
		t.Fatalf("file count mismatch: %d vs %d", len(cb1.Files), len(cb2.Files))
	}
	for i := range cb1.Files {
		assertEqual(t, cb1.Files[i].Path, cb2.Files[i].Path)
	}

	// Packages
	if len(cb1.Packages) != len(cb2.Packages) {
		t.Fatalf("package count mismatch: %d vs %d", len(cb1.Packages), len(cb2.Packages))
	}
	for i := range cb1.Packages {
		assertEqual(t, cb1.Packages[i].Name, cb2.Packages[i].Name)
		assertEqual(t, cb1.Packages[i].Path, cb2.Packages[i].Path)
	}

	// Dependencies
	if len(cb1.Dependencies) != len(cb2.Dependencies) {
		t.Fatalf("dependency count mismatch: %d vs %d", len(cb1.Dependencies), len(cb2.Dependencies))
	}
	for i := range cb1.Dependencies {
		assertEqual(t, cb1.Dependencies[i].From, cb2.Dependencies[i].From)
		assertEqual(t, cb1.Dependencies[i].To, cb2.Dependencies[i].To)
	}
}
