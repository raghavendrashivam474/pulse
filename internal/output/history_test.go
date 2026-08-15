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

func makeHistorySnap(info gitinfo.GitInfo) snapshot.ProjectSnapshot {
	return snapshot.ProjectSnapshot{
		Git: info,
	}
}

func sampleHistoryInfo() gitinfo.GitInfo {
	return gitinfo.GitInfo{
		IsRepository: true,
		History: gitinfo.CommitHistory{
			Count: 5,
			Latest: gitinfo.CommitInfo{
				Hash:      "def456",
				Message:   "docs: add history capability",
				Author:    "Raghavendra Singh",
				Timestamp: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		Contributors: gitinfo.Contributors{
			Count: 1,
			Items: []gitinfo.Contributor{
				{Name: "Raghavendra Singh", CommitCount: 5},
			},
		},
	}
}

func TestPrintHistory_heading_and_sections(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintHistory(makeHistorySnap(sampleHistoryInfo()))

	out := outBuf.String()
	if !strings.Contains(out, "History") {
		t.Fatalf("expected History heading, got:\n%s", out)
	}
	if !strings.Contains(out, "Commits") || !strings.Contains(out, "Latest Commit") || !strings.Contains(out, "Contributors") {
		t.Fatalf("expected history sections, got:\n%s", out)
	}
}

func TestPrintHistory_not_repository(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintHistory(makeHistorySnap(gitinfo.GitInfo{IsRepository: false}))

	out := outBuf.String()
	if !strings.Contains(out, "Not a Git repository") {
		t.Fatalf("expected non-repository message, got:\n%s", out)
	}
}

func TestPrintHistory_no_git_state_leak(t *testing.T) {
	w, outBuf, _ := newTestWriter()
	w.PrintHistory(makeHistorySnap(sampleHistoryInfo()))

	out := outBuf.String()
	if strings.Contains(out, "Working Tree") || strings.Contains(out, "HEAD") || strings.Contains(out, "Branch:") {
		t.Fatalf("history output must not contain current git state sections, got:\n%s", out)
	}
}

func TestPrintHistoryJSON_envelope(t *testing.T) {
	w, outBuf, _ := newTestWriter()

	if err := w.PrintHistoryJSON(makeHistorySnap(sampleHistoryInfo())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.JSONHistoryResult
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Capability != "history" {
		t.Fatalf("expected capability=history, got %q", result.Capability)
	}
	if result.History.CommitCount != 5 {
		t.Fatalf("expected commit_count=5, got %d", result.History.CommitCount)
	}
	if result.History.Latest == nil {
		t.Fatal("expected latest commit")
	}
}

func TestPrintHistoryJSON_no_git_state_leak(t *testing.T) {
	w, outBuf, _ := newTestWriter()

	if err := w.PrintHistoryJSON(makeHistorySnap(sampleHistoryInfo())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := outBuf.String()
	if strings.Contains(raw, `"branch"`) || strings.Contains(raw, `"working_tree_state"`) || strings.Contains(raw, `"head"`) {
		t.Fatalf("history JSON must not contain git state fields, got:\n%s", raw)
	}
}
