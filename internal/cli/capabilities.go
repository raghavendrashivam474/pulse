package cli

import (
	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/output"
)

// ── overview ─────────────────────────────────────────────────────────────────

type overviewCapability struct {
	writer *output.Writer
}

func (c overviewCapability) Name() string        { return "overview" }
func (c overviewCapability) Description() string { return "Show a concise project overview" }
func (c overviewCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintOverviewJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintOverview(ctx.Snap)
	}
	return capability.Result{CapabilityName: "overview"}, nil
}

// ── project (M6.3) ───────────────────────────────────────────────────────────

type projectCapability struct {
	writer *output.Writer
}

func (c projectCapability) Name() string        { return "project" }
func (c projectCapability) Description() string { return "Show project identity and metadata" }
func (c projectCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintProjectJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintProject(ctx.Snap)
	}
	return capability.Result{CapabilityName: "project"}, nil
}

// ── structure (M6.4) ─────────────────────────────────────────────────────────

type structureCapability struct {
	writer *output.Writer
}

func (c structureCapability) Name() string        { return "structure" }
func (c structureCapability) Description() string { return "Show filesystem and language structure" }
func (c structureCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintStructureJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintStructure(ctx.Snap)
	}
	return capability.Result{CapabilityName: "structure"}, nil
}

// ── relationships (M6.5) ─────────────────────────────────────────────────────

type relationshipsCapability struct {
	writer *output.Writer
}

func (c relationshipsCapability) Name() string        { return "relationships" }
func (c relationshipsCapability) Description() string { return "Show relationship graph intelligence" }
func (c relationshipsCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintRelationshipsJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintRelationships(ctx.Snap)
	}
	return capability.Result{CapabilityName: "relationships"}, nil
}

// ── registry ─────────────────────────────────────────────────────────────────

func newCapabilityRegistry(w *output.Writer) *capability.Registry {
	reg := capability.NewRegistry()

	mustRegister := func(c capability.Capability) {
		if err := reg.Register(c); err != nil {
			panic(err)
		}
	}

	mustRegister(overviewCapability{writer: w})
	mustRegister(projectCapability{writer: w})
	mustRegister(structureCapability{writer: w})
	mustRegister(relationshipsCapability{writer: w})

	return reg
}

func capabilityHelpEntries(reg *capability.Registry) []output.CapabilityInfo {
	names := reg.Names()
	entries := make([]output.CapabilityInfo, 0, len(names))
	for _, name := range names {
		c, ok := reg.Lookup(name)
		if !ok {
			continue
		}
		entries = append(entries, output.CapabilityInfo{
			Name:        c.Name(),
			Description: c.Description(),
		})
	}
	return entries
}
