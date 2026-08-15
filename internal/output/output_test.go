package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/project"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

func newTestWriter() (*output.Writer, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	w := &output.Writer{
		Out: stdout,
		Err: stderr,
	}
	return w, stdout, stderr
}

func TestPrintVersion(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintVersion()
	got := stdout.String()
	if !strings.Contains(got, "Aryntra Aayam v") {
		t.Errorf("expected version string, got: %q", got)
	}
	if !strings.Contains(got, "1.0.0") {
		t.Errorf("expected version 1.0.0, got: %q", got)
	}
}

func TestPrintHelp(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintHelp()
	got := stdout.String()
	if !strings.Contains(got, "Usage:") {
		t.Errorf("expected Usage: in help output, got: %q", got)
	}
	if !strings.Contains(got, "--help") {
		t.Errorf("expected --help flag in help output, got: %q", got)
	}
	if !strings.Contains(got, "--version") {
		t.Errorf("expected --version flag, got: %q", got)
	}
	if !strings.Contains(got, "--json") {
		t.Errorf("expected --json flag, got: %q", got)
	}
}

func TestPrintSummary(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintSummary()
	got := stdout.String()
	if !strings.Contains(got, "Aryntra Aayam v") {
		t.Errorf("expected Aryntra Aayam header in summary, got: %q", got)
	}
	if !strings.Contains(got, "No project analysis available yet.") {
		t.Errorf("expected placeholder message, got: %q", got)
	}
}

func TestPrintJSON(t *testing.T) {
	w, stdout, _ := newTestWriter()
	err := w.PrintJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "\"version\"") {
		t.Errorf("expected version field in JSON, got: %q", got)
	}
	if !strings.Contains(got, "\"status\"") {
		t.Errorf("expected status field in JSON, got: %q", got)
	}
}

func TestPrintError_WritesToStderr(t *testing.T) {
	w, stdout, stderr := newTestWriter()
	w.PrintError("target path does not exist: ./missing")
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected Error: prefix on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "target path does not exist") {
		t.Errorf("expected error message on stderr, got: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("expected stdout to be empty, got: %q", stdout.String())
	}
}

func TestPrintError_Format(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintError("something went wrong")
	got := stderr.String()
	expected := "Error: something went wrong\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// --- M2.8: Discovery output tests ---

func testSnapshot() snapshot.ProjectSnapshot {
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

func TestPrintDiscovery_ContainsName(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	if !strings.Contains(stdout.String(), "myapp") {
		t.Errorf("expected project name, got: %q", stdout.String())
	}
}

func TestPrintDiscovery_ContainsType(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	if !strings.Contains(stdout.String(), "Go") {
		t.Errorf("expected project type, got: %q", stdout.String())
	}
}

func TestPrintDiscovery_ContainsRoot(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	if !strings.Contains(stdout.String(), "/home/dev/myapp") {
		t.Errorf("expected root path, got: %q", stdout.String())
	}
}

func TestPrintDiscovery_ContainsLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	got := stdout.String()
	for _, lang := range []string{"Go", "Markdown", "PowerShell"} {
		if !strings.Contains(got, lang) {
			t.Errorf("expected language %q, got: %q", lang, got)
		}
	}
}

func TestPrintDiscovery_ContainsCounts(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	got := stdout.String()
	if !strings.Contains(got, "42") {
		t.Errorf("expected file count 42, got: %q", got)
	}
	if !strings.Contains(got, "14") {
		t.Errorf("expected directory count 14, got: %q", got)
	}
}

func TestPrintDiscovery_NoLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := testSnapshot()
	snap.Languages = nil
	w.PrintDiscovery(snap)
	if !strings.Contains(stdout.String(), "(none detected)") {
		t.Errorf("expected no-languages message, got: %q", stdout.String())
	}
}

func TestPrintDiscovery_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintDiscovery(testSnapshot())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

func TestPrintDiscoveryJSON_ValidJSON(t *testing.T) {
	w, stdout, _ := newTestWriter()
	err := w.PrintDiscoveryJSON(testSnapshot())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(stdout.Bytes(), &parsed); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, stdout.String())
	}
}

