package codebase_test

import (
	"testing"

	"pulse/internal/codebase"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// ---------------------------------------------------------------------------
// Single import
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_SingleImport(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import "testmod/internal/cli"

func main() { cli.Run() }
`,
		"internal/cli/cli.go": "package cli\n\nfunc Run() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d: %v", len(deps), deps)
	}

	assertEqual(t, "main", deps[0].From)
	assertEqual(t, "cli", deps[0].To)
	assertEqual(t, codebase.DependencyImport, deps[0].Type)
}

// ---------------------------------------------------------------------------
// Multiple imports
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_MultipleImports(t *testing.T) {
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
		"internal/cli/cli.go":       "package cli\n\nfunc Run() {}\n",
		"internal/output/output.go": "package output\n\nfunc Print() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies, got %d: %v", len(deps), deps)
	}

	// Sorted: (main→cli, main→output)
	assertEqual(t, "main", deps[0].From)
	assertEqual(t, "cli", deps[0].To)

	assertEqual(t, "main", deps[1].From)
	assertEqual(t, "output", deps[1].To)
}

// ---------------------------------------------------------------------------
// Standard library imports ignored
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_StdlibIgnored(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import (
    "fmt"
    "os"
    "testmod/internal/cli"
)

func main() {
    fmt.Println(os.Args)
    cli.Run()
}
`,
		"internal/cli/cli.go": "package cli\n\nfunc Run() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency (stdlib ignored), got %d: %v", len(deps), deps)
	}

	assertEqual(t, "main", deps[0].From)
	assertEqual(t, "cli", deps[0].To)
}

// ---------------------------------------------------------------------------
// Duplicate dependency deduplication
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_Deduplication(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"a.go": `package main

import "testmod/internal/cli"

func a() { cli.Run() }
`,
		"b.go": `package main

import "testmod/internal/cli"

func b() { cli.Run() }
`,
		"internal/cli/cli.go": "package cli\n\nfunc Run() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency (deduplicated), got %d: %v", len(deps), deps)
	}
}

// ---------------------------------------------------------------------------
// Multiple files importing same package
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_MultipleFilesImportingSame(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"internal/cli/run.go": `package cli

import "testmod/internal/config"

func Run() { config.Load() }
`,
		"internal/cli/help.go": `package cli

import "testmod/internal/config"

func Help() { config.Load() }
`,
		"internal/config/config.go": "package config\n\nfunc Load() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	// cli→config should appear exactly once even though two files import it.
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d: %v", len(deps), deps)
	}

	assertEqual(t, "cli", deps[0].From)
	assertEqual(t, "config", deps[0].To)
}

// ---------------------------------------------------------------------------
// Cyclic dependencies don't crash
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_CyclicSafe(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"internal/a/a.go": `package a

import "testmod/internal/b"

func A() { b.B() }
`,
		"internal/b/b.go": `package b

import "testmod/internal/a"

func B() { a.A() }
`,
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	// Should have both edges without crashing.
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies for cycle, got %d: %v", len(deps), deps)
	}

	// Sorted: (a→b, b→a)
	assertEqual(t, "a", deps[0].From)
	assertEqual(t, "b", deps[0].To)

	assertEqual(t, "b", deps[1].From)
	assertEqual(t, "a", deps[1].To)
}

// ---------------------------------------------------------------------------
// No go.mod — no dependencies
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_NoGoMod(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main\n\nfunc main() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies without go.mod, got %d", len(deps))
	}
}

// ---------------------------------------------------------------------------
// Malformed Go file handled
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_MalformedFile(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import "testmod/internal/cli"

func main() { cli.Run() }
`,
		"bad.go":              "this is not valid go at all\n",
		"internal/cli/cli.go": "package cli\n\nfunc Run() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	// Should still find the valid dependency.
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency despite malformed file, got %d: %v", len(deps), deps)
	}

	assertEqual(t, "main", deps[0].From)
	assertEqual(t, "cli", deps[0].To)
}

// ---------------------------------------------------------------------------
// Nested internal packages
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_NestedPackages(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"internal/a/a.go": `package a

import "testmod/internal/a/b"

func A() { b.B() }
`,
		"internal/a/b/b.go": "package b\n\nfunc B() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d: %v", len(deps), deps)
	}

	assertEqual(t, "a", deps[0].From)
	assertEqual(t, "b", deps[0].To)
}

// ---------------------------------------------------------------------------
// Deterministic ordering
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_DeterministicOrdering(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import (
    "testmod/internal/z"
    "testmod/internal/a"
    "testmod/internal/m"
)

func main() {
    z.Z()
    a.A()
    m.M()
}
`,
		"internal/z/z.go": "package z\n\nfunc Z() {}\n",
		"internal/a/a.go": "package a\n\nfunc A() {}\n",
		"internal/m/m.go": "package m\n\nfunc M() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps1 := codebase.DiscoverDependencies(root, inv, pkgs)
	deps2 := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps1) != len(deps2) {
		t.Fatalf("different counts: %d vs %d", len(deps1), len(deps2))
	}

	for i := range deps1 {
		if deps1[i].From != deps2[i].From || deps1[i].To != deps2[i].To {
			t.Errorf("non-deterministic at %d", i)
		}
	}

	// Sorted alphabetically: (main→a, main→m, main→z)
	assertEqual(t, "a", deps1[0].To)
	assertEqual(t, "m", deps1[1].To)
	assertEqual(t, "z", deps1[2].To)
}

// ---------------------------------------------------------------------------
// Aliased imports don't change identity
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_AliasedImport(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module testmod\n\ngo 1.21\n",
		"main.go": `package main

import myalias "testmod/internal/cli"

func main() { myalias.Run() }
`,
		"internal/cli/cli.go": "package cli\n\nfunc Run() {}\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d: %v", len(deps), deps)
	}

	// Target is the actual package name, not the alias.
	assertEqual(t, "main", deps[0].From)
	assertEqual(t, "cli", deps[0].To)
}

// ---------------------------------------------------------------------------
// Non-Go project: no crash
// ---------------------------------------------------------------------------

func TestDiscoverDependencies_NonGoProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{
		"package.json": "{}\n",
		"index.js":     "console.log('hello');\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	pkgs := codebase.DiscoverPackages(root, inv)
	deps := codebase.DiscoverDependencies(root, inv, pkgs)

	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies for non-Go project, got %d", len(deps))
	}
}
