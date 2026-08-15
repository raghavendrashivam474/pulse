package output

import (
	"encoding/json"
	"fmt"

	"github.com/raghavendrashivam474/aayam/internal/codebase"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintRelationships renders the relationship graph summary to stdout.
//
// This is the dedicated renderer for: aayam relationships .
// It answers: "How are the things in this project connected?"
// It does NOT include project identity, structure counts, or git.
func (w *Writer) PrintRelationships(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam — Relationships\n\n")

	fmt.Fprintf(w.Out, "Graph\n")
	fmt.Fprintf(w.Out, "  Nodes:  %d\n", snap.Graph.NodeCount())
	fmt.Fprintf(w.Out, "  Edges:  %d\n", snap.Graph.EdgeCount())

	byKind := snap.Graph.EdgesByKind()

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Relationship Types\n")

	contains := byKind[codebase.RelationshipContains]
	belongsTo := byKind[codebase.RelationshipBelongsTo]
	imports := byKind[codebase.RelationshipImports]
	dependsOn := byKind[codebase.RelationshipDependsOn]

	if contains+belongsTo+imports+dependsOn == 0 {
		fmt.Fprintf(w.Out, "  (none)\n")
		return
	}
	if contains > 0 {
		fmt.Fprintf(w.Out, "  Contains:    %d\n", contains)
	}
	if belongsTo > 0 {
		fmt.Fprintf(w.Out, "  Belongs to:  %d\n", belongsTo)
	}
	if imports > 0 {
		fmt.Fprintf(w.Out, "  Imports:     %d\n", imports)
	}
	if dependsOn > 0 {
		fmt.Fprintf(w.Out, "  Depends on:  %d\n", dependsOn)
	}
}

// JSONRelationshipsResult is the top-level structure for machine-readable
// relationships output.
type JSONRelationshipsResult struct {
	Application   string                   `json:"application"`
	Version       string                   `json:"version"`
	Capability    string                   `json:"capability"`
	Relationships JSONRelationshipsPayload `json:"relationships"`
}

// JSONRelationshipsPayload holds the relationships-capability-specific fields.
type JSONRelationshipsPayload struct {
	Nodes     int `json:"nodes"`
	Edges     int `json:"edges"`
	Contains  int `json:"contains"`
	BelongsTo int `json:"belongs_to"`
	Imports   int `json:"imports"`
	DependsOn int `json:"depends_on"`
}

// PrintRelationshipsJSON renders the relationship graph summary as JSON to stdout.
func (w *Writer) PrintRelationshipsJSON(snap snapshot.ProjectSnapshot) error {
	byKind := snap.Graph.EdgesByKind()

	result := JSONRelationshipsResult{
		Application: "Aryntra Aayam",
		Version:     version,
		Capability:  "relationships",
		Relationships: JSONRelationshipsPayload{
			Nodes:     snap.Graph.NodeCount(),
			Edges:     snap.Graph.EdgeCount(),
			Contains:  byKind[codebase.RelationshipContains],
			BelongsTo: byKind[codebase.RelationshipBelongsTo],
			Imports:   byKind[codebase.RelationshipImports],
			DependsOn: byKind[codebase.RelationshipDependsOn],
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
