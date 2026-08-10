package cli

import (
	"bytes"
	"strings"
	"testing"
)

func newTestApp() (*App, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app := &App{
		stdout: stdout,
		stderr: stderr,
	}
	return app, stdout, stderr
}

func TestRun_NoArgs_PrintsDefault(t *testing.T) {
	app, stdout, _ := newTestApp()

	if err := app.Run([]string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Pulse") {
		t.Errorf("expected output to contain 'Pulse', got: %q", out)
	}
	if !strings.Contains(out, "v1.0.0") {
		t.Errorf("expected output to contain 'v1.0.0', got: %q", out)
	}
}

func TestRun_Version(t *testing.T) {
	app, stdout, _ := newTestApp()

	if err := app.Run([]string{"--version"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "v1.0.0") {
		t.Errorf("expected version output to contain 'v1.0.0', got: %q", out)
	}
}

func TestRun_Help(t *testing.T) {
	app, stdout, _ := newTestApp()

	if err := app.Run([]string{"--help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected help output to contain 'Usage', got: %q", out)
	}
	if !strings.Contains(out, "--version") {
		t.Errorf("expected help output to contain '--version', got: %q", out)
	}
	if !strings.Contains(out, "--json") {
		t.Errorf("expected help output to contain '--json', got: %q", out)
	}
}

func TestRun_InvalidFlag_ReturnsError(t *testing.T) {
	app, _, _ := newTestApp()

	err := app.Run([]string{"--notaflag"})
	if err == nil {
		t.Error("expected error for unknown flag, got nil")
	}
}

func TestRun_JSON(t *testing.T) {
	app, stdout, _ := newTestApp()

	if err := app.Run([]string{"--json"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output to contain '{', got: %q", out)
	}
	if !strings.Contains(out, "Pulse") {
		t.Errorf("expected JSON output to contain 'Pulse', got: %q", out)
	}
}

func TestRun_PathArgument(t *testing.T) {
	app, stdout, _ := newTestApp()

	if err := app.Run([]string{"./some-project"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Path is captured but not yet acted upon.
	// Verify the application still runs cleanly.
	out := stdout.String()
	if !strings.Contains(out, "Pulse") {
		t.Errorf("expected output to contain 'Pulse', got: %q", out)
	}
}
