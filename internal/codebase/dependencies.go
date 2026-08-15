package codebase

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/raghavendrashivam474/aayam/internal/scanner"
)

// DependencyType classifies the kind of relationship between two packages.
type DependencyType string

const (
	// DependencyImport represents a Go import relationship.
	DependencyImport DependencyType = "import"
)

// Dependency represents a directed relationship between two packages
// within the analysed project.
//
// Future graph role: Edge.
type Dependency struct {
	// From is the source package name.
	From string

	// To is the target package name.
	To string

	// Type classifies the relationship (currently always "import").
	Type DependencyType
}

// DiscoverDependencies analyses Go source files to find import relationships
// between packages that are internal to the project.
//
// The function:
//   - Reads go.mod to determine the module path.
//   - Parses each Go file for its import declarations.
//   - Filters to only internal imports (those starting with the module path).
//   - Resolves import paths to package names using the discovered packages.
//   - Deduplicates edges.
//   - Returns a deterministically ordered slice.
//
// Standard library and external module imports are excluded.
// Malformed files are skipped without error.
// Cyclic dependencies are represented without crashing.
func DiscoverDependencies(root string, inv scanner.Inventory, packages []Package) []Dependency {
	modulePath := readModulePath(root)
	if modulePath == "" {
		// Not a Go module: no dependencies to discover.
		return nil
	}

	// Build a lookup from import path suffix → package name.
	// For example: "github.com/raghavendrashivam474/aayam/internal/cli" → "cli"
	// The suffix is the part after the module path.
	importToPackage := buildImportLookup(modulePath, packages)

	// Track unique edges.
	type edge struct {
		from string
		to   string
	}
	seen := make(map[edge]bool)
	var deps []Dependency

	fset := token.NewFileSet()

	for _, entry := range inv.Files {
		if entry.Extension != ".go" {
			continue
		}

		absPath := filepath.Join(root, filepath.FromSlash(entry.RelPath))

		src, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		f, err := parser.ParseFile(fset, absPath, src, parser.ImportsOnly)
		if err != nil {
			continue
		}

		if f.Name == nil {
			continue
		}

		fromPkg := f.Name.Name

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Only consider imports within this module.
			if !strings.HasPrefix(importPath, modulePath) {
				continue
			}

			toPkg, ok := importToPackage[importPath]
			if !ok {
				continue
			}

			// Don't create self-edges.
			if fromPkg == toPkg {
				continue
			}

			e := edge{from: fromPkg, to: toPkg}
			if seen[e] {
				continue
			}
			seen[e] = true

			deps = append(deps, Dependency{
				From: fromPkg,
				To:   toPkg,
				Type: DependencyImport,
			})
		}
	}

	// Deterministic ordering: by (From, To).
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].From != deps[j].From {
			return deps[i].From < deps[j].From
		}
		return deps[i].To < deps[j].To
	})

	return deps
}

// buildImportLookup creates a map from full Go import path to package name.
//
// For example, given module "Aryntra Aayam" and a package with Path "internal/cli"
// and Name "cli", the entry is:
//
//	"github.com/raghavendrashivam474/aayam/internal/cli" → "cli"
//
// Root-level packages (Path "") map to: "Aryntra Aayam" → "main"
func buildImportLookup(modulePath string, packages []Package) map[string]string {
	lookup := make(map[string]string)

	for _, pkg := range packages {
		var importPath string
		if pkg.Path == "" {
			importPath = modulePath
		} else {
			importPath = modulePath + "/" + pkg.Path
		}
		// If two packages share a directory (e.g. cli and cli_test),
		// the non-test package wins for import resolution since
		// external test packages cannot be imported.
		if existing, ok := lookup[importPath]; ok {
			if IsTestPackage(existing) && !IsTestPackage(pkg.Name) {
				lookup[importPath] = pkg.Name
			}
		} else {
			lookup[importPath] = pkg.Name
		}
	}

	return lookup
}

// readModulePath reads the module directive from go.mod in root.
// Returns empty string if go.mod doesn't exist or can't be parsed.
func readModulePath(root string) string {
	modPath := filepath.Join(root, "go.mod")

	f, err := os.Open(modPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}
