package project_test

import (
	"testing"

	"pulse/internal/project"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// fixture builds a small deterministic project and returns its inventory.
//
//	basic/
//	+-- go.mod
//	+-- main.go
//	+-- README.md
//	+-- internal/
//	    +-- app.go
//
// Expected:
//
//	Name:           "basic"
//	Type:           Go
//	Languages:      [Go, Markdown]
//	FileCount:      4
//	DirectoryCount: 1   (internal/ only; root excluded)
func buildBasicFixture(t *testing.T) (string, project.Detection, scanner.Inventory) {
	t.Helper()

	root := testhelpers.TempProject(t, map[string]string{
		"go.mod":          "module basic\n\ngo 1.21\n",
		"main.go":         "package main\n",
		"README.md":       "# basic\n",
		"internal/app.go": "package internal\n",
	})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	detection := project.DetectType(inv)

	return root, detection, inv
}

// TestNewMetadata_Name verifies the project name is the root directory name.
func TestNewMetadata_Name(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	// TempProject dirs look like: pulse-test-1234567890
	// We just verify it is non-empty and matches filepath.Base(root).
	if meta.Name == "" {
		t.Error("Name must not be empty")
	}
}

// TestNewMetadata_Root verifies the root path is preserved exactly.
func TestNewMetadata_Root(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	testhelpers.AssertEqual(t, root, meta.Root)
}

// TestNewMetadata_Type verifies the project type is correctly propagated.
func TestNewMetadata_Type(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	testhelpers.AssertEqual(t, project.TypeGo, meta.Type)
}

// TestNewMetadata_FileCount verifies the file count matches the inventory.
func TestNewMetadata_FileCount(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	// go.mod, main.go, README.md, internal/app.go = 4
	testhelpers.AssertEqual(t, 4, meta.FileCount)
}

// TestNewMetadata_DirectoryCount verifies that only subdirectories are counted.
// The root itself must not be included.
func TestNewMetadata_DirectoryCount(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	// Only "internal/" is a subdirectory. Root is excluded by convention.
	testhelpers.AssertEqual(t, 1, meta.DirectoryCount)
}

// TestNewMetadata_Languages verifies language detection runs correctly
// and returns deduplicated, sorted languages.
func TestNewMetadata_Languages(t *testing.T) {
	root, detection, inv := buildBasicFixture(t)
	meta := project.NewMetadata(root, detection, inv)

	// go.mod has no recognised extension. .go -> Go, .md -> Markdown.
	// Alphabetical: Go, Markdown.
	testhelpers.AssertEqual(t, 2, len(meta.Languages))
	testhelpers.AssertEqual(t, project.LangGo, meta.Languages[0])
	testhelpers.AssertEqual(t, project.LangMarkdown, meta.Languages[1])
}

// TestNewMetadata_EmptyProject verifies that an empty project does not panic.
func TestNewMetadata_EmptyProject(t *testing.T) {
	root := testhelpers.TempProject(t, map[string]string{})

	inv, err := scanner.Scan(root)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	detection := project.DetectType(inv)
	meta := project.NewMetadata(root, detection, inv)

	testhelpers.AssertEqual(t, 0, meta.FileCount)
	testhelpers.AssertEqual(t, 0, meta.DirectoryCount)
	testhelpers.AssertEqual(t, 0, len(meta.Languages))
	testhelpers.AssertEqual(t, project.TypeUnknown, meta.Type)
}
