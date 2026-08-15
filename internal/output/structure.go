package output

import (
	"encoding/json"
	"fmt"

	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintStructure renders filesystem and language structure to stdout.
//
// This is the dedicated renderer for: aayam structure .
// It answers: "How is this project organized?"
// It does NOT include project identity, git, or relationships.
func (w *Writer) PrintStructure(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam — Structure\n\n")

	fmt.Fprintf(w.Out, "Files:        %d\n", snap.FileCount)
	fmt.Fprintf(w.Out, "Directories:  %d\n", snap.DirectoryCount)

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

// JSONStructureResult is the top-level structure for machine-readable structure output.
type JSONStructureResult struct {
	Application string               `json:"application"`
	Version     string               `json:"version"`
	Capability  string               `json:"capability"`
	Structure   JSONStructurePayload `json:"structure"`
}

// JSONStructurePayload holds the structure-capability-specific fields.
type JSONStructurePayload struct {
	FileCount      int      `json:"file_count"`
	DirectoryCount int      `json:"directory_count"`
	Languages      []string `json:"languages"`
}

// PrintStructureJSON renders filesystem and language structure as JSON to stdout.
func (w *Writer) PrintStructureJSON(snap snapshot.ProjectSnapshot) error {
	langs := make([]string, len(snap.Languages))
	for i, l := range snap.Languages {
		langs[i] = string(l)
	}

	result := JSONStructureResult{
		Application: "Aryntra Aayam",
		Version:     version,
		Capability:  "structure",
		Structure: JSONStructurePayload{
			FileCount:      snap.FileCount,
			DirectoryCount: snap.DirectoryCount,
			Languages:      langs,
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
