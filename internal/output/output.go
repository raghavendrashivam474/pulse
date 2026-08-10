package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Renderer writes Pulse output to the configured writer.
type Renderer struct {
	w    io.Writer
	json bool
}

// NewRenderer creates a Renderer.
func NewRenderer(w io.Writer, asJSON bool) *Renderer {
	return &Renderer{w: w, json: asJSON}
}

// RenderDefault prints the default Pulse output when no analysis is available.
func (r *Renderer) RenderDefault() {
	if r.json {
		r.renderDefaultJSON()
		return
	}
	r.renderDefaultText()
}

func (r *Renderer) renderDefaultText() {
	fmt.Fprintln(r.w, "Pulse v1.0.0")
	fmt.Fprintln(r.w)
	fmt.Fprintln(r.w, "Project intelligence for developers.")
	fmt.Fprintln(r.w)
	fmt.Fprintln(r.w, "No project analysis available yet.")
}

func (r *Renderer) renderDefaultJSON() {
	payload := map[string]string{
		"application": "Pulse",
		"version":     "v1.0.0",
		"status":      "no analysis available",
	}
	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}
