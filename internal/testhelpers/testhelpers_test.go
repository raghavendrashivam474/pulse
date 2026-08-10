package testhelpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"pulse/internal/testhelpers"
)

func TestTempProject_CreatesDirectory(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{})
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected temp directory to exist: %q", dir)
	}
}

func TestTempProject_CreatesFiles(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Hello",
		"main.go":   "package main",
	})
	if _, err := os.Stat(filepath.Join(dir, "README.md")); os.IsNotExist(err) {
		t.Errorf("expected README.md to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); os.IsNotExist(err) {
		t.Errorf("expected main.go to exist")
	}
}

func TestTempProject_FileContent(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"hello.txt": "hello world",
	})
	content, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", string(content))
	}
}

func TestTempProject_CreatesNestedDirectories(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"sub/dir/file.txt": "nested",
	})
	path := filepath.Join(dir, "sub", "dir", "file.txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected nested file to exist: %q", path)
	}
}

func TestProjectFixturePath_Basic(t *testing.T) {
	path := testhelpers.ProjectFixturePath(t, "basic")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected basic fixture to exist at %q", path)
	}
}

func TestProjectFixturePath_IsAbsolute(t *testing.T) {
	path := testhelpers.ProjectFixturePath(t, "basic")
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
}
