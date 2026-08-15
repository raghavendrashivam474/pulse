package output

import (
	"encoding/json"
	"fmt"

	gitinfo "github.com/raghavendrashivam474/aayam/internal/git"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintHistory renders repository history summary to stdout.
//
// This is the dedicated renderer for: aayam history .
// It answers: "What has happened in this project's history?"
// It does NOT include project structure or relationship graph output.
func (w *Writer) PrintHistory(snap snapshot.ProjectSnapshot) {
	fmt.Fprintf(w.Out, "Aryntra Aayam - History\n\n")

	info := snap.Git
	if !info.IsRepository {
		fmt.Fprintf(w.Out, "Repository\n")
		fmt.Fprintf(w.Out, "  State: Not a Git repository\n")
		return
	}

	fmt.Fprintf(w.Out, "Commits\n")
	fmt.Fprintf(w.Out, "  Total: %d\n", info.History.Count)

	fmt.Fprintf(w.Out, "\nLatest Commit\n")
	if info.History.Count == 0 {
		fmt.Fprintf(w.Out, "  (none)\n")
	} else {
		fmt.Fprintf(w.Out, "  %q\n", info.History.Latest.Message)
		fmt.Fprintf(w.Out, "  Author: %s\n", info.History.Latest.Author)
		fmt.Fprintf(w.Out, "  Date: %s\n", formatGitDate(info.History.Latest.Timestamp))
	}

	fmt.Fprintf(w.Out, "\nContributors\n")
	fmt.Fprintf(w.Out, "  Count: %d\n", info.Contributors.Count)

	if info.Contributors.Count > 0 {
		fmt.Fprintf(w.Out, "\n")
		for _, c := range info.Contributors.Items {
			fmt.Fprintf(w.Out, "  %s\n", c.Name)
			fmt.Fprintf(w.Out, "    Commits: %d\n", c.CommitCount)
		}
	}
}

// JSONHistoryResult is the top-level structure for machine-readable
// history output.
type JSONHistoryResult struct {
	Application string             `json:"application"`
	Version     string             `json:"version"`
	Capability  string             `json:"capability"`
	History     JSONHistoryPayload `json:"history"`
}

// JSONHistoryPayload holds the history-capability-specific fields.
type JSONHistoryPayload struct {
	IsRepository bool                    `json:"is_repository"`
	CommitCount  int                     `json:"commit_count"`
	Latest       *JSONHistoryCommit      `json:"latest"`
	Contributors JSONHistoryContributors `json:"contributors"`
}

// JSONHistoryCommit holds latest commit details.
type JSONHistoryCommit struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
}

// JSONHistoryContributors holds contributor summary data.
type JSONHistoryContributors struct {
	Count int                      `json:"count"`
	Items []JSONHistoryContributor `json:"items"`
}

// JSONHistoryContributor represents one contributor.
type JSONHistoryContributor struct {
	Name        string `json:"name"`
	CommitCount int    `json:"commit_count"`
}

// PrintHistoryJSON renders repository history summary as JSON to stdout.
func (w *Writer) PrintHistoryJSON(snap snapshot.ProjectSnapshot) error {
	info := snap.Git

	items := make([]JSONHistoryContributor, 0, len(info.Contributors.Items))
	for _, c := range info.Contributors.Items {
		items = append(items, JSONHistoryContributor{
			Name:        c.Name,
			CommitCount: c.CommitCount,
		})
	}

	result := JSONHistoryResult{
		Application: "Aryntra Aayam",
		Version:     version,
		Capability:  "history",
		History: JSONHistoryPayload{
			IsRepository: info.IsRepository,
			CommitCount:  info.History.Count,
			Latest:       jsonHistoryCommitPtr(info.History.Latest, info.History.Count > 0),
			Contributors: JSONHistoryContributors{
				Count: info.Contributors.Count,
				Items: items,
			},
		},
	}

	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func jsonHistoryCommitPtr(c gitinfo.CommitInfo, ok bool) *JSONHistoryCommit {
	if !ok {
		return nil
	}
	return &JSONHistoryCommit{
		Hash:      c.Hash,
		Message:   c.Message,
		Author:    c.Author,
		Timestamp: c.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
