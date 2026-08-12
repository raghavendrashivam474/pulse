package project

import (
	"os"
	"path/filepath"
)

// rootMarkers are filesystem entries whose presence strongly suggests
// that the containing directory is a project root.
var rootMarkers = []string{
	".git",
	"go.mod",
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"pom.xml",
	"build.gradle",
}

// RootDiscovery is the result of root discovery from a target directory.
type RootDiscovery struct {
	// Root is the discovered project root directory.
	Root string

	// Marker is the filename that indicated the root.
	Marker string

	// MarkerFound reports whether a root marker was found.
	MarkerFound bool
}

// DiscoverRoot resolves the project root for a validated target.
//
// Hardening rule:
// If the target was explicitly supplied by the user, that directory is the
// analysis boundary. Discovery must not walk upward into ancestors.
//
// If target.Explicit is false, upward walking is still allowed.
func DiscoverRoot(target Target) RootDiscovery {
	if target.Explicit {
		if marker, found := findMarker(target.Path); found {
			return RootDiscovery{
				Root:        target.Path,
				Marker:      marker,
				MarkerFound: true,
			}
		}

		return RootDiscovery{
			Root:        target.Path,
			Marker:      "",
			MarkerFound: false,
		}
	}

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
		if parent == current {
			break
		}

		current = parent
	}

	return RootDiscovery{
		Root:        target.Path,
		Marker:      "",
		MarkerFound: false,
	}
}

// findMarker checks whether any known root marker exists inside dir.
func findMarker(dir string) (string, bool) {
	for _, marker := range rootMarkers {
		candidate := filepath.Join(dir, marker)
		if _, err := os.Stat(candidate); err == nil {
			return marker, true
		}
	}
	return "", false
}
