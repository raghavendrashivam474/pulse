package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// projectSnap returns a minimal deterministic snapshot for project tests.
func projectSnap() snapshot.ProjectSnapshot {
	return snapshot.ProjectSnapshot{
		Name: "myapp",
		Root: "/home/dev/myapp",
		Type: project.TypeGo,
		Languages: []project.Language{
			project.LangGo,
			project.LangMarkdown,
			project.LangPowerShell,
		},
		FileCount:      42,
		DirectoryCount: 14,
	}
}

// ── terminal output ──────────────────────────────────────────────────────────

func TestPrintProject_ContainsName(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	if !strings.Contains(stdout.String(), "myapp") {
		t.Errorf("expected project name, got: %q", stdout.String())
	}
}

func TestPrintProject_ContainsType(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	if !strings.Contains(stdout.String(), "Go") {
		t.Errorf("expected project type, got: %q", stdout.String())
	}
}

func TestPrintProject_ContainsRoot(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	if !strings.Contains(stdout.String(), "/home/dev/myapp") {
		t.Errorf("expected root path, got: %q", stdout.String())
	}
}

func TestPrintProject_ContainsFileCounts(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	got := stdout.String()
	if !strings.Contains(got, "42") {
		t.Errorf("expected file count 42, got: %q", got)
	}
	if !strings.Contains(got, "14") {
		t.Errorf("expected directory count 14, got: %q", got)
	}
}

func TestPrintProject_ContainsLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	got := stdout.String()
	for _, lang := range []string{"Go", "Markdown", "PowerShell"} {
		if !strings.Contains(got, lang) {
			t.Errorf("expected language %q in output, got: %q", lang, got)
		}
	}
}

func TestPrintProject_NoLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := projectSnap()
	snap.Languages = nil
	w.PrintProject(snap)
	if !strings.Contains(stdout.String(), "(none detected)") {
		t.Errorf("expected no-languages message, got: %q", stdout.String())
	}
}

func TestPrintProject_DoesNotContainGit(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	got := stdout.String()
	if strings.Contains(got, "Git") || strings.Contains(got, "Branch") {
		t.Errorf("project capability must not contain git info, got: %q", got)
	}
}

func TestPrintProject_DoesNotContainRelationships(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintProject(projectSnap())
	got := stdout.String()
	if strings.Contains(got, "Relationships") || strings.Contains(got, "Edges") {
		t.Errorf("project capability must not contain relationship info, got: %q", got)
	}
}

func TestPrintProject_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintProject(projectSnap())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

// ── JSON output ───────────────────────────────────────────────────────────────

func TestPrintProjectJSON_ValidJSON(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintProjectJSON(projectSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, stdout.String())
	}
}

func TestPrintProjectJSON_CapabilityField(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintProjectJSON(projectSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONProjectResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Capability != "project" {
		t.Errorf("expected capability %q, got %q", "project", result.Capability)
	}
}

func TestPrintProjectJSON_ExpectedFields(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintProjectJSON(projectSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONProjectResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Application != "Aryntra Aayam" {
		t.Errorf("expected application %q, got %q", "Aryntra Aayam", result.Application)
	}
	if result.Project.Name != "myapp" {
		t.Errorf("expected name %q, got %q", "myapp", result.Project.Name)
	}
	if result.Project.Root != "/home/dev/myapp" {
		t.Errorf("expected root %q, got %q", "/home/dev/myapp", result.Project.Root)
	}
	if result.Project.Type != "Go" {
		t.Errorf("expected type %q, got %q", "Go", result.Project.Type)
	}
	if result.Project.FileCount != 42 {
		t.Errorf("expected file_count 42, got %d", result.Project.FileCount)
	}
	if result.Project.DirectoryCount != 14 {
		t.Errorf("expected directory_count 14, got %d", result.Project.DirectoryCount)
	}
	if len(result.Project.Languages) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(result.Project.Languages))
	}
}

func TestPrintProjectJSON_EmptyLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := projectSnap()
	snap.Languages = nil
	if err := w.PrintProjectJSON(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONProjectResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(result.Project.Languages) != 0 {
		t.Errorf("expected 0 languages, got %d", len(result.Project.Languages))
	}
}

func TestPrintProjectJSON_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	_ = w.PrintProjectJSON(projectSnap())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}
