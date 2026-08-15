package codebase

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/scanner"
)

// Package represents a discovered package within the codebase.
//
// Future graph role: Node.
//
// A Package is defined by its source-declared name, not by directory
// naming conventions. For Go, the package declaration in source files
// is authoritative.
type Package struct {
	// Name is the declared package name (e.g. "cli", "main", "cli_test").
	Name string

	// Path is the directory path (forward-slash, relative to project root)
	// that contains this package's files.
	Path string

	// Language identifies the ecosystem this package belongs to.
	Language project.Language

	// Files lists the relative paths of source files belonging to this package.
	// Sorted alphabetically for deterministic ordering.
	Files []string
}

// DiscoverPackages analyses the project files and returns all detected
// packages.
//
// Currently only Go is implemented. The function dispatches by language:
// Go files are parsed for their package declaration; other languages
// return no packages without error.
//
// The returned slice is sorted by (Path, Name) for deterministic ordering.
func DiscoverPackages(root string, inv scanner.Inventory) []Package {
	return discoverGoPackages(root, inv)
}

// discoverGoPackages inspects Go source files to determine their declared
// package name, then groups them into Package values.
//
// Implementation uses go/parser to read the package clause rather than
// deriving it from directory names. This preserves the distinction
// between "package cli" and "package cli_test" in the same directory.
func discoverGoPackages(root string, inv scanner.Inventory) []Package {
	// Key: "dirPath:packageName" → list of file relative paths
	type pkgKey struct {
		dir  string
		name string
	}

	groups := make(map[pkgKey][]string)

	fset := token.NewFileSet()

	for _, entry := range inv.Files {
		if entry.Extension != ".go" {
			continue
		}

		absPath := filepath.Join(root, filepath.FromSlash(entry.RelPath))

		// Parse only the package clause — no need for the full AST.
		src, err := os.ReadFile(absPath)
		if err != nil {
			// Unreadable file: skip rather than abort.
			continue
		}

		f, err := parser.ParseFile(fset, absPath, src, parser.PackageClauseOnly)
		if err != nil {
			// Malformed Go file: skip rather than abort.
			continue
		}

		if f.Name == nil {
			continue
		}

		pkgName := f.Name.Name
		dirPath := filepath.ToSlash(filepath.Dir(entry.RelPath))
		if dirPath == "." {
			dirPath = ""
		}

		key := pkgKey{dir: dirPath, name: pkgName}
		groups[key] = append(groups[key], entry.RelPath)
	}

	// Convert map to sorted slice.
	packages := make([]Package, 0, len(groups))
	for key, files := range groups {
		sort.Strings(files)
		packages = append(packages, Package{
			Name:     key.name,
			Path:     key.dir,
			Language: project.LangGo,
			Files:    files,
		})
	}

	// Sort by (Path, Name) for deterministic output.
	sort.Slice(packages, func(i, j int) bool {
		if packages[i].Path != packages[j].Path {
			return packages[i].Path < packages[j].Path
		}
		return packages[i].Name < packages[j].Name
	})

	return packages
}

// PackageNames returns a sorted, deduplicated list of package names.
// This is useful for summary output without exposing internal structure.
func PackageNames(packages []Package) []string {
	seen := make(map[string]bool)
	var names []string

	for _, p := range packages {
		if !seen[p.Name] {
			seen[p.Name] = true
			names = append(names, p.Name)
		}
	}

	sort.Strings(names)
	return names
}

// IsTestPackage reports whether the package name follows Go's external
// test package convention (suffix "_test").
func IsTestPackage(name string) bool {
	return strings.HasSuffix(name, "_test")
}
