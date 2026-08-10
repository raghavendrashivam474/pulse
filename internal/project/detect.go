package project

import (
	"pulse/internal/scanner"
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
	// Always set — TypeUnknown when no markers are recognised.
	Primary ProjectType

	// AllDetected contains every ecosystem detected from markers,
	// ordered by detection priority. A project with both go.mod and
	// package.json will have two entries here.
	AllDetected []ProjectType

	// Markers are the filenames that contributed to detection.
	Markers []string
}

// markerTypes maps known project marker filenames to their ecosystem type.
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
// The first match in this list becomes the Primary type.
// Every key in markerTypes must be reachable via this list.
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
//
// Detection rules:
//  1. Every file in the inventory is checked by filename (not full path).
//  2. All matching ecosystems are recorded.
//  3. Primary is chosen by the priority order in detectionOrder.
//  4. When no markers are found, Primary is TypeUnknown.
//
// TypeUnknown is not an error condition. It means Pulse does not recognise
// the project ecosystem — the project is still valid and can be analysed
// structurally. The distinction is:
//
//	"unable to analyse" != "unknown project type"
//
// Detection is deterministic: given identical inventory input, the output
// is always identical.
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
		return Detection{
			Primary:     TypeUnknown,
			AllDetected: []ProjectType{TypeUnknown},
			Markers:     nil,
		}
	}

	// Build AllDetected in priority order so the output is deterministic.
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

// baseName returns the final segment of a forward-slash-separated path.
// "internal/config/go.mod" -> "go.mod"
// "go.mod"                 -> "go.mod"
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
