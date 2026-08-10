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
func (w *Writer) PrintSummary() {
	fmt.Fprintf(w.Out, "Pulse v%s\n\n", version)
	fmt.Fprintf(w.Out, "Project intelligence for developers.\n\n")
	fmt.Fprintf(w.Out, "No project analysis available yet.\n")
}

// JSONResult is the top-level structure for machine-readable output.
type JSONResult struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// PrintJSON writes a JSON result to stdout.
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
