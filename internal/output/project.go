package output

import (
	"encoding/json"
	"fmt"

	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintProject renders project identity and metadata to stdout.
//
// This is the dedicated renderer for: aayam project .
// It answers: "What is this project?"
// It does NOT include git, relationships, packages, or dependencies.
func (w *Writer) PrintProject(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam — Project\n\n")

	fmt.Fprintf(w.Out, "Name:          %s\n", snap.Name)
	fmt.Fprintf(w.Out, "Type:          %s\n", string(snap.Type))
	fmt.Fprintf(w.Out, "Root:          %s\n", snap.Root)

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Files:         %d\n", snap.FileCount)
	fmt.Fprintf(w.Out, "Directories:   %d\n", snap.DirectoryCount)

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Languages\n")
	if len(snap.Languages) == 0 {
		fmt.Fprintf(w.Out, "  (none detected)\n")
	} else {
		for _, lang := range snap.Languages {
			fmt.Fprintf(w.Out, "  %s\n", string(lang))
		}
	}
}

// JSONProjectResult is the top-level structure for machine-readable project output.
type JSONProjectResult struct {
	Application string             `json:"application"`
	Version     string             `json:"version"`
	Capability  string             `json:"capability"`
	Project     JSONProjectPayload `json:"project"`
}

// JSONProjectPayload holds the project-capability-specific fields.
type JSONProjectPayload struct {
	Name           string   `json:"name"`
	Root           string   `json:"root"`
	Type           string   `json:"type"`
	FileCount      int      `json:"file_count"`
	DirectoryCount int      `json:"directory_count"`
	Languages      []string `json:"languages"`
}

// PrintProjectJSON renders project identity and metadata as JSON to stdout.
func (w *Writer) PrintProjectJSON(snap snapshot.ProjectSnapshot) error {
	langs := make([]string, len(snap.Languages))
	for i, l := range snap.Languages {
		langs[i] = string(l)
	}

	result := JSONProjectResult{
		Application: "Aryntra Aayam",
		Version:     version,
		Capability:  "project",
		Project: JSONProjectPayload{
			Name:           snap.Name,
			Root:           snap.Root,
			Type:           string(snap.Type),
			FileCount:      snap.FileCount,
			DirectoryCount: snap.DirectoryCount,
			Languages:      langs,
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
