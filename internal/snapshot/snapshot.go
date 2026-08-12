// Package snapshot provides the unified project state representation.
//
// A ProjectSnapshot captures everything Pulse has discovered about a
// project in a single, immutable, deterministic value. Downstream
// consumers (terminal output, JSON output, future engines) read from
// the snapshot rather than re-running discovery.
//
// Construction: call Discover with a resolved configuration.
package snapshot

import (
	"pulse/internal/codebase"
	"pulse/internal/git"
	"pulse/internal/project"
	"pulse/internal/scanner"
)

// ProjectSnapshot is the canonical representation of discovered project state.
//
// Design constraints:
//   - No filesystem handles or transient runtime state.
//   - All collections are deterministically ordered.
//   - No output-specific formatting; this is a domain model.
//   - Easy to serialize (JSON, future formats).
//   - Easy to test via value comparison.
type ProjectSnapshot struct {
	// Name is the project name (derived from root directory name).
	Name string

	// Root is the absolute path to the project root.
	Root string

	// Type is the detected primary project ecosystem.
	Type project.ProjectType

	// Languages is the sorted, deduplicated set of detected languages.
	Languages []project.Language

	// Files contains all discovered file entries from the scanner.
	Files []scanner.FileEntry

	// Directories contains all discovered directory entries from the scanner.
	Directories []scanner.DirEntry

	// FileCount is the number of scanned files.
	FileCount int

	// DirectoryCount is the number of directories under root (root excluded).
	DirectoryCount int

	// Codebase holds the structured code model: files, directories,
	// packages, and their dependency relationships.
	// This is the S4 deliverable that S5 will consume for graph construction.
	Codebase codebase.Codebase

	// Git holds the Git intelligence for this project.
	// Git.IsRepository is false when the target is not inside a Git
	// repository -- that is a valid, expected state.
	Git git.GitInfo
}

// Discover runs the complete project discovery pipeline and returns
// a ProjectSnapshot.
//
// Pipeline:
//
//	Config -> Target -> Root -> Scan -> Detect -> Metadata -> Git -> Codebase -> Snapshot
//
// This is the single orchestration point. No caller should need to
// run the individual steps manually unless they have a specific reason.
func Discover(targetPath string) (ProjectSnapshot, error) {
	// M2.1: Validate the target path.
	resolvedTarget, err := project.ResolveTarget(targetPath)
	if err != nil {
		return ProjectSnapshot{}, err
	}

	// M2.2: Discover the project root.
	rootResult := project.DiscoverRoot(resolvedTarget)

	// M2.3: Scan the filesystem.
	inv, err := scanner.Scan(rootResult.Root)
	if err != nil {
		return ProjectSnapshot{}, err
	}

	// M2.4: Detect project type.
	detection := project.DetectType(inv)

	// M2.5 + M2.6: Build metadata (includes language detection).
	meta := project.NewMetadata(rootResult.Root, detection, inv)

	// S3: Git intelligence -- run against the resolved target so the
	// Git package identifies the repository root independently of the
	// project root.
	gitInfo, err := git.Discover(resolvedTarget.Path)
	if err != nil {
		return ProjectSnapshot{}, err
	}

	// S4: Codebase structure intelligence.
	cb := codebase.Discover(rootResult.Root, inv)

	return ProjectSnapshot{
		Name:           meta.Name,
		Root:           meta.Root,
		Type:           meta.Type,
		Languages:      meta.Languages,
		Files:          inv.Files,
		Directories:    inv.Dirs,
		FileCount:      meta.FileCount,
		DirectoryCount: meta.DirectoryCount,
		Codebase:       cb,
		Git:            gitInfo,
	}, nil
}
