package project

import (
	"os"
	"path/filepath"
)

// rootMarkers are filesystem entries whose presence strongly suggests
// that the containing directory is a project root.
//
// Deliberately small — well-known, unambiguous markers only.
// Extended in later sprints as Pulse learns more ecosystems.
var rootMarkers = []string{
	".git",
	"go.mod",
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"pom.xml",
	"build.gradle",
}

// RootDiscovery is the result of walking upward from a target directory.
type RootDiscovery struct {
	// Root is the discovered project root directory.
	// Always a valid absolute path.
	Root string

	// Marker is the filename that indicated the project root.
	// Empty string when no marker was found (MarkerFound == false).
	Marker string

	// MarkerFound reports whether a root marker was found during the walk.
	// When false, Root equals the original target path (fallback behaviour).
	MarkerFound bool
}

// DiscoverRoot walks upward from target.Path looking for a recognised
// project root marker.
//
// Behaviour when no marker is found:
//   - The target directory itself is treated as the root.
//   - RootDiscovery.MarkerFound will be false.
//   - This is not an error. Pulse can still analyse the project.
//
// The walk stops at the filesystem root to prevent runaway traversal.
// Symlinks in the path are not resolved — the walk follows the logical
// path as provided.
func DiscoverRoot(target Target) RootDiscovery {
	current := target.Path

	for {
		if marker, found := findMarker(current); found {
			return RootDiscovery{
				Root:        current,
				Marker:      marker,
				MarkerFound: true,
			}
		}

		parent := filepath.Dir(current)

		// filepath.Dir returns the same string when we have reached the
		// filesystem root. Stop here to avoid an infinite loop.
		if parent == current {
			break
		}

		current = parent
	}

	// No marker found anywhere in the ancestor chain.
	// Fall back to the original target as the root.
	// MarkerFound == false documents this explicitly to callers.
	return RootDiscovery{
		Root:        target.Path,
		Marker:      "",
		MarkerFound: false,
	}
}

// findMarker checks whether any known root marker exists inside dir.
// Returns the marker name and true if found, empty string and false otherwise.
func findMarker(dir string) (string, bool) {
	for _, marker := range rootMarkers {
		candidate := filepath.Join(dir, marker)
		if _, err := os.Stat(candidate); err == nil {
			return marker, true
		}
	}
	return "", false
}
