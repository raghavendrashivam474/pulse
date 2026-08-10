package scanner

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// FileEntry represents a single file discovered during a scan.
type FileEntry struct {
	// RelPath is the path relative to the scan root, using forward slashes.
	RelPath string

	// Extension is the lowercase file extension including the leading dot.
	// Empty string for files with no extension (e.g. Makefile, go.mod).
	Extension string

	// SizeBytes is the size of the file in bytes at scan time.
	SizeBytes int64
}

// DirEntry represents a single directory discovered during a scan.
type DirEntry struct {
	// RelPath is the path relative to the scan root, using forward slashes.
	RelPath string
}

// Inventory is the complete result of scanning a project root.
// All slices are sorted by RelPath for deterministic ordering.
type Inventory struct {
	// Files contains all discovered files, sorted by RelPath.
	Files []FileEntry

	// Dirs contains all discovered directories, sorted by RelPath.
	// The root itself is not included.
	Dirs []DirEntry
}

// defaultExclusions lists directory names that are never scanned.
// These are generated artefacts, dependency stores, or VCS internals
// that are not part of the project source.
var defaultExclusions = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"vendor":       true,
	".cache":       true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
}

// Scan performs a deterministic filesystem inventory of root.
//
// Behaviour:
//   - Only regular files and directories are recorded.
//   - Symlinks are not followed. Symlinked entries are silently skipped.
//   - Excluded directories are skipped entirely, their contents not visited.
//   - Results are sorted by RelPath for deterministic ordering.
//   - The root directory itself does not appear in Dirs.
//   - Unreadable entries are skipped rather than aborting the scan.
//
// The returned Inventory is always valid even for an empty project.
func Scan(root string) (Inventory, error) {
	var files []FileEntry
	var dirs []DirEntry

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Cannot access this entry. Skip it rather than aborting
			// the entire scan. A partial inventory is more useful than none.
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}

		// Normalise separators to forward slash for cross-platform consistency.
		rel = filepath.ToSlash(rel)

		// The root itself: skip recording, continue walking.
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			// Check exclusions by directory name (not full path).
			if defaultExclusions[d.Name()] {
				return filepath.SkipDir
			}
			dirs = append(dirs, DirEntry{RelPath: rel})
			return nil
		}

		// Skip symlinks. Do not follow them — we cannot know where they lead
		// and recursing outside the project root would be incorrect.
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		// Record the regular file.
		info, infoErr := d.Info()
		var size int64
		if infoErr == nil {
			size = info.Size()
		}

		files = append(files, FileEntry{
			RelPath:   rel,
			Extension: strings.ToLower(filepath.Ext(d.Name())),
			SizeBytes: size,
		})

		return nil
	})

	if walkErr != nil {
		return Inventory{}, walkErr
	}

	// Sort both slices by RelPath.
	// Critical requirement: tests and snapshot comparisons depend on this.
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].RelPath < dirs[j].RelPath
	})

	return Inventory{
		Files: files,
		Dirs:  dirs,
	}, nil
}
