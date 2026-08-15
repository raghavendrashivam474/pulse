package project

import (
	"github.com/raghavendrashivam474/aayam/internal/scanner"
)

// ProjectType represents the detected primary ecosystem of a project.
type ProjectType string

const (
	TypeGo      ProjectType = "Go"
	TypeNode    ProjectType = "Node.js"
	TypeRust    ProjectType = "Rust"
	TypePython  ProjectType = "Python"
	TypeJava    ProjectType = "Java"
	TypeUnknown ProjectType = "Unknown"
)

// Detection is the result of project type analysis.
type Detection struct {
	// Primary is the single most-confident project type.
	Primary ProjectType

	// AllDetected contains every ecosystem detected from markers.
	AllDetected []ProjectType

	// Markers are the filenames that contributed to detection.
	Markers []string
}

// markerTypes maps known marker filenames to ecosystem type.
var markerTypes = map[string]ProjectType{
	"go.mod":           TypeGo,
	"package.json":     TypeNode,
	"Cargo.toml":       TypeRust,
	"pyproject.toml":   TypePython,
	"requirements.txt": TypePython,
	"pom.xml":          TypeJava,
	"build.gradle":     TypeJava,
}

// detectionOrder defines priority when multiple markers are present.
var detectionOrder = []string{
	"go.mod",
	"Cargo.toml",
	"pyproject.toml",
	"pom.xml",
	"build.gradle",
	"package.json",
	"requirements.txt",
}

// DetectType analyses a filesystem inventory and returns the detected
// project type.
func DetectType(inv scanner.Inventory) Detection {
	seenTypes := make(map[ProjectType]bool)
	var matchedMarkers []string

	for _, f := range inv.Files {
		name := baseName(f.RelPath)
		if t, ok := markerTypes[name]; ok {
			seenTypes[t] = true
			matchedMarkers = append(matchedMarkers, name)
		}
	}

	if len(seenTypes) == 0 {
		return inferTypeFromFiles(inv)
	}

	var allDetected []ProjectType
	var primary ProjectType

	for _, marker := range detectionOrder {
		t := markerTypes[marker]
		if seenTypes[t] && !typeInSlice(allDetected, t) {
			allDetected = append(allDetected, t)
			if primary == "" {
				primary = t
			}
		}
	}

	return Detection{
		Primary:     primary,
		AllDetected: allDetected,
		Markers:     matchedMarkers,
	}
}

// inferTypeFromFiles provides a minimal fallback when no marker files exist.
// Pre-S3 hardening only needs Go inference for the go-project fixture.
func inferTypeFromFiles(inv scanner.Inventory) Detection {
	for _, f := range inv.Files {
		if f.Extension == ".go" {
			return Detection{
				Primary:     TypeGo,
				AllDetected: []ProjectType{TypeGo},
				Markers:     nil,
			}
		}
	}

	return Detection{
		Primary:     TypeUnknown,
		AllDetected: []ProjectType{TypeUnknown},
		Markers:     nil,
	}
}

// baseName returns the final segment of a forward-slash-separated path.
func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// typeInSlice reports whether t is already present in the slice.
func typeInSlice(slice []ProjectType, t ProjectType) bool {
	for _, existing := range slice {
		if existing == t {
			return true
		}
	}
	return false
}
