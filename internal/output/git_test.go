package output_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	gitinfo "github.com/raghavendrashivam474/aayam/internal/git"
	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

func makeGitSnap(info gitinfo.GitInfo) snapshot.ProjectSnapshot {
	return snapshot.ProjectSnapshot{
		Git: info,
	}
}

func sampleGitInfo() gitinfo.GitInfo {
	return gitinfo.GitInfo{
		IsRepository:     true,
		RepositoryRoot:   "C:/repo/aayam",
		RepositoryName:   "aayam",
		Branch:           "main",
		WorkingTreeState: gitinfo.WorkingTreeClean,
		History: gitinfo.CommitHistory{
			Count: 3,
			Latest: gitinfo.CommitInfo{
				Hash:      "abc123",
				Message:   "feat: add git capability",
				Author:    "Raghavendra Singh",
				Timestamp: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		Contributors: gitinfo.Contributors{
			Count: 1,
			Items: []gitinfo.Contributor{
				{Name: "Raghavendra Singh", CommitCount: 3},
			},
		},
		Health: gitinfo.RepositoryHealth{
			HasCommits:       true,
			HasContributors:  true,
			WorkingTreeClean: true,
			DetachedHEAD:     false,
		},
	}
}

func TestPrintGit_heading_and_sections(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintGit(makeGitSnap(sampleGitInfo()))

	out := outBuf.String()
	if !strings.Contains(out, "Git") {
		t.Fatalf("expected Git heading, got:\n%s", out)
	}
	if !strings.Contains(out, "Repository") || !strings.Contains(out, "HEAD") || !strings.Contains(out, "History") {
		t.Fatalf("expected repository/head/history sections, got:\n%s", out)
	}
}

func TestPrintGit_not_repository(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintGit(makeGitSnap(gitinfo.GitInfo{IsRepository: false}))

	out := outBuf.String()
	if !strings.Contains(out, "Not a Git repository") {
		t.Fatalf("expected non-repository message, got:\n%s", out)
	}
}

func TestPrintGit_no_structure_leak(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintGit(makeGitSnap(sampleGitInfo()))

	out := outBuf.String()
	if strings.Contains(out, "Files") || strings.Contains(out, "Directories") || strings.Contains(out, "Nodes") {
		t.Fatalf("git output must not contain structure/graph data, got:\n%s", out)
	}
}

func TestPrintGitJSON_envelope(t *testing.T) {
	w, outBuf, _ := newTestWriter()

	if err := w.PrintGitJSON(makeGitSnap(sampleGitInfo())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.JSONGitResult
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Capability != "git" {
		t.Fatalf("expected capability=git, got %q", result.Capability)
	}
	if !result.Git.IsRepository {
		t.Fatal("expected is_repository=true")
	}
	if result.Git.RepositoryName != "aayam" {
		t.Fatalf("expected repository_name=aayam, got %q", result.Git.RepositoryName)
	}
}

func TestPrintGitJSON_no_structure_leak(t *testing.T) {
	w, outBuf, _ := newTestWriter()

	if err := w.PrintGitJSON(makeGitSnap(sampleGitInfo())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := outBuf.String()
	if strings.Contains(raw, `"files"`) || strings.Contains(raw, `"directories"`) || strings.Contains(raw, `"relationships"`) {
		t.Fatalf("git JSON must not contain structure/graph data, got:\n%s", raw)
	}
}
