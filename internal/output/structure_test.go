package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// structureSnap returns a deterministic snapshot for structure tests.
func structureSnap() snapshot.ProjectSnapshot {
	return snapshot.ProjectSnapshot{
		Name: "myapp",
		Root: "/home/dev/myapp",
		Type: project.TypeNode,
		Languages: []project.Language{
			project.LangCSS,
			project.LangHTML,
			project.LangTypeScript,
		},
		FileCount:      5,
		DirectoryCount: 1,
	}
}

// ── terminal output ──────────────────────────────────────────────────────────

func TestPrintStructure_ContainsFileCount(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	if !strings.Contains(stdout.String(), "5") {
		t.Errorf("expected file count 5, got: %q", stdout.String())
	}
}

func TestPrintStructure_ContainsDirectoryCount(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	if !strings.Contains(stdout.String(), "1") {
		t.Errorf("expected directory count 1, got: %q", stdout.String())
	}
}

func TestPrintStructure_ContainsLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	got := stdout.String()
	for _, lang := range []string{"CSS", "HTML", "TypeScript"} {
		if !strings.Contains(got, lang) {
			t.Errorf("expected language %q in output, got: %q", lang, got)
		}
	}
}

func TestPrintStructure_NoLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := structureSnap()
	snap.Languages = nil
	w.PrintStructure(snap)
	if !strings.Contains(stdout.String(), "(none detected)") {
		t.Errorf("expected no-languages message, got: %q", stdout.String())
	}
}

func TestPrintStructure_DoesNotContainProjectName(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	// structure capability must not expose project identity fields
	if strings.Contains(stdout.String(), "Name:") {
		t.Errorf("structure capability must not contain Name field, got: %q", stdout.String())
	}
}

func TestPrintStructure_DoesNotContainGit(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	got := stdout.String()
	if strings.Contains(got, "Git") || strings.Contains(got, "Branch") {
		t.Errorf("structure capability must not contain git info, got: %q", got)
	}
}

func TestPrintStructure_DoesNotContainRelationships(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintStructure(structureSnap())
	got := stdout.String()
	if strings.Contains(got, "Relationships") || strings.Contains(got, "Edges") {
		t.Errorf("structure capability must not contain relationship info, got: %q", got)
	}
}

func TestPrintStructure_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintStructure(structureSnap())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

func TestPrintStructure_MixedLanguageProject(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := snapshot.ProjectSnapshot{
		FileCount:      4,
		DirectoryCount: 1,
		Languages: []project.Language{
			project.LangGo,
			project.LangJavaScript,
			project.LangMarkdown,
			project.LangPython,
		},
	}
	w.PrintStructure(snap)
	got := stdout.String()
	for _, lang := range []string{"Go", "JavaScript", "Markdown", "Python"} {
		if !strings.Contains(got, lang) {
			t.Errorf("expected language %q, got: %q", lang, got)
		}
	}
}

// ── JSON output ───────────────────────────────────────────────────────────────

func TestPrintStructureJSON_ValidJSON(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintStructureJSON(structureSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, stdout.String())
	}
}

func TestPrintStructureJSON_CapabilityField(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintStructureJSON(structureSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONStructureResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Capability != "structure" {
		t.Errorf("expected capability %q, got %q", "structure", result.Capability)
	}
}

func TestPrintStructureJSON_ExpectedFields(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintStructureJSON(structureSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONStructureResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Structure.FileCount != 5 {
		t.Errorf("expected file_count 5, got %d", result.Structure.FileCount)
	}
	if result.Structure.DirectoryCount != 1 {
		t.Errorf("expected directory_count 1, got %d", result.Structure.DirectoryCount)
	}
	if len(result.Structure.Languages) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(result.Structure.Languages))
	}
}

func TestPrintStructureJSON_DeterministicLanguageOrder(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintStructureJSON(structureSnap()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONStructureResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	expected := []string{"CSS", "HTML", "TypeScript"}
	for i, lang := range expected {
		if result.Structure.Languages[i] != lang {
			t.Errorf("languages[%d]: expected %q, got %q", i, lang, result.Structure.Languages[i])
		}
	}
}

func TestPrintStructureJSON_EmptyProject(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := snapshot.ProjectSnapshot{}
	if err := w.PrintStructureJSON(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONStructureResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Structure.FileCount != 0 {
		t.Errorf("expected 0 files, got %d", result.Structure.FileCount)
	}
	if len(result.Structure.Languages) != 0 {
		t.Errorf("expected 0 languages, got %d", len(result.Structure.Languages))
	}
}

func TestPrintStructureJSON_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	_ = w.PrintStructureJSON(structureSnap())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}
