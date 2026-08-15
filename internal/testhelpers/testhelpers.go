// Package testhelpers provides shared utilities for Aryntra Aayam tests.
package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
)

// TempProject creates a temporary directory populated with the provided files.
// The directory is automatically removed when the test completes.
func TempProject(t *testing.T, files map[string]string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "Aryntra Aayam-test-*")
	if err != nil {
		t.Fatalf("testhelpers.TempProject: could not create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	for relPath, content := range files {
		fullPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("testhelpers.TempProject: could not create parent dir for %q: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("testhelpers.TempProject: could not write file %q: %v", relPath, err)
		}
	}

	return dir
}

// ProjectFixturePath returns the absolute path to a named fixture
// under testdata/projects/. Fails the test if the fixture does not exist.
func ProjectFixturePath(t *testing.T, name string) string {
	t.Helper()

	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("testhelpers.ProjectFixturePath: could not locate repository root: %v", err)
	}

	path := filepath.Join(root, "testdata", "projects", name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("testhelpers.ProjectFixturePath: fixture %q does not exist at %q", name, path)
	}

	return path
}

// findRepoRoot walks up from the current working directory to locate go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// AssertEqual fails the test if got != want.
// Provides a clear diff-style message showing expected and actual values.
func AssertEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if got != want {
		t.Errorf("AssertEqual:\n  want: %v\n   got: %v", want, got)
	}
}
