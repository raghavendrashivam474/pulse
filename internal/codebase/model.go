package codebase

import (
	"pulse/internal/scanner"
)

// Codebase is the complete structural model of a project's source code.
//
// It aggregates files, directories, packages, and their dependency
// relationships into a single, immutable, deterministic value.
//
// This is the primary S4 deliverable. Downstream consumers (terminal
// output, JSON output, future graph engine) consume this model rather
// than re-scanning or re-parsing.
//
// Architecture:
//
//	Filesystem / Go source
//	        │
//	        ▼
//	 Codebase Model  ← this type
//	        │
//	        ├── Terminal output
//	        ├── JSON output
//	        └── Future graph (S5)
type Codebase struct {
	// Files contains all source artifacts, sorted by Path.
	Files []File

	// Directories contains all filesystem containers, sorted by Path.
	Directories []Directory

	// Packages contains all detected packages, sorted by (Path, Name).
	Packages []Package

	// Dependencies contains all inter-package relationships, sorted by (From, To).
	Dependencies []Dependency
}

// Discover builds a complete Codebase model from a project root and
// its filesystem inventory.
//
// Pipeline:
//
//	Inventory → Files + Directories
//	            → Packages (Go parser)
//	            → Dependencies (Go imports)
//	            → Codebase
//
// This is the single orchestration point for codebase analysis.
func Discover(root string, inv scanner.Inventory) Codebase {
	files := BuildFiles(inv.Files)
	dirs := BuildDirectories(inv.Dirs)
	packages := DiscoverPackages(root, inv)
	dependencies := DiscoverDependencies(root, inv, packages)

	return Codebase{
		Files:        files,
		Directories:  dirs,
		Packages:     packages,
		Dependencies: dependencies,
	}
}
