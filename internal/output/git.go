package output

import (
	"encoding/json"
	"fmt"
	"time"

	gitinfo "github.com/raghavendrashivam474/aayam/internal/git"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintGit renders the current Git state to stdout.
//
// This is the dedicated renderer for: aayam git .
// It answers: "What is the current Git state of this project?"
// It does NOT include project structure or relationship graph output.
func (w *Writer) PrintGit(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam - Git\n\n")

	info := snap.Git
	if !info.IsRepository {
		fmt.Fprintf(w.Out, "Repository\n")
		fmt.Fprintf(w.Out, "  State: Not a Git repository\n")
		return
	}

	fmt.Fprintf(w.Out, "Repository\n")
	fmt.Fprintf(w.Out, "  Name: %s\n", info.RepositoryName)
	fmt.Fprintf(w.Out, "  Branch: %s\n", info.Branch)
	fmt.Fprintf(w.Out, "  Working Tree: %s\n", formatWorkingTreeState(info.WorkingTreeState))

	fmt.Fprintf(w.Out, "\nHEAD\n")
	fmt.Fprintf(w.Out, "  State: %s\n", terminalHeadState(info))
	if !info.Health.DetachedHEAD {
		fmt.Fprintf(w.Out, "  Branch: %s\n", info.Branch)
	}

	fmt.Fprintf(w.Out, "\nHistory\n")
	fmt.Fprintf(w.Out, "  Commits: %d\n", info.History.Count)
	if info.History.Count > 0 {
		fmt.Fprintf(w.Out, "  Latest: %q\n", info.History.Latest.Message)
		fmt.Fprintf(w.Out, "  Author: %s\n", info.History.Latest.Author)
		fmt.Fprintf(w.Out, "  Date: %s\n", formatGitDate(info.History.Latest.Timestamp))
	}

	fmt.Fprintf(w.Out, "\nContributors\n")
	fmt.Fprintf(w.Out, "  Count: %d\n", info.Contributors.Count)
	if info.Contributors.Count > 0 {
		fmt.Fprintf(w.Out, "\n")
		for _, c := range info.Contributors.Items {
			fmt.Fprintf(w.Out, "  %s    %d\n", c.Name, c.CommitCount)
		}
	}
}

// JSONGitResult is the top-level structure for machine-readable git output.
type JSONGitResult struct {
	Application string         `json:"application"`
	Version     string         `json:"version"`
	Capability  string         `json:"capability"`
	Git         JSONGitPayload `json:"git"`
}

// JSONGitPayload holds the git-capability-specific fields.
type JSONGitPayload struct {
	IsRepository     bool                `json:"is_repository"`
	RepositoryRoot   string              `json:"repository_root"`
	RepositoryName   string              `json:"repository_name"`
	Branch           string              `json:"branch"`
	WorkingTreeState string              `json:"working_tree_state"`
	Head             JSONGitHead         `json:"head"`
	History          JSONGitHistory      `json:"history"`
	Contributors     JSONGitContributors `json:"contributors"`
	Health           JSONGitHealth       `json:"health"`
}

// JSONGitHead describes the current HEAD state.
type JSONGitHead struct {
	State  string `json:"state"`
	Branch string `json:"branch,omitempty"`
}

// JSONGitHistory holds commit history summary data.
type JSONGitHistory struct {
	Count  int            `json:"count"`
	Latest *JSONGitCommit `json:"latest"`
}

// JSONGitCommit holds machine-readable latest commit details.
type JSONGitCommit struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// JSONGitContributors holds contributor summary data.
type JSONGitContributors struct {
	Count int                  `json:"count"`
	Items []JSONGitContributor `json:"items"`
}

// JSONGitContributor represents one contributor.
type JSONGitContributor struct {
	Name        string `json:"name"`
	CommitCount int    `json:"commit_count"`
}

// JSONGitHealth exposes objective repository state signals.
type JSONGitHealth struct {
	HasCommits       bool `json:"has_commits"`
	HasContributors  bool `json:"has_contributors"`
	WorkingTreeClean bool `json:"working_tree_clean"`
	DetachedHEAD     bool `json:"detached_head"`
}

// PrintGitJSON renders the current Git state as JSON to stdout.
func (w *Writer) PrintGitJSON(snap snapshot.ProjectSnapshot) error {
	info := snap.Git

	items := make([]JSONGitContributor, 0, len(info.Contributors.Items))
	for _, c := range info.Contributors.Items {
		items = append(items, JSONGitContributor{
			Name:        c.Name,
			CommitCount: c.CommitCount,
		})
	}

	head := JSONGitHead{
		State: jsonHeadState(info),
	}
	if info.IsRepository && !info.Health.DetachedHEAD {
		head.Branch = info.Branch
	}

	result := JSONGitResult{
		Application: "Aryntra Aayam",
		Version:     version,
		Capability:  "git",
		Git: JSONGitPayload{
			IsRepository:     info.IsRepository,
			RepositoryRoot:   info.RepositoryRoot,
			RepositoryName:   info.RepositoryName,
			Branch:           info.Branch,
			WorkingTreeState: string(info.WorkingTreeState),
			Head:             head,
			History: JSONGitHistory{
				Count:  info.History.Count,
				Latest: jsonGitCommitPtr(info.History.Latest, info.History.Count > 0),
			},
			Contributors: JSONGitContributors{
				Count: info.Contributors.Count,
				Items: items,
			},
			Health: JSONGitHealth{
				HasCommits:       info.Health.HasCommits,
				HasContributors:  info.Health.HasContributors,
				WorkingTreeClean: info.Health.WorkingTreeClean,
				DetachedHEAD:     info.Health.DetachedHEAD,
			},
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func jsonGitCommitPtr(c gitinfo.CommitInfo, ok bool) *JSONGitCommit {
	if !ok {
		return nil
	}
	return &JSONGitCommit{
		Hash:      c.Hash,
		Message:   c.Message,
		Author:    c.Author,
		Timestamp: c.Timestamp,
	}
}

func jsonHeadState(info gitinfo.GitInfo) string {
	if !info.IsRepository {
		return "none"
	}
	if info.Health.DetachedHEAD {
		return "detached"
	}
	return "branch"
}

func terminalHeadState(info gitinfo.GitInfo) string {
	switch jsonHeadState(info) {
	case "detached":
		return "Detached"
	case "branch":
		return "Branch"
	default:
		return "None"
	}
}

func formatWorkingTreeState(state gitinfo.WorkingTreeState) string {
	switch state {
	case gitinfo.WorkingTreeClean:
		return "Clean"
	case gitinfo.WorkingTreeDirty:
		return "Dirty"
	default:
		return "Unknown"
	}
}

func formatGitDate(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format("2006-01-02")
}
