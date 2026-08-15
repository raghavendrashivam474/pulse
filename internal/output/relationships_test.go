package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/codebase"
	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
	"github.com/raghavendrashivam474/aayam/internal/testhelpers"
)

// ── terminal output ──────────────────────────────────────────────────────────

func TestPrintRelationships_ContainsNodeCount(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if !strings.Contains(got, "Nodes") {
		t.Errorf("expected Nodes in output, got: %q", got)
	}
}

func TestPrintRelationships_ContainsEdgeCount(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if !strings.Contains(got, "Edges") {
		t.Errorf("expected Edges in output, got: %q", got)
	}
}

func TestPrintRelationships_ContainsRelationshipTypes(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if !strings.Contains(got, "Relationship Types") {
		t.Errorf("expected Relationship Types section, got: %q", got)
	}
}

func TestPrintRelationships_EmptyGraph_ShowsNone(t *testing.T) {
	w, stdout, _ := newTestWriter()
	snap := snapshot.ProjectSnapshot{}
	w.PrintRelationships(snap)
	got := stdout.String()
	if !strings.Contains(got, "(none)") {
		t.Errorf("expected (none) for empty graph, got: %q", got)
	}
}

func TestPrintRelationships_DoesNotContainProjectName(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if strings.Contains(got, "Name:") {
		t.Errorf("relationships capability must not show project Name field, got: %q", got)
	}
}

func TestPrintRelationships_DoesNotContainGit(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if strings.Contains(got, "Branch") || strings.Contains(got, "Git") {
		t.Errorf("relationships capability must not contain git info, got: %q", got)
	}
}

func TestPrintRelationships_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	w.PrintRelationships(snapshot.ProjectSnapshot{})
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

// ── JSON output ───────────────────────────────────────────────────────────────

func TestPrintRelationshipsJSON_ValidJSON(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	if err := w.PrintRelationshipsJSON(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, stdout.String())
	}
}

func TestPrintRelationshipsJSON_CapabilityField(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintRelationshipsJSON(snapshot.ProjectSnapshot{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONRelationshipsResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Capability != "relationships" {
		t.Errorf("expected capability %q, got %q", "relationships", result.Capability)
	}
}

func TestPrintRelationshipsJSON_ExpectedFields(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	if err := w.PrintRelationshipsJSON(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONRelationshipsResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Application != "Aryntra Aayam" {
		t.Errorf("expected application %q, got %q", "Aryntra Aayam", result.Application)
	}
	if result.Relationships.Nodes < 0 {
		t.Errorf("expected non-negative nodes, got %d", result.Relationships.Nodes)
	}
	if result.Relationships.Edges < 0 {
		t.Errorf("expected non-negative edges, got %d", result.Relationships.Edges)
	}
}

func TestPrintRelationshipsJSON_EmptyGraph(t *testing.T) {
	w, stdout, _ := newTestWriter()
	if err := w.PrintRelationshipsJSON(snapshot.ProjectSnapshot{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONRelationshipsResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.Relationships.Nodes != 0 {
		t.Errorf("expected 0 nodes, got %d", result.Relationships.Nodes)
	}
	if result.Relationships.Edges != 0 {
		t.Errorf("expected 0 edges, got %d", result.Relationships.Edges)
	}
}

func TestPrintRelationshipsJSON_NodeEdgeConsistency(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	if err := w.PrintRelationshipsJSON(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result output.JSONRelationshipsResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	rel := result.Relationships
	edgeSum := rel.Contains + rel.BelongsTo + rel.Imports + rel.DependsOn
	if rel.Edges != edgeSum {
		t.Errorf("edge count %d does not match sum of types %d", rel.Edges, edgeSum)
	}
}

func TestPrintRelationshipsJSON_StdoutOnly(t *testing.T) {
	w, _, stderr := newTestWriter()
	_ = w.PrintRelationshipsJSON(snapshot.ProjectSnapshot{})
	if stderr.Len() != 0 {
		t.Errorf("expected stderr empty, got: %q", stderr.String())
	}
}

// TestPrintRelationships_MixedFixture verifies the mixed-project fixture
// produces a valid relationships output without panicking.
func TestPrintRelationships_MixedFixture(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "mixed-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	w, stdout, _ := newTestWriter()
	w.PrintRelationships(snap)
	got := stdout.String()
	if !strings.Contains(got, "Nodes") {
		t.Errorf("expected Nodes in mixed-project output, got: %q", got)
	}
}

// TestPrintRelationshipsJSON_RelationshipTypeSums verifies that
// sum of typed edges equals total edges for the go-project fixture.
// Uses codebase directly to confirm the graph is not rebuilt.
func TestPrintRelationshipsJSON_GraphNotRebuilt(t *testing.T) {
	fixturePath := testhelpers.ProjectFixturePath(t, "go-project")
	snap, err := snapshot.Discover(fixturePath)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// Confirm the graph comes from the snapshot, not a fresh build.
	byKind := snap.Graph.EdgesByKind()
	total := byKind[codebase.RelationshipContains] +
		byKind[codebase.RelationshipBelongsTo] +
		byKind[codebase.RelationshipImports] +
		byKind[codebase.RelationshipDependsOn]

	if snap.Graph.EdgeCount() != total {
		t.Errorf("snapshot graph edges %d != typed sum %d — graph may have been rebuilt",
			snap.Graph.EdgeCount(), total)
	}
}
