package project_test

import (
	"path/filepath"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/testhelpers"
)

// resolveTarget wraps ResolveTarget and fails fast on error.
func resolveTarget(t *testing.T, path string) project.Target {
	t.Helper()
	target, err := project.ResolveTarget(path)
	if err != nil {
		t.Fatalf("ResolveTarget(%q) unexpected error: %v", path, err)
	}
	return target
}

func TestDiscoverRoot_TargetIsAlreadyRoot(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/test\n\ngo 1.21\n",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if result.Root != filepath.Clean(dir) {
		t.Errorf("expected root %q, got %q", filepath.Clean(dir), result.Root)
	}
	if !result.MarkerFound {
		t.Error("expected MarkerFound to be true")
	}
	if result.Marker != "go.mod" {
		t.Errorf("expected marker %q, got %q", "go.mod", result.Marker)
	}
}

func TestDiscoverRoot_NestedOneLevel_NonExplicitTargetCanAscend(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"go.mod":          "module example.com/test\n\ngo 1.21\n",
		"internal/app.go": "package internal\n",
	})

	nested := filepath.Join(dir, "internal")
	target := resolveTarget(t, nested)
	target.Explicit = false

	result := project.DiscoverRoot(target)

	if result.Root != filepath.Clean(dir) {
		t.Errorf("expected root %q, got %q", filepath.Clean(dir), result.Root)
	}
	if !result.MarkerFound {
		t.Error("expected MarkerFound to be true")
	}
}

func TestDiscoverRoot_DeeplyNested_NonExplicitTargetCanAscend(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"go.mod":                      "module example.com/test\n\ngo 1.21\n",
		"internal/service/handler.go": "package service\n",
	})

	deep := filepath.Join(dir, "internal", "service")
	target := resolveTarget(t, deep)
	target.Explicit = false

	result := project.DiscoverRoot(target)

	if result.Root != filepath.Clean(dir) {
		t.Errorf("expected root %q, got %q", filepath.Clean(dir), result.Root)
	}
	if !result.MarkerFound {
		t.Error("expected MarkerFound to be true")
	}
}

func TestDiscoverRoot_NoMarker_FallsBackToTarget(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Unknown project\n",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if result.MarkerFound {
		t.Error("expected MarkerFound to be false for project with no markers")
	}
	if result.Root != filepath.Clean(dir) {
		t.Errorf("expected fallback root %q, got %q", filepath.Clean(dir), result.Root)
	}
	if result.Marker != "" {
		t.Errorf("expected empty marker, got %q", result.Marker)
	}
}

func TestDiscoverRoot_GitDirectory(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		".git/config": "[core]\n\trepositoryformatversion = 0\n",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if !result.MarkerFound {
		t.Error("expected MarkerFound to be true for .git directory")
	}
	if result.Marker != ".git" {
		t.Errorf("expected marker %q, got %q", ".git", result.Marker)
	}
}

func TestDiscoverRoot_PackageJSON(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"package.json": "{\"name\":\"test\",\"version\":\"1.0.0\"}",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if !result.MarkerFound {
		t.Error("expected MarkerFound for package.json")
	}
	if result.Marker != "package.json" {
		t.Errorf("expected marker %q, got %q", "package.json", result.Marker)
	}
}

func TestDiscoverRoot_CargoToml(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"Cargo.toml": "[package]\nname = \"test\"\n",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if !result.MarkerFound {
		t.Error("expected MarkerFound for Cargo.toml")
	}
	if result.Marker != "Cargo.toml" {
		t.Errorf("expected marker %q, got %q", "Cargo.toml", result.Marker)
	}
}

func TestDiscoverRoot_RootAlwaysAbsolute(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/test\n",
	})

	target := resolveTarget(t, dir)
	result := project.DiscoverRoot(target)

	if !filepath.IsAbs(result.Root) {
		t.Errorf("expected Root to be absolute, got %q", result.Root)
	}
}

func TestDiscoverRoot_ExplicitTarget_DoesNotEscapeToAncestor(t *testing.T) {
	outer := testhelpers.TempProject(t, map[string]string{
		"go.mod":        "module outer\n\ngo 1.21\n",
		"inner/main.go": "package main\n",
	})

	inner := filepath.Join(outer, "inner")
	target := resolveTarget(t, inner)

	result := project.DiscoverRoot(target)

	if result.Root != filepath.Clean(inner) {
		t.Errorf("boundary violation: got root %q, want %q", result.Root, filepath.Clean(inner))
	}
	if result.Root == filepath.Clean(outer) {
		t.Fatalf("explicit target escaped to ancestor root %q", result.Root)
	}
}

func TestDiscoverRoot_ExplicitTarget_WithOwnMarker_StaysOnTarget(t *testing.T) {
	outer := testhelpers.TempProject(t, map[string]string{
		"go.mod":             "module outer\n\ngo 1.21\n",
		"inner/package.json": "{\"name\":\"inner\"}\n",
		"inner/index.js":     "console.log('hi')\n",
	})

	inner := filepath.Join(outer, "inner")
	target := resolveTarget(t, inner)

	result := project.DiscoverRoot(target)

	if result.Root != filepath.Clean(inner) {
		t.Errorf("expected root %q, got %q", filepath.Clean(inner), result.Root)
	}
	if !result.MarkerFound {
		t.Error("expected MarkerFound to be true")
	}
	if result.Marker != "package.json" {
		t.Errorf("expected marker %q, got %q", "package.json", result.Marker)
	}
}
