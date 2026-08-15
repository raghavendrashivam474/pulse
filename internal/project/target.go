package project

import (
	"fmt"
	"os"
	"path/filepath"

	aayamErrors "github.com/raghavendrashivam474/aayam/internal/errors"
)

// Target represents a validated analysis target.
// A Target is guaranteed to exist, be accessible, and be a directory.
type Target struct {
	// Path is the absolute, cleaned path to the analysis target directory.
	Path string

	// Explicit reports whether the user explicitly supplied this path.
	// When true, discovery must not walk above Path.
	Explicit bool
}

// ResolveTarget validates that the given path is suitable for analysis.
//
// The path must:
//   - not be empty
//   - exist on the filesystem
//   - be a directory
//
// The path received here should already be absolute and cleaned by Config.
// ResolveTarget adds semantic validation on top.
func ResolveTarget(path string) (Target, error) {
	if path == "" {
		return Target{}, aayamErrors.User("target path must not be empty")
	}

	info, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return Target{}, aayamErrors.User(
				fmt.Sprintf("target path does not exist: %s", path),
			)
		}

		return Target{}, aayamErrors.Environment(
			fmt.Sprintf("target path cannot be accessed: %s", path),
			statErr,
		)
	}

	if !info.IsDir() {
		return Target{}, aayamErrors.User(
			fmt.Sprintf("target must be a directory, not a file: %s", path),
		)
	}

	return Target{
		Path:     filepath.Clean(path),
		Explicit: true,
	}, nil
}
