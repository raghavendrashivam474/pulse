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

	"pulse/internal/codebase"
	"pulse/internal/git"
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

	// S4: Packages section
	pkgNames := codebase.PackageNames(snap.Codebase.Packages)
	if len(pkgNames) > 0 {
		fmt.Fprintf(w.Out, "\n")
		fmt.Fprintf(w.Out, "Packages\n")
		for _, name := range pkgNames {
			fmt.Fprintf(w.Out, "  %s\n", name)
		}
	}

	// S4: Dependencies section
	if len(snap.Codebase.Dependencies) > 0 {
		fmt.Fprintf(w.Out, "\n")
		fmt.Fprintf(w.Out, "Dependencies\n")
		for _, dep := range snap.Codebase.Dependencies {
			fmt.Fprintf(w.Out, "  %-20s → %s\n", dep.From, dep.To)
		}
	}

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Git\n")
	if !snap.Git.IsRepository {
		fmt.Fprintf(w.Out, "  Repository: No\n")
		return
	}

	fmt.Fprintf(w.Out, "  Repository:   %s\n", snap.Git.RepositoryName)
	fmt.Fprintf(w.Out, "  Branch:       %s\n", snap.Git.Branch)
	fmt.Fprintf(w.Out, "  Working Tree: %s\n", workingTreeLabel(snap.Git.WorkingTreeState))

	// History section (M3.5)
	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "History\n")
	if snap.Git.History.Count == 0 {
		fmt.Fprintf(w.Out, "  Commits: 0\n")
		fmt.Fprintf(w.Out, "  Latest:  None\n")
	} else {
		h := snap.Git.History
		fmt.Fprintf(w.Out, "  Commits: %d\n", h.Count)
		fmt.Fprintf(w.Out, "  Latest:  %q\n", h.Latest.Message)
		fmt.Fprintf(w.Out, "  Author:  %s\n", h.Latest.Author)
		fmt.Fprintf(w.Out, "  Date:    %s\n", h.Latest.Timestamp.Format("2006-01-02"))
	}

	// Contributors section (M3.6)
	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Contributors\n")
	if snap.Git.Contributors.Count == 0 {
		fmt.Fprintf(w.Out, "  Count: 0\n")
	} else {
		fmt.Fprintf(w.Out, "  Count: %d\n", snap.Git.Contributors.Count)
		for _, c := range snap.Git.Contributors.Items {
			fmt.Fprintf(w.Out, "  %-30s %d\n", c.Name, c.CommitCount)
		}
	}

	// Health section (M3.7)
	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Health\n")
	fmt.Fprintf(w.Out, "  Commits:      %s\n", yesNo(snap.Git.Health.HasCommits))
	fmt.Fprintf(w.Out, "  Contributors: %s\n", yesNo(snap.Git.Health.HasContributors))
	fmt.Fprintf(w.Out, "  Working Tree: %s\n", workingTreeLabel(snap.Git.WorkingTreeState))
	fmt.Fprintf(w.Out, "  HEAD:         %s\n", headLabel(snap.Git.Health.DetachedHEAD))
}

// workingTreeLabel converts a WorkingTreeState to a display string.
func workingTreeLabel(state git.WorkingTreeState) string {
	switch state {
	case git.WorkingTreeClean:
		return "Clean"
	case git.WorkingTreeDirty:
		return "Dirty"
	default:
		return "Unknown"
	}
}

// yesNo converts a bool to a human-readable Yes/No string.
func yesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// headLabel converts the detached HEAD bool to a display string.
func headLabel(detached bool) string {
	if detached {
		return "Detached"
	}
	return "Branch"
}

// ---------------------------------------------------------------------------
// JSON output
// ---------------------------------------------------------------------------

// JSONDiscoveryResult is the top-level structure for machine-readable
// project discovery output.
type JSONDiscoveryResult struct {
	Application string             `json:"application"`
	Version     string             `json:"version"`
	Project     JSONProjectSection `json:"project"`
}

// JSONProjectSection holds project-specific fields in the JSON output.
type JSONProjectSection struct {
	Name           string              `json:"name"`
	Root           string              `json:"root"`
	Type           string              `json:"type"`
	Languages      []string            `json:"languages"`
	FileCount      int                 `json:"file_count"`
	DirectoryCount int                 `json:"directory_count"`
	Codebase       JSONCodebaseSection `json:"codebase"`
	Git            JSONGitSection      `json:"git"`
}

// JSONCodebaseSection holds the structured codebase model in JSON output.
type JSONCodebaseSection struct {
	Files        []JSONFileEntry    `json:"files"`
	Directories  []JSONDirEntry     `json:"directories"`
	Packages     []JSONPackageEntry `json:"packages"`
	Dependencies []JSONDependency   `json:"dependencies"`
}

// JSONFileEntry represents a file in JSON output.
type JSONFileEntry struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Language  string `json:"language,omitempty"`
}

// JSONDirEntry represents a directory in JSON output.
type JSONDirEntry struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// JSONPackageEntry represents a package in JSON output.
type JSONPackageEntry struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Language string   `json:"language"`
	Files    []string `json:"files"`
}

