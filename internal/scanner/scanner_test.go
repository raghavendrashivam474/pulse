package scanner_test

import (
	"path/filepath"
	"testing"

	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

func TestScan_EmptyProject(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(inv.Files))
	}
	if len(inv.Dirs) != 0 {
		t.Errorf("expected 0 dirs, got %d", len(inv.Dirs))
	}
}

func TestScan_SingleFile(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Hello\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(inv.Files))
	}
	if inv.Files[0].RelPath != "README.md" {
		t.Errorf("expected RelPath %q, got %q", "README.md", inv.Files[0].RelPath)
	}
	if inv.Files[0].Extension != ".md" {
		t.Errorf("expected extension %q, got %q", ".md", inv.Files[0].Extension)
	}
}

func TestScan_NestedDirectories(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go":             "package main\n",
		"internal/app.go":     "package internal\n",
		"internal/sub/sub.go": "package sub\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(inv.Files), inv.Files)
	}
	if len(inv.Dirs) != 2 {
		t.Errorf("expected 2 dirs (internal, internal/sub), got %d: %v", len(inv.Dirs), inv.Dirs)
	}
}

func TestScan_MultipleFileTypes(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go":     "package main\n",
		"README.md":   "# Readme\n",
		"go.mod":      "module example.com/test\n",
		"config.yaml": "key: value\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Files) != 4 {
		t.Errorf("expected 4 files, got %d", len(inv.Files))
	}

	extMap := make(map[string]bool)
	for _, f := range inv.Files {
		extMap[f.Extension] = true
	}

	for _, expected := range []string{".go", ".md", ".yaml"} {
		if !extMap[expected] {
			t.Errorf("expected extension %q in inventory", expected)
		}
	}
	// go.mod has no extension
	if !extMap[""] {
		t.Error("expected empty extension for go.mod")
	}
}

func TestScan_ExcludedDirectories(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go":                      "package main\n",
		"node_modules/lodash/index.js": "module.exports = {};\n",
		"vendor/pkg/pkg.go":            "package pkg\n",
		".git/config":                  "[core]\n",
		"dist/bundle.js":               "// bundle\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Files) != 1 {
		t.Errorf("expected 1 file (excluding excluded dirs), got %d: %v", len(inv.Files), inv.Files)
	}
	if inv.Files[0].RelPath != "main.go" {
		t.Errorf("expected main.go, got %q", inv.Files[0].RelPath)
	}

	excluded := map[string]bool{
		"node_modules": true, "vendor": true, ".git": true, "dist": true,
	}
	for _, d := range inv.Dirs {
		if excluded[d.RelPath] {
			t.Errorf("excluded directory %q appeared in inventory", d.RelPath)
		}
	}
}

func TestScan_DeterministicOrdering(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"z_last.go":   "package main\n",
		"a_first.go":  "package main\n",
		"m_middle.go": "package main\n",
	})

	inv1, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("scan 1 error: %v", err)
	}
	inv2, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("scan 2 error: %v", err)
	}

	if len(inv1.Files) != len(inv2.Files) {
		t.Fatalf("scans returned different file counts: %d vs %d", len(inv1.Files), len(inv2.Files))
	}
	for i := range inv1.Files {
		if inv1.Files[i].RelPath != inv2.Files[i].RelPath {
			t.Errorf("non-deterministic ordering at index %d: %q vs %q",
				i, inv1.Files[i].RelPath, inv2.Files[i].RelPath)
		}
	}

	expected := []string{"a_first.go", "m_middle.go", "z_last.go"}
	for i, e := range expected {
		if inv1.Files[i].RelPath != e {
			t.Errorf("expected file[%d] = %q, got %q", i, e, inv1.Files[i].RelPath)
		}
	}
}

func TestScan_ForwardSlashPaths(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"internal/service/handler.go": "package service\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, f := range inv.Files {
		for _, ch := range f.RelPath {
			if ch == '\\' {
				t.Errorf("RelPath contains backslash: %q", f.RelPath)
			}
		}
	}
	for _, d := range inv.Dirs {
		for _, ch := range d.RelPath {
			if ch == '\\' {
				t.Errorf("Dir RelPath contains backslash: %q", d.RelPath)
			}
		}
	}
}

func TestScan_FileWithNoExtension(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"Makefile": "all:\n\techo done\n",
		"go.mod":   "module example.com/test\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundEmpty := false
	for _, f := range inv.Files {
		if f.Extension == "" {
			foundEmpty = true
		}
	}
	if !foundEmpty {
		t.Error("expected at least one file with empty extension")
	}
}

func TestScan_RootNotInDirs(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, d := range inv.Dirs {
		if d.RelPath == "." || d.RelPath == "" {
			t.Errorf("root appeared in Dirs: %q", d.RelPath)
		}
	}
}

func TestScan_SizePopulated(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"hello.go": "package main\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(inv.Files))
	}
	if inv.Files[0].SizeBytes == 0 {
		t.Error("expected non-zero SizeBytes for non-empty file")
	}
}

func TestScan_ExtensionIsLowercase(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"internal/app.go": "package internal\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, f := range inv.Files {
		for _, ch := range f.Extension {
			if ch >= 'A' && ch <= 'Z' {
				t.Errorf("extension not lowercase: %q in file %q", f.Extension, f.RelPath)
			}
		}
	}
}

func TestScan_NestedDirPaths(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"a/b/c/deep.go": "package deep\n",
	})

	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have dirs: a, a/b, a/b/c
	if len(inv.Dirs) != 3 {
		t.Errorf("expected 3 dirs, got %d: %v", len(inv.Dirs), inv.Dirs)
	}

	expectedDirs := []string{"a", "a/b", "a/b/c"}
	for i, expected := range expectedDirs {
		got := filepath.ToSlash(inv.Dirs[i].RelPath)
		if got != expected {
			t.Errorf("dirs[%d]: expected %q, got %q", i, expected, got)
		}
	}
}
