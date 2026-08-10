package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderDefault_Text(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, false)
	r.RenderDefault()

	out := buf.String()
	if !strings.Contains(out, "Pulse") {
		t.Errorf("expected 'Pulse' in output, got: %q", out)
	}
	if !strings.Contains(out, "v1.0.0") {
		t.Errorf("expected 'v1.0.0' in output, got: %q", out)
	}
	if !strings.Contains(out, "No project analysis available yet") {
		t.Errorf("expected status message in output, got: %q", out)
	}
}

func TestRenderDefault_JSON(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf, true)
	r.RenderDefault()

	out := buf.String()
	if !strings.Contains(out, `"application"`) {
		t.Errorf("expected 'application' key in JSON, got: %q", out)
	}
	if !strings.Contains(out, `"version"`) {
		t.Errorf("expected 'version' key in JSON, got: %q", out)
	}
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected 'status' key in JSON, got: %q", out)
	}
	if !strings.Contains(out, "Pulse") {
		t.Errorf("expected 'Pulse' in JSON output, got: %q", out)
	}
}
