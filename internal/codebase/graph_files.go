package codebase

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"pulse/internal/scanner"
)

// FileNodeID returns the stable graph node ID for a file path.
//
// Convention: "file:<relative/forward/slash/path>"
//
// The path must be relative to the project root and use forward slashes.
// Absolute paths and machine-specific paths must never be used as IDs.
func FileNodeID(relPath string) string {
	return "file:" + filepath.ToSlash(relPath)
}

// DirNodeID returns the stable graph node ID for a directory path.
//
// Convention: "dir:<relative/forward/slash/path>"
//
// The root directory itself uses "dir:.".
func DirNodeID(relPath string) string {
	return "dir:" + filepath.ToSlash(relPath)
}

// ProjectNodeID returns the stable graph node ID for a project.
//
// Convention: "project:<name>"
func ProjectNodeID(name string) string {
	return "project:" + name
}

// BuildFileGraph constructs a CodeGraph capturing file-level relationships.
//
// Relationships produced:
//
//	directory  --contains-->   file
//	file       --belongs_to--> package   (Go only, from package declaration)
//	file       --imports-->    package   (Go only, from import declarations)
//
// Source of truth:
//   - File paths come from the existing Codebase.Files list (no re-scan).
//   - Package declarations are re-parsed from source using go/parser.
//   - Import declarations are re-parsed from source using go/parser.
//
// Unresolvable imports (stdlib, external modules) are silently skipped.
// Malformed files are silently skipped.
// Windows absolute paths are never used as node IDs.
//
// The returned graph is normalised (deterministic ordering).
func BuildFileGraph(root string, cb Codebase, inv scanner.Inventory) CodeGraph {
	g := NewCodeGraph()

	// Build a lookup from package name -> Node ID so we can wire
	// file->package edges without duplicating ID logic.
	pkgNodeIDs := make(map[string]string) // packageName -> nodeID
	for _, pkg := range cb.Packages {
		id := PackageNodeID(pkg.Name)
		pkgNodeIDs[pkg.Name] = id
		g.AddNode(Node{
			ID:   id,
			Kind: NodePackage,
			Name: pkg.Name,
			Path: pkg.Path,
		})
	}

	// Determine module path for resolving internal imports.
	modulePath := readModulePath(root)

	// Build import-path -> package name lookup (same logic as dependencies.go).
	importToPackage := buildImportLookup(modulePath, cb.Packages)

	fset := token.NewFileSet()

	for _, f := range cb.Files {
		// Add a node for the file itself.
		fileID := FileNodeID(f.Path)
		g.AddNode(Node{
			ID:   fileID,
			Kind: NodeFile,
			Name: f.Name,
			Path: f.Path,
		})

		// Add a node for the file's parent directory and a contains edge.
		dirRel := parentDir(f.Path)
		dirID := DirNodeID(dirRel)
		g.AddNode(Node{
			ID:   dirID,
			Kind: NodeDirectory,
			Name: dirName(dirRel),
			Path: dirRel,
		})
		g.AddEdge(Relationship{
			From: dirID,
			To:   fileID,
			Kind: RelationshipContains,
		})

		// Go-specific relationships require parsing.
		if f.Extension != ".go" {
			continue
		}

		absPath := filepath.Join(root, filepath.FromSlash(f.Path))
		src, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		parsed, err := parser.ParseFile(fset, absPath, src, parser.ImportsOnly)
		if err != nil {
			continue
		}
		if parsed.Name == nil {
			continue
		}

		declaredPkg := parsed.Name.Name

		// belongs_to: file -> package (from package declaration).
		if pkgID, ok := pkgNodeIDs[declaredPkg]; ok {
			g.AddEdge(Relationship{
				From: fileID,
				To:   pkgID,
				Kind: RelationshipBelongsTo,
			})
		}

		// imports: file -> package (from import declarations).
		if modulePath == "" {
			continue
		}
		for _, imp := range parsed.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Only internal imports — those within this module.
			if !strings.HasPrefix(importPath, modulePath) {
				continue
			}

			toPkg, ok := importToPackage[importPath]
			if !ok {
				continue
			}

			// Never create a self-import edge.
			if toPkg == declaredPkg {
				continue
			}

			pkgID, ok := pkgNodeIDs[toPkg]
			if !ok {
				continue
			}

			g.AddEdge(Relationship{
				From: fileID,
				To:   pkgID,
				Kind: RelationshipImports,
			})
		}
	}

	g.Normalise()
	return g
}

// parentDir returns the forward-slash parent directory of a relative file path.
// Files in the root return ".".
func parentDir(relPath string) string {
	slashed := filepath.ToSlash(relPath)
	if i := strings.LastIndex(slashed, "/"); i >= 0 {
		return slashed[:i]
	}
	return "."
}
