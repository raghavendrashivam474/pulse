package output

import (
	"encoding/json"
	"fmt"

	"github.com/raghavendrashivam474/aayam/internal/codebase"
	"github.com/raghavendrashivam474/aayam/internal/git"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintOverview renders a concise project overview to stdout.
//
// Overview summarizes. Capabilities explain.
// This is the default output for: aayam .
func (w *Writer) PrintOverview(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam v%s\n\n", version)

	fmt.Fprintf(w.Out, "Project\n")
	fmt.Fprintf(w.Out, "  Name: %s\n", snap.Name)
	fmt.Fprintf(w.Out, "  Type: %s\n", string(snap.Type))

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Structure\n")
	fmt.Fprintf(w.Out, "  Files:       %d\n", snap.FileCount)
	fmt.Fprintf(w.Out, "  Directories: %d\n", snap.DirectoryCount)

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
	fmt.Fprintf(w.Out, "Relationships\n")
	fmt.Fprintf(w.Out, "  Nodes: %d\n", snap.Graph.NodeCount())
	fmt.Fprintf(w.Out, "  Edges: %d\n", snap.Graph.EdgeCount())
	if snap.Graph.EdgeCount() > 0 {
		byKind := snap.Graph.EdgesByKind()
		if n := byKind[codebase.RelationshipContains]; n > 0 {
			fmt.Fprintf(w.Out, "  Contains:   %d\n", n)
		}
		if n := byKind[codebase.RelationshipBelongsTo]; n > 0 {
			fmt.Fprintf(w.Out, "  Belongs to: %d\n", n)
		}
		if n := byKind[codebase.RelationshipImports]; n > 0 {
			fmt.Fprintf(w.Out, "  Imports:    %d\n", n)
		}
		if n := byKind[codebase.RelationshipDependsOn]; n > 0 {
			fmt.Fprintf(w.Out, "  Depends on: %d\n", n)
		}
	}

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Git\n")
	if !snap.Git.IsRepository {
		fmt.Fprintf(w.Out, "  Repository: No\n")
		return
	}
	fmt.Fprintf(w.Out, "  Branch:       %s\n", snap.Git.Branch)
	fmt.Fprintf(w.Out, "  Working Tree: %s\n", workingTreeLabel(snap.Git.WorkingTreeState))

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Health\n")
	fmt.Fprintf(w.Out, "  Commits:      %s\n", yesNo(snap.Git.Health.HasCommits))
	fmt.Fprintf(w.Out, "  Contributors: %s\n", yesNo(snap.Git.Health.HasContributors))
	fmt.Fprintf(w.Out, "  Working Tree: %s\n", workingTreeLabel(snap.Git.WorkingTreeState))
}

// JSONOverviewResult is the top-level structure for machine-readable overview output.
type JSONOverviewResult struct {
	Application    string                    `json:"application"`
	Version        string                    `json:"version"`
	Capability     string                    `json:"capability"`
	Name           string                    `json:"name"`
	Type           string                    `json:"type"`
	FileCount      int                       `json:"file_count"`
	DirectoryCount int                       `json:"directory_count"`
	Languages      []string                  `json:"languages"`
	Relationships  JSONOverviewRelationships `json:"relationships"`
	Git            JSONOverviewGit           `json:"git"`
}

// JSONOverviewRelationships holds the relationship summary in JSON overview output.
type JSONOverviewRelationships struct {
	Nodes     int `json:"nodes"`
	Edges     int `json:"edges"`
	Contains  int `json:"contains"`
	BelongsTo int `json:"belongs_to"`
	Imports   int `json:"imports"`
	DependsOn int `json:"depends_on"`
}

// JSONOverviewGit holds the Git summary in JSON overview output.
type JSONOverviewGit struct {
	IsRepository    bool   `json:"is_repository"`
	Branch          string `json:"branch,omitempty"`
	WorkingTree     string `json:"working_tree,omitempty"`
	HasCommits      bool   `json:"has_commits"`
	HasContributors bool   `json:"has_contributors"`
}

// PrintOverviewJSON renders a concise project overview as JSON to stdout.
func (w *Writer) PrintOverviewJSON(snap snapshot.ProjectSnapshot) error {
	langs := make([]string, len(snap.Languages))
	for i, l := range snap.Languages {
		langs[i] = string(l)
	}

	byKind := snap.Graph.EdgesByKind()
	relSection := JSONOverviewRelationships{
		Nodes:     snap.Graph.NodeCount(),
		Edges:     snap.Graph.EdgeCount(),
		Contains:  byKind[codebase.RelationshipContains],
		BelongsTo: byKind[codebase.RelationshipBelongsTo],
		Imports:   byKind[codebase.RelationshipImports],
		DependsOn: byKind[codebase.RelationshipDependsOn],
	}

	gitSection := JSONOverviewGit{
		IsRepository: snap.Git.IsRepository,
	}
	if snap.Git.IsRepository {
		gitSection.Branch = snap.Git.Branch
		gitSection.WorkingTree = string(snap.Git.WorkingTreeState)
		gitSection.HasCommits = snap.Git.Health.HasCommits
		gitSection.HasContributors = snap.Git.Health.HasContributors
	}

	result := JSONOverviewResult{
		Application:    "Aryntra Aayam",
		Version:        version,
		Capability:     "overview",
		Name:           snap.Name,
		Type:           string(snap.Type),
		FileCount:      snap.FileCount,
		DirectoryCount: snap.DirectoryCount,
		Languages:      langs,
		Relationships:  relSection,
		Git:            gitSection,
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// workingTreeLabelForOverview is defined in output.go as workingTreeLabel.
// We reuse it directly since we are in the same package.
// (No duplicate needed.)

// gitWorkingTreeState converts WorkingTreeState for overview JSON.
func gitWorkingTreeState(state git.WorkingTreeState) string {
	return workingTreeLabel(state)
}
