package project

import (
	"path/filepath"

	"pulse/internal/scanner"
)

// Metadata holds the basic identity and structural facts about a project.
// It is constructed once from the results of discovery and is immutable
// after creation.
//
// Ownership rule: all calculations happen here.
// The output package must only render values that already exist on Metadata.
type Metadata struct {
	// Name is the project name derived from the root directory name.
	// Future versions may parse go.mod / package.json for a canonical name.
	Name string

	// Root is the absolute path to the project root directory.
	Root string

	// Type is the detected primary project ecosystem.
	Type ProjectType

	// Languages is the sorted, deduplicated set of languages detected
	// from the project file extensions.
	Languages []Language

	// FileCount is the number of files included by the scanner.
	// Excluded directories do not contribute to this count.
	FileCount int

	// DirectoryCount is the number of directories under the project root,
	// excluding the root itself.
	// Convention: root is not counted. "internal/" counts as 1.
	DirectoryCount int
}

// NewMetadata constructs a Metadata value from the outputs of the
// discovery pipeline.
//
// Parameters:
//   - root: absolute project root path (from RootDiscovery.Root)
//   - detection: project type result (from DetectType)
//   - inv: filesystem inventory (from scanner.Scan)
//
// Language detection is performed here from the inventory file list so
// that no second filesystem walk is required.
func NewMetadata(root string, detection Detection, inv scanner.Inventory) Metadata {
	// Derive name from the final path segment of the root directory.
	name := filepath.Base(root)

	// Build the file path slice that DetectLanguages expects.
	paths := make([]string, len(inv.Files))
	for i, f := range inv.Files {
		paths[i] = f.RelPath
	}

	return Metadata{
		Name:           name,
		Root:           root,
		Type:           detection.Primary,
		Languages:      DetectLanguages(paths),
		FileCount:      len(inv.Files),
		DirectoryCount: len(inv.Dirs),
	}
}