func TestPrintDiscoveryJSON_ExpectedFields(t *testing.T) {
	w, stdout, _ := newTestWriter()
	err := w.PrintDiscoveryJSON(testSnapshot())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.JSONDiscoveryResult
	if jsonErr := json.Unmarshal(stdout.Bytes(), &result); jsonErr != nil {
		t.Fatalf("unmarshal error: %v", jsonErr)
	}

	if result.Application != "Aryntra Aayam" {
		t.Errorf("expected application Aryntra Aayam, got %q", result.Application)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", result.Version)
	}
	if result.Project.Name != "myapp" {
		t.Errorf("expected name myapp, got %q", result.Project.Name)
	}
	if result.Project.Root != "/home/dev/myapp" {
		t.Errorf("expected root /home/dev/myapp, got %q", result.Project.Root)
	}
	if result.Project.Type != "Go" {
		t.Errorf("expected type Go, got %q", result.Project.Type)
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

func TestPrintDiscoveryJSON_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	_ = w.PrintDiscoveryJSON(testSnapshot())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

// --- M6.2: Overview output tests ---

func TestPrintOverview_ContainsName(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	if !strings.Contains(stdout.String(), "myapp") {
		t.Errorf("expected project name in overview, got: %q", stdout.String())
	}
}

func TestPrintOverview_ContainsType(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	if !strings.Contains(stdout.String(), "Go") {
		t.Errorf("expected project type in overview, got: %q", stdout.String())
	}
}

func TestPrintOverview_ContainsCounts(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	if !strings.Contains(got, "42") {
		t.Errorf("expected file count 42 in overview, got: %q", got)
	}
	if !strings.Contains(got, "14") {
		t.Errorf("expected directory count 14 in overview, got: %q", got)
	}
}

func TestPrintOverview_ContainsLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	for _, lang := range []string{"Go", "Markdown", "PowerShell"} {
		if !strings.Contains(got, lang) {
			t.Errorf("expected language %q in overview, got: %q", lang, got)
		}
	}
}

func TestPrintOverview_NoLanguages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := testSnapshot()
	snap.Languages = nil
	w.PrintOverview(snap)
	if !strings.Contains(stdout.String(), "(none detected)") {
		t.Errorf("expected no-languages message in overview, got: %q", stdout.String())
	}
}

func TestPrintOverview_DoesNotContainPackages(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	if strings.Contains(got, "Packages") {
		t.Errorf("overview should not contain Packages section, got: %q", got)
	}
}

func TestPrintOverview_DoesNotContainDependencies(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	if strings.Contains(got, "Dependencies") {
		t.Errorf("overview should not contain Dependencies section, got: %q", got)
	}
}

func TestPrintOverview_DoesNotContainHistory(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	if strings.Contains(got, "History") {
		t.Errorf("overview should not contain History section, got: %q", got)
	}
}

func TestPrintOverview_DoesNotContainContributors(t *testing.T) {
	w, stdout, _ := newTestWriter()
	w.PrintOverview(testSnapshot())
	got := stdout.String()
	if strings.Contains(got, "Contributors") {
		t.Errorf("overview should not contain Contributors section, got: %q", got)
	}
}

func TestPrintOverview_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintOverview(testSnapshot())
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

func TestPrintOverviewJSON_ValidJSON(t *testing.T) {
	w, stdout, _ := newTestWriter()
	err := w.PrintOverviewJSON(testSnapshot())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(stdout.Bytes(), &parsed); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, stdout.String())
	}
}

func TestPrintOverviewJSON_ExpectedFields(t *testing.T) {
	w, stdout, _ := newTestWriter()
	err := w.PrintOverviewJSON(testSnapshot())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.JSONOverviewResult
	if jsonErr := json.Unmarshal(stdout.Bytes(), &result); jsonErr != nil {
		t.Fatalf("unmarshal error: %v", jsonErr)
	}

	if result.Capability != "overview" {
		t.Errorf("expected capability 'overview', got %q", result.Capability)
	}
	if result.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", result.Name)
	}
	if result.Type != "Go" {
		t.Errorf("expected type 'Go', got %q", result.Type)
	}
	if result.FileCount != 42 {
		t.Errorf("expected file_count 42, got %d", result.FileCount)
	}
	if result.DirectoryCount != 14 {
		t.Errorf("expected directory_count 14, got %d", result.DirectoryCount)
	}
	if len(result.Languages) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(result.Languages))
	}
}
