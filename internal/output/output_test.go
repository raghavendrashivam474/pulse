package output_test

import (
	"bytes"
	"strings"
	"testing"

	"pulse/internal/output"
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
	if !strings.Contains(got, "Pulse v") {
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
	if !strings.Contains(got, "Pulse v") {
		t.Errorf("expected Pulse header in summary, got: %q", got)
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
	if !strings.Contains(got, `"version"`) {
		t.Errorf("expected version field in JSON, got: %q", got)
	}
	if !strings.Contains(got, `"status"`) {
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
