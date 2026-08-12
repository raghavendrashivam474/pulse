package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/project"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Single Go package
// ---------------------------------------------------------------------------

func TestDiscoverPackages_SinglePackage(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main\n\nfunc main() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d: %v", len(pkgs), pkgs)
	}

	assertEqual(t, "main", pkgs[0].Name)
	assertEqual(t, "", pkgs[0].Path)
	assertEqual(t, project.LangGo, pkgs[0].Language)

	if len(pkgs[0].Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(pkgs[0].Files))
	}
	assertEqual(t, "main.go", pkgs[0].Files[0])
}

// ---------------------------------------------------------------------------
// Multiple packages
// ---------------------------------------------------------------------------

func TestDiscoverPackages_MultiplePackages(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"main.go":                "package main\n",
		"internal/cli/cli.go":    "package cli\n",
		"internal/output/out.go": "package output\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}

	// Sorted by path then name: "" < "internal/cli" < "internal/output"
	assertEqual(t, "main", pkgs[0].Name)
	assertEqual(t, "", pkgs[0].Path)

	assertEqual(t, "cli", pkgs[1].Name)
	assertEqual(t, "internal/cli", pkgs[1].Path)

	assertEqual(t, "output", pkgs[2].Name)
	assertEqual(t, "internal/output", pkgs[2].Path)
}

// ---------------------------------------------------------------------------
// Multiple files in one package
// ---------------------------------------------------------------------------

func TestDiscoverPackages_MultipleFilesInPackage(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"internal/cli/cli.go":   "package cli\n",
		"internal/cli/flags.go": "package cli\n",
		"internal/cli/run.go":   "package cli\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	assertEqual(t, "cli", pkgs[0].Name)

	if len(pkgs[0].Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(pkgs[0].Files))
	}

	// Files should be sorted
	assertEqual(t, "internal/cli/cli.go", pkgs[0].Files[0])
	assertEqual(t, "internal/cli/flags.go", pkgs[0].Files[1])
	assertEqual(t, "internal/cli/run.go", pkgs[0].Files[2])
}

// ---------------------------------------------------------------------------
// External _test package
// ---------------------------------------------------------------------------

func TestDiscoverPackages_ExternalTestPackage(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"internal/cli/cli.go":      "package cli\n",
		"internal/cli/cli_test.go": "package cli_test\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (cli and cli_test), got %d", len(pkgs))
	}

	// Both in same directory, sorted by name
	assertEqual(t, "cli", pkgs[0].Name)
	assertEqual(t, "internal/cli", pkgs[0].Path)

	assertEqual(t, "cli_test", pkgs[1].Name)
	assertEqual(t, "internal/cli", pkgs[1].Path)
}

func TestIsTestPackage(t *testing.T) {
	if !codebase.IsTestPackage("cli_test") {
		t.Error("expected cli_test to be a test package")
	}
	if codebase.IsTestPackage("cli") {
		t.Error("expected cli to not be a test package")
	}
	if codebase.IsTestPackage("testing") {
		t.Error("expected testing to not be a test package")
	}
}

// ---------------------------------------------------------------------------
// Nested packages
// ---------------------------------------------------------------------------

func TestDiscoverPackages_NestedPackages(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"a/b/c/deep.go": "package deep\n",
		"a/b/mid.go":    "package mid\n",
		"a/top.go":      "package top\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}

	assertEqual(t, "top", pkgs[0].Name)
	assertEqual(t, "a", pkgs[0].Path)

	assertEqual(t, "mid", pkgs[1].Name)
	assertEqual(t, "a/b", pkgs[1].Path)

	assertEqual(t, "deep", pkgs[2].Name)
	assertEqual(t, "a/b/c", pkgs[2].Path)
}

// ---------------------------------------------------------------------------
// No Go files
// ---------------------------------------------------------------------------

func TestDiscoverPackages_NoGoFiles(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"README.md":    "# hello\n",
		"package.json": "{}\n",
		"index.js":     "console.log('hi');\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages for non-Go project, got %d", len(pkgs))
	}
}

// ---------------------------------------------------------------------------
// Empty project
// ---------------------------------------------------------------------------

func TestDiscoverPackages_EmptyProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

// ---------------------------------------------------------------------------
// Malformed Go file
// ---------------------------------------------------------------------------

func TestDiscoverPackages_MalformedGoFile(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"good.go": "package main\n",
		"bad.go":  "this is not valid go source\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	// Should not crash; should discover the valid package only.
	pkgs := codebase.DiscoverPackages(root, inv)

	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (from good.go), got %d", len(pkgs))
	}
	assertEqual(t, "main", pkgs[0].Name)
}

// ---------------------------------------------------------------------------
// Deterministic ordering
// ---------------------------------------------------------------------------

func TestDiscoverPackages_DeterministicOrdering(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"z/z.go": "package zpkg\n",
		"a/a.go": "package apkg\n",
		"m/m.go": "package mpkg\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs1 := codebase.DiscoverPackages(root, inv)
	pkgs2 := codebase.DiscoverPackages(root, inv)

	if len(pkgs1) != len(pkgs2) {
		t.Fatalf("different package counts: %d vs %d", len(pkgs1), len(pkgs2))
	}

	for i := range pkgs1 {
		if pkgs1[i].Name != pkgs2[i].Name || pkgs1[i].Path != pkgs2[i].Path {
			t.Errorf("non-deterministic at %d: (%q,%q) vs (%q,%q)",
				i, pkgs1[i].Path, pkgs1[i].Name, pkgs2[i].Path, pkgs2[i].Name)
		}
	}

	assertEqual(t, "a", pkgs1[0].Path)
	assertEqual(t, "m", pkgs1[1].Path)
	assertEqual(t, "z", pkgs1[2].Path)
}

// ---------------------------------------------------------------------------
// PackageNames helper
// ---------------------------------------------------------------------------

func TestPackageNames(t *testing.T) {
	pkgs := []codebase.Package{
		{Name: "output", Path: "internal/output"},
		{Name: "cli", Path: "internal/cli"},
		{Name: "cli_test", Path: "internal/cli"},
		{Name: "main", Path: ""},
	}

	names := codebase.PackageNames(pkgs)

	expected := []string{"cli", "cli_test", "main", "output"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d: %v", len(expected), len(names), names)
	}
	for i := range expected {
		assertEqual(t, expected[i], names[i])
	}
}