// JSONDependency represents a dependency in JSON output.
type JSONDependency struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// JSONGitSection holds Git-specific fields in the JSON output.
type JSONGitSection struct {
	IsRepository   bool                `json:"is_repository"`
	RepositoryRoot string              `json:"repository_root,omitempty"`
	RepositoryName string              `json:"repository_name,omitempty"`
	Branch         string              `json:"branch,omitempty"`
	WorkingTree    string              `json:"working_tree,omitempty"`
	History        *JSONHistorySection `json:"history,omitempty"`
	Contributors   *JSONContributors   `json:"contributors,omitempty"`
	Health         *JSONHealthSection  `json:"health,omitempty"`
}

// JSONHistorySection holds commit history in the JSON output.
type JSONHistorySection struct {
	Count  int             `json:"count"`
	Latest *JSONCommitInfo `json:"latest,omitempty"`
}

// JSONCommitInfo holds a single commit in the JSON output.
type JSONCommitInfo struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
}

// JSONContributors holds contributor data in the JSON output.
type JSONContributors struct {
	Count int                   `json:"count"`
	Items []JSONContributorItem `json:"items"`
}

// JSONContributorItem holds one contributor in the JSON output.
type JSONContributorItem struct {
	Name        string `json:"name"`
	CommitCount int    `json:"commit_count"`
}

// JSONHealthSection holds health signals in the JSON output.
type JSONHealthSection struct {
	HasCommits       bool `json:"has_commits"`
	HasContributors  bool `json:"has_contributors"`
	WorkingTreeClean bool `json:"working_tree_clean"`
	DetachedHEAD     bool `json:"detached_head"`
}

// PrintDiscoveryJSON renders a ProjectSnapshot as JSON to stdout.
func (w *Writer) PrintDiscoveryJSON(snap snapshot.ProjectSnapshot) error {
	langs := make([]string, len(snap.Languages))
	for i, l := range snap.Languages {
		langs[i] = string(l)
	}

	// Build codebase section.
	cbFiles := make([]JSONFileEntry, len(snap.Codebase.Files))
	for i, f := range snap.Codebase.Files {
		cbFiles[i] = JSONFileEntry{
			Path:      f.Path,
			Name:      f.Name,
			Extension: f.Extension,
			Language:  string(f.Language),
		}
	}

	cbDirs := make([]JSONDirEntry, len(snap.Codebase.Directories))
	for i, d := range snap.Codebase.Directories {
		cbDirs[i] = JSONDirEntry{
			Path: d.Path,
			Name: d.Name,
		}
	}

	cbPkgs := make([]JSONPackageEntry, len(snap.Codebase.Packages))
	for i, p := range snap.Codebase.Packages {
		cbPkgs[i] = JSONPackageEntry{
			Name:     p.Name,
			Path:     p.Path,
			Language: string(p.Language),
			Files:    p.Files,
		}
	}

	cbDeps := make([]JSONDependency, len(snap.Codebase.Dependencies))
	for i, d := range snap.Codebase.Dependencies {
		cbDeps[i] = JSONDependency{
			From: d.From,
			To:   d.To,
			Type: string(d.Type),
		}
	}

	codebaseSection := JSONCodebaseSection{
		Files:        cbFiles,
		Directories:  cbDirs,
		Packages:     cbPkgs,
		Dependencies: cbDeps,
	}

	// Build git section.
	gitSection := JSONGitSection{
		IsRepository: snap.Git.IsRepository,
	}

	if snap.Git.IsRepository {
		gitSection.RepositoryRoot = snap.Git.RepositoryRoot
		gitSection.RepositoryName = snap.Git.RepositoryName
		gitSection.Branch = snap.Git.Branch
		gitSection.WorkingTree = string(snap.Git.WorkingTreeState)

		// History
		h := snap.Git.History
		histSection := &JSONHistorySection{Count: h.Count}
		if h.Count > 0 {
			histSection.Latest = &JSONCommitInfo{
				Hash:      h.Latest.Hash,
				Message:   h.Latest.Message,
				Author:    h.Latest.Author,
				Timestamp: h.Latest.Timestamp.Format("2006-01-02T15:04:05Z"),
			}
		}
		gitSection.History = histSection

		// Contributors
		contribItems := make([]JSONContributorItem, len(snap.Git.Contributors.Items))
		for i, c := range snap.Git.Contributors.Items {
			contribItems[i] = JSONContributorItem{
				Name:        c.Name,
				CommitCount: c.CommitCount,
			}
		}
		gitSection.Contributors = &JSONContributors{
			Count: snap.Git.Contributors.Count,
			Items: contribItems,
		}

		// Health
		gitSection.Health = &JSONHealthSection{
			HasCommits:       snap.Git.Health.HasCommits,
			HasContributors:  snap.Git.Health.HasContributors,
			WorkingTreeClean: snap.Git.Health.WorkingTreeClean,
			DetachedHEAD:     snap.Git.Health.DetachedHEAD,
		}
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
			Codebase:       codebaseSection,
			Git:            gitSection,
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
