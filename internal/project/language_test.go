package project_test

import (
	"testing"

	"pulse/internal/project"
	"pulse/internal/testhelpers"
)

// TestDetectLanguages_SingleLanguage verifies that a project containing
// only Go files reports exactly one language.
func TestDetectLanguages_SingleLanguage(t *testing.T) {
	files := []string{
		"main.go",
		"internal/app.go",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 1, len(langs))
	testhelpers.AssertEqual(t, project.LangGo, langs[0])
}

// TestDetectLanguages_MultipleLanguages verifies that a project with
// several recognised extensions returns all unique languages.
func TestDetectLanguages_MultipleLanguages(t *testing.T) {
	files := []string{
		"main.go",
		"README.md",
		"build.ps1",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 3, len(langs))
	// Alphabetical: Go, Markdown, PowerShell
	testhelpers.AssertEqual(t, project.LangGo, langs[0])
	testhelpers.AssertEqual(t, project.LangMarkdown, langs[1])
	testhelpers.AssertEqual(t, project.LangPowerShell, langs[2])
}

// TestDetectLanguages_DeduplicatesLanguages verifies that multiple files
// sharing the same language produce a single language entry.
func TestDetectLanguages_DeduplicatesLanguages(t *testing.T) {
	files := []string{
		"main.ts",
		"app.ts",
		"router.tsx",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 1, len(langs))
	testhelpers.AssertEqual(t, project.LangTypeScript, langs[0])
}

// TestDetectLanguages_UnknownExtensionsIgnored verifies that files with
// unrecognised extensions do not cause errors and are silently skipped.
func TestDetectLanguages_UnknownExtensionsIgnored(t *testing.T) {
	files := []string{
		"config.xyz",
		"data.bin",
		"archive.tar",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 0, len(langs))
}

// TestDetectLanguages_MixedKnownAndUnknown verifies that unknown
// extensions do not suppress the languages that are recognised.
func TestDetectLanguages_MixedKnownAndUnknown(t *testing.T) {
	files := []string{
		"main.go",
		"config.xyz",
		"README.md",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 2, len(langs))
	testhelpers.AssertEqual(t, project.LangGo, langs[0])
	testhelpers.AssertEqual(t, project.LangMarkdown, langs[1])
}

// TestDetectLanguages_DeterministicOrdering verifies that two calls with
// the same input — delivered in different orders — produce identical output.
func TestDetectLanguages_DeterministicOrdering(t *testing.T) {
	filesA := []string{"main.go", "README.md", "build.ps1"}
	filesB := []string{"build.ps1", "README.md", "main.go"}

	langsA := project.DetectLanguages(filesA)
	langsB := project.DetectLanguages(filesB)

	testhelpers.AssertEqual(t, len(langsA), len(langsB))
	for i := range langsA {
		testhelpers.AssertEqual(t, langsA[i], langsB[i])
	}
}

// TestDetectLanguages_NestedFiles verifies that files in subdirectories
// are recognised correctly.
func TestDetectLanguages_NestedFiles(t *testing.T) {
	files := []string{
		"cmd/pulse/main.go",
		"internal/project/language.go",
		"docs/guide.md",
	}

	langs := project.DetectLanguages(files)

	testhelpers.AssertEqual(t, 2, len(langs))
	testhelpers.AssertEqual(t, project.LangGo, langs[0])
	testhelpers.AssertEqual(t, project.LangMarkdown, langs[1])
}

// TestDetectLanguages_EmptyInput verifies that an empty file list
// produces an empty language list without panicking.
func TestDetectLanguages_EmptyInput(t *testing.T) {
	langs := project.DetectLanguages([]string{})

	testhelpers.AssertEqual(t, 0, len(langs))
}

// TestDetectLanguages_CAndCPPDistinct verifies that C and C++ are
// reported as distinct languages when both are present.
func TestDetectLanguages_CAndCPPDistinct(t *testing.T) {
	files := []string{
		"lib.c",
		"engine.cpp",
		"header.h",
	}

	langs := project.DetectLanguages(files)

	// Both C and C++ should appear; alphabetically C comes before C++
	testhelpers.AssertEqual(t, 2, len(langs))
	testhelpers.AssertEqual(t, project.LangC, langs[0])
	testhelpers.AssertEqual(t, project.LangCPP, langs[1])
}
