// Package output handles all user-facing rendering for Pulse.
//
// Normal output goes to stdout.
// Errors and diagnostics go to stderr.
//
// No other package should write directly to stdout or stderr.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"pulse/internal/snapshot"
)

const version = "1.0.0"

// Writer holds the output destinations for Pulse.
type Writer struct {
	Out io.Writer
	Err io.Writer
}

// Default returns a Writer that uses os.Stdout and os.Stderr.
func Default() *Writer {
	return &Writer{
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

// PrintVersion writes the Pulse version to stdout.
func (w *Writer) PrintVersion() {
	fmt.Fprintf(w.Out, "Pulse v%s\n", version)
}

// PrintHelp writes usage information to stdout.
func (w *Writer) PrintHelp() {
	fmt.Fprintf(w.Out, "Pulse v%s\n\n", version)
	fmt.Fprintf(w.Out, "Project intelligence for developers.\n\n")
	fmt.Fprintf(w.Out, "Usage:\n")
	fmt.Fprintf(w.Out, "  pulse [path] [flags]\n\n")
	fmt.Fprintf(w.Out, "Flags:\n")
	fmt.Fprintf(w.Out, "  --help       Show this help message\n")
	fmt.Fprintf(w.Out, "  --version    Show version information\n")
	fmt.Fprintf(w.Out, "  --json       Output results as JSON\n")
}

// PrintSummary writes the standard human-readable Pulse summary to stdout.
// Used when no target path is provided.
func (w *Writer) PrintSummary() {
	fmt.Fprintf(w.Out, "Pulse v%s\n\n", version)
	fmt.Fprintf(w.Out, "Project intelligence for developers.\n\n")
	fmt.Fprintf(w.Out, "No project analysis available yet.\n")
}

// JSONResult is the top-level structure for machine-readable output
// when no target path is provided.
type JSONResult struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// PrintJSON writes a placeholder JSON result to stdout.
// Used when no target path is provided.
func (w *Writer) PrintJSON() error {
	result := JSONResult{
		Version: version,
		Status:  "no_analysis",
	}
	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// PrintError writes a formatted error message to stderr.
func (w *Writer) PrintError(message string) {
	fmt.Fprintf(w.Err, "Error: %s\n", message)
}

// PrintDiscovery renders a ProjectSnapshot as human-readable text to stdout.
func (w *Writer) PrintDiscovery(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Pulse v%s\n\n", version)

	fmt.Fprintf(w.Out, "Project\n")
	fmt.Fprintf(w.Out, "  Name: %s\n", snap.Name)
	fmt.Fprintf(w.Out, "  Type: %s\n", string(snap.Type))
	fmt.Fprintf(w.Out, "  Root: %s\n", snap.Root)

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Languages\n")
	if len(snap.Languages) == 0 {
		fmt.Fprintf(w.Out, "  (none detected)\n")
	} else {
		for _, lang := range snap.Languages {
			fmt.Fprintf(w.Out, "  %s\n", string(lang))
		}
	}

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Structure\n")
	fmt.Fprintf(w.Out, "  Files:       %d\n", snap.FileCount)
	fmt.Fprintf(w.Out, "  Directories: %d\n", snap.DirectoryCount)
}

// JSONDiscoveryResult is the top-level structure for machine-readable
// project discovery output.
type JSONDiscoveryResult struct {
	Application string             `json:"application"`
	Version     string             `json:"version"`
	Project     JSONProjectSection `json:"project"`
}

// JSONProjectSection holds project-specific fields in the JSON output.
type JSONProjectSection struct {
	Name           string   `json:"name"`
	Root           string   `json:"root"`
	Type           string   `json:"type"`
	Languages      []string `json:"languages"`
	FileCount      int      `json:"file_count"`
	DirectoryCount int      `json:"directory_count"`
}

// PrintDiscoveryJSON renders a ProjectSnapshot as JSON to stdout.
func (w *Writer) PrintDiscoveryJSON(snap snapshot.ProjectSnapshot) error {
	langs := make([]string, len(snap.Languages))
	for i, l := range snap.Languages {
		langs[i] = string(l)
	}

	result := JSONDiscoveryResult{
		Application: "Pulse",
		Version:     version,
		Project: JSONProjectSection{
			Name:           snap.Name,
			Root:           snap.Root,
			Type:           string(snap.Type),
			Languages:      langs,
			FileCount:      snap.FileCount,
			DirectoryCount: snap.DirectoryCount,
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
