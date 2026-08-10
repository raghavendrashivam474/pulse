package project_test

import (
	"path/filepath"
	"testing"

	"pulse/internal/project"
	"pulse/internal/testhelpers"
)

func TestResolveTarget_ValidDirectory(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test",
	})

	target, err := project.ResolveTarget(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if target.Path != filepath.Clean(dir) {
		t.Errorf("expected path %q, got %q", filepath.Clean(dir), target.Path)
	}
}

func TestResolveTarget_MissingPath(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{})
	missing := filepath.Join(dir, "does-not-exist")

	_, err := project.ResolveTarget(missing)
	if err == nil {
		t.Fatal("expected an error for missing path, got nil")
	}
}

func TestResolveTarget_FileInsteadOfDirectory(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test",
	})

	filePath := filepath.Join(dir, "README.md")
	_, err := project.ResolveTarget(filePath)
	if err == nil {
		t.Fatal("expected an error when target is a file, got nil")
	}
}

func TestResolveTarget_EmptyPath(t *testing.T) {
	_, err := project.ResolveTarget("")
	if err == nil {
		t.Fatal("expected an error for empty path, got nil")
	}
}

func TestResolveTarget_AbsolutePath(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main",
	})

	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("could not make absolute path: %v", err)
	}

	target, err := project.ResolveTarget(abs)
	if err != nil {
		t.Fatalf("expected no error for absolute path, got: %v", err)
	}
	if target.Path == "" {
		t.Error("expected non-empty target path")
	}
}

func TestResolveTarget_PathIsClean(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"go.mod": "module example.com/test",
	})

	target, err := project.ResolveTarget(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Path != filepath.Clean(dir) {
		t.Errorf("path not cleaned: got %q, want %q", target.Path, filepath.Clean(dir))
	}
}

func TestResolveTarget_ErrorMessage_MissingPath(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{})
	missing := filepath.Join(dir, "no-such-dir")

	_, err := project.ResolveTarget(missing)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error message should be meaningful, not a raw Go error
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestResolveTarget_ErrorMessage_FileTarget(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main",
	})

	_, err := project.ResolveTarget(filepath.Join(dir, "main.go"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}
